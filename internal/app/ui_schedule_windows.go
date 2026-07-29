//go:build windows

package app

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Libellés courts des jours, index = time.Weekday (0=dimanche … 6=samedi).
var weekdayLabels = [7]string{"D", "L", "M", "M", "J", "V", "S"}

// Ordre d'affichage : on commence par lundi (plus naturel qu'à dimanche).
var weekdayOrder = [7]int{1, 2, 3, 4, 5, 6, 0}

// buildScheduleTab : éditeur des plages horaires (Temps 2).
func (u *ui) buildScheduleTab() fyne.CanvasObject {
	intro := widget.NewLabelWithStyle(
		"Quand le mode planifié est actif, le maintien tourne automatiquement "+
			"pendant les plages déclarées, et reste coupé en dehors.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)
	intro.Wrapping = fyne.TextWrapWord

	// Interrupteur du mode planifié.
	modeCheck := widget.NewCheck("Activer le mode planifié", func(on bool) {
		u.scheduler.SetEnabled(on)
		u.refresh(u.engine.IsEnabled()) // met à jour l'onglet État (toggle grisé)
	})
	modeCheck.SetChecked(u.scheduler.IsEnabled())

	// Conteneur de la liste des plages, reconstruit à chaque modification.
	u.scheduleList = container.NewVBox()
	u.rebuildScheduleList()

	addBtn := widget.NewButtonWithIcon("Ajouter une plage", theme.ContentAddIcon(), func() {
		u.cfg.Schedules = append(u.cfg.Schedules, Schedule{
			Start: "09:00", End: "18:00", Days: []int{1, 2, 3, 4, 5},
		})
		_ = u.cfg.save()
		u.rebuildScheduleList()
		u.scheduler.SchedulesChanged()
	})

	header := container.NewVBox(
		intro,
		container.NewPadded(modeCheck),
		widget.NewSeparator(),
	)

	// La liste défile si beaucoup de plages ; le bouton reste en bas.
	scroll := container.NewVScroll(container.NewPadded(u.scheduleList))

	return container.NewBorder(
		header,
		container.NewPadded(addBtn),
		nil, nil,
		scroll,
	)
}

// rebuildScheduleList reconstruit les lignes d'édition à partir de la config.
func (u *ui) rebuildScheduleList() {
	u.scheduleList.Objects = nil
	if len(u.cfg.Schedules) == 0 {
		empty := widget.NewLabel("Aucune plage. Ajoutez-en une ci-dessous.")
		empty.Importance = widget.LowImportance
		u.scheduleList.Add(empty)
	}
	for i := range u.cfg.Schedules {
		u.scheduleList.Add(u.buildScheduleRow(i))
	}
	u.scheduleList.Refresh()
}

// buildScheduleRow construit la carte d'édition d'une plage (index i), au
// design sobre : en-tête (pastille d'état + heures + suppression), jours, puis
// un switch d'activation cohérent avec le reste de l'app.
func (u *ui) buildScheduleRow(i int) fyne.CanvasObject {
	sc := u.cfg.Schedules[i]

	startEntry := widget.NewEntry()
	startEntry.SetText(sc.Start)
	startEntry.SetPlaceHolder("HH:MM")
	startEntry.Validator = hmValidator

	endEntry := widget.NewEntry()
	endEntry.SetText(sc.End)
	endEntry.SetPlaceHolder("HH:MM")
	endEntry.Validator = hmValidator

	startBox := container.NewGridWrap(fyne.NewSize(72, startEntry.MinSize().Height), startEntry)
	endBox := container.NewGridWrap(fyne.NewSize(72, endEntry.MinSize().Height), endEntry)

	commit := func() {
		u.cfg.Schedules[i].Start = startEntry.Text
		u.cfg.Schedules[i].End = endEntry.Text
		_ = u.cfg.save()
		u.scheduler.SchedulesChanged()
	}
	startEntry.OnChanged = func(string) { commit() }
	endEntry.OnChanged = func(string) { commit() }

	// Pastille d'état (point coloré) reflétant activée/désactivée.
	statusDot := canvas.NewCircle(scheduleDotColor(sc.Enabled()))
	dotBox := container.NewGridWrap(fyne.NewSize(10, 10), statusDot)

	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		u.cfg.Schedules = append(u.cfg.Schedules[:i], u.cfg.Schedules[i+1:]...)
		_ = u.cfg.save()
		u.rebuildScheduleList()
		u.scheduler.SchedulesChanged()
	})
	delBtn.Importance = widget.LowImportance

	// En-tête : pastille + "HH:MM → HH:MM" + bouton supprimer aligné à droite.
	header := container.NewBorder(nil, nil,
		container.NewHBox(container.NewCenter(dotBox), startBox, widget.NewLabel("→"), endBox),
		delBtn,
	)

	// Jours, cases à cocher (conservées ici comme demandé).
	dayChecks := container.NewHBox()
	for _, wd := range weekdayOrder {
		wd := wd
		check := widget.NewCheck(weekdayLabels[wd], nil)
		check.SetChecked(dayInList(u.cfg.Schedules[i].Days, wd))
		check.OnChanged = func(on bool) {
			u.cfg.Schedules[i].Days = toggleDay(u.cfg.Schedules[i].Days, wd, on)
			_ = u.cfg.save()
			u.scheduler.SchedulesChanged()
		}
		dayChecks.Add(check)
	}

	// Switch d'activation + libellé d'état, cohérent avec les autres onglets.
	statusLabel := widget.NewLabel(scheduleToggleText(sc.Enabled()))
	statusLabel.Importance = widget.LowImportance
	sw := newToggleSwitch(sc.Enabled(), func(on bool) {
		u.cfg.Schedules[i].Disabled = !on
		_ = u.cfg.save()
		statusDot.FillColor = scheduleDotColor(on)
		statusDot.Refresh()
		statusLabel.SetText(scheduleToggleText(on))
		u.scheduler.SchedulesChanged()
	})
	activationRow := container.NewBorder(nil, nil, statusLabel, container.NewCenter(sw))

	return widget.NewCard("", "", container.NewVBox(
		header,
		dayChecks,
		widget.NewSeparator(),
		activationRow,
	))
}

// scheduleDotColor : vert si la plage est active, gris sinon.
func scheduleDotColor(enabled bool) color.Color {
	if enabled {
		return colorActive
	}
	return colorPaused
}

func scheduleToggleText(enabled bool) string {
	if enabled {
		return "Activée"
	}
	return "Désactivée"
}

// hmValidator valide un champ heure "HH:MM" pour le retour visuel de Fyne.
func hmValidator(s string) error {
	_, err := parseHM(s)
	return err
}

// dayInList indique si le jour wd est présent.
func dayInList(days []int, wd int) bool {
	for _, d := range days {
		if d == wd {
			return true
		}
	}
	return false
}

// toggleDay ajoute ou retire le jour wd de la liste.
func toggleDay(days []int, wd int, add bool) []int {
	if add {
		if !dayInList(days, wd) {
			days = append(days, wd)
		}
		return days
	}
	out := days[:0:0]
	for _, d := range days {
		if d != wd {
			out = append(out, d)
		}
	}
	return out
}
