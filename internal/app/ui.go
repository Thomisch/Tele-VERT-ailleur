//go:build windows

package app

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Couleurs de l'indicateur d'état.
var (
	colorActive = color.NRGBA{R: 0x2E, G: 0xCC, B: 0x71, A: 0xFF} // vert
	colorPaused = color.NRGBA{R: 0x95, G: 0xA5, B: 0xA6, A: 0xFF} // gris
)

// ui regroupe les widgets dont l'état doit rester synchronisé.
type ui struct {
	cfg       *Config
	engine    *Engine
	scheduler *Scheduler

	win        fyne.Window
	dot        *canvas.Circle
	banner     *canvas.Rectangle
	statusText *canvas.Text
	toggle     *toggleSwitch

	// Widgets du planning, rafraîchis quand le mode planifié change.
	scheduleList *fyne.Container
	modeLabel    *widget.Label
}

// runUI construit et lance l'interface (fenêtre + tray). Bloque jusqu'à quit.
func runUI(cfg *Config, engine *Engine, scheduler *Scheduler) {
	a := app.NewWithID("io.github.thomisch.televertailleur")
	a.Settings().SetTheme(theme.DarkTheme())

	u := &ui{cfg: cfg, engine: engine, scheduler: scheduler}
	u.win = a.NewWindow("FuckTeamsStatus")
	u.win.SetContent(u.buildContent())
	u.win.Resize(fyne.NewSize(440, 520))
	u.win.CenterOnScreen()

	// Le moteur peut changer d'état depuis le tray : on garde la GUI à jour.
	engine.OnChange(func(enabled bool) {
		fyne.Do(func() { u.refresh(enabled) })
	})

	u.setupTray(a)

	// Fermer la fenêtre la cache (l'appli continue dans le tray) plutôt que
	// de tout quitter.
	u.win.SetCloseIntercept(func() { u.win.Hide() })

	// Instance unique : quand un autre lancement signale l'event, on ramène la
	// fenêtre au premier plan.
	go watchShowRequests(func() {
		fyne.Do(func() {
			u.win.Show()
			u.win.RequestFocus()
		})
	})

	u.refresh(engine.IsEnabled())
	u.win.ShowAndRun()
}

// buildContent assemble l'interface à onglets : « État » et « Planning ».
func (u *ui) buildContent() fyne.CanvasObject {
	banner := u.buildBanner()

	tabs := container.NewAppTabs(
		container.NewTabItem("État", u.buildStateTab()),
		container.NewTabItem("Planning", u.buildScheduleTab()),
		container.NewTabItem("Avancé", u.buildAdvancedTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	// Réaligne l'onglet État (toggle grisé/étiquette) à chaque sélection, au
	// cas où le mode planifié a changé entre-temps.
	tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "État" {
			u.refresh(u.engine.IsEnabled())
		}
	}

	return container.NewBorder(
		container.NewPadded(banner), // haut : bandeau d'état toujours visible
		nil, nil, nil,
		tabs,
	)
}

// buildBanner construit le bandeau d'état coloré (pastille + libellé).
func (u *ui) buildBanner() fyne.CanvasObject {
	u.banner = canvas.NewRectangle(colorPaused)
	u.banner.CornerRadius = 8
	u.banner.SetMinSize(fyne.NewSize(0, 56))

	u.dot = canvas.NewCircle(color.White)
	dotBox := container.NewGridWrap(fyne.NewSize(14, 14), u.dot)

	u.statusText = canvas.NewText("", color.White)
	u.statusText.TextStyle = fyne.TextStyle{Bold: true}
	u.statusText.TextSize = 20

	bannerContent := container.NewPadded(
		container.NewHBox(
			layout.NewSpacer(),
			container.NewCenter(dotBox),
			container.NewCenter(u.statusText),
			layout.NewSpacer(),
		),
	)
	return container.NewStack(u.banner, bannerContent)
}

// buildStateTab : Temps 1 (toggle keep-alive) + options démarrage/anti-veille.
func (u *ui) buildStateTab() fyne.CanvasObject {
	subtitle := widget.NewLabelWithStyle(
		"Maintien ta présence Teams en vert\ncomme un bon télétravailleur !",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Toggle principal keep-alive (switch élégant).
	toggleRowObj, sw := toggleRow("Maintien actif", u.engine.IsEnabled(), func(on bool) {
		u.engine.SetEnabled(on)
	})
	u.toggle = sw

	// État initial : si le planning pilote, le toggle est grisé dès la création
	// (sans dépendre d'un refresh ultérieur, fragile au tout premier rendu).
	scheduled := u.scheduler != nil && u.scheduler.IsEnabled()
	u.toggle.disabled = scheduled

	// Étiquette indiquant si le planning pilote l'état (toggle alors désactivé).
	u.modeLabel = widget.NewLabel("")
	u.modeLabel.Wrapping = fyne.TextWrapWord
	u.modeLabel.Importance = widget.LowImportance
	if scheduled {
		u.modeLabel.SetText("Piloté par le planning — voir l'onglet Planning.")
	}

	// Option démarrage auto.
	autostartRow, _ := toggleRow("Lancer au démarrage de Windows", u.cfg.LaunchAtStartup, func(on bool) {
		u.cfg.LaunchAtStartup = on
		if err := setAutoStart(on); err == nil {
			_ = u.cfg.save()
		}
	})

	// Option anti-veille : empêche la mise en veille système tant que le
	// maintien tourne. Suit l'état effectif via applySleepPolicy.
	preventSleepRow, _ := toggleRow("Empêcher la mise en veille de l'ordinateur", u.cfg.PreventSleep, func(on bool) {
		u.cfg.PreventSleep = on
		_ = u.cfg.save()
		applySleepPolicy(u.cfg, u.engine)
	})

	info := widget.NewLabelWithStyle(
		"Réveil système toutes les 60 s (F15 + micro-mouvement souris).",
		fyne.TextAlignLeading,
		fyne.TextStyle{},
	)
	info.Wrapping = fyne.TextWrapWord
	info.Importance = widget.LowImportance

	return container.NewPadded(container.NewVBox(
		subtitle,
		widget.NewSeparator(),
		toggleRowObj,
		u.modeLabel,
		autostartRow,
		preventSleepRow,
		widget.NewSeparator(),
		info,
	))
}

// refresh aligne tous les widgets visuels sur l'état réel du moteur.
func (u *ui) refresh(enabled bool) {
	if u.toggle != nil {
		u.toggle.SetOn(enabled)
	}

	// Quand le planning pilote, le toggle manuel devient indicatif (grisé) et
	// une étiquette l'explique.
	scheduled := u.scheduler != nil && u.scheduler.IsEnabled()
	if u.toggle != nil {
		u.toggle.SetDisabled(scheduled)
	}
	if u.modeLabel != nil {
		if scheduled {
			u.modeLabel.SetText("Piloté par le planning — voir l'onglet Planning.")
		} else {
			u.modeLabel.SetText("")
		}
	}

	if enabled {
		u.banner.FillColor = colorActive
		u.statusText.Text = "Actif"
	} else {
		u.banner.FillColor = colorPaused
		u.statusText.Text = "En pause"
	}
	// Pastille et texte restent blancs : ils contrastent sur le bandeau coloré.
	u.dot.FillColor = color.White
	u.statusText.Color = color.White

	u.banner.Refresh()
	u.dot.Refresh()
	u.statusText.Refresh()
}

// setupTray installe l'icône et le menu de la barre des tâches via le support
// natif de Fyne (pas de seconde lib tray).
func (u *ui) setupTray(a fyne.App) {
	deskApp, ok := a.(desktop.App)
	if !ok {
		return // environnement sans tray : on continue avec la fenêtre seule
	}

	showItem := fyne.NewMenuItem("Ouvrir", func() {
		u.win.Show()
		u.win.RequestFocus()
	})
	toggleItem := fyne.NewMenuItem("Activer / Désactiver", func() {
		// Ignoré quand le planning pilote (l'état est dicté par l'horaire).
		if u.scheduler == nil || !u.scheduler.IsEnabled() {
			u.engine.Toggle()
		}
	})

	menu := fyne.NewMenu("FuckTeamsStatus", showItem, toggleItem)
	deskApp.SetSystemTrayMenu(menu)
	deskApp.SetSystemTrayIcon(trayIconResource(u.engine.IsEnabled()))

	// Met à jour l'icône tray selon l'état.
	u.engine.OnChange(func(enabled bool) {
		fyne.Do(func() { deskApp.SetSystemTrayIcon(trayIconResource(enabled)) })
	})
}

// trayIconResource renvoie l'icône Fyne correspondant à l'état.
func trayIconResource(active bool) fyne.Resource {
	if active {
		return fyne.NewStaticResource("active.png", iconActivePNG)
	}
	return fyne.NewStaticResource("paused.png", iconPausedPNG)
}
