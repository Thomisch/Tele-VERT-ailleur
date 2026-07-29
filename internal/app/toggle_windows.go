//go:build windows

package app

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Dimensions du switch.
const (
	toggleWidth   = 40
	toggleHeight  = 22
	togglePadding = 3
)

var (
	toggleOffColor = color.NRGBA{R: 0x55, G: 0x5A, B: 0x5E, A: 0xFF} // rail gris éteint
	toggleKnob     = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // pastille blanche
)

// toggleSwitch est un interrupteur on/off élégant (rail arrondi + pastille qui
// glisse), vert = on, gris = off. Remplace les cases à cocher dans les onglets
// État et Avancé pour un rendu plus sobre.
type toggleSwitch struct {
	widget.BaseWidget
	on       bool
	disabled bool
	onChange func(bool)

	rail *canvas.Rectangle
	knob *canvas.Circle
}

func newToggleSwitch(initial bool, onChange func(bool)) *toggleSwitch {
	t := &toggleSwitch{on: initial, onChange: onChange}
	t.ExtendBaseWidget(t)
	return t
}

func (t *toggleSwitch) CreateRenderer() fyne.WidgetRenderer {
	t.rail = canvas.NewRectangle(t.railColor())
	t.rail.CornerRadius = toggleHeight / 2
	t.knob = canvas.NewCircle(toggleKnob)
	r := &toggleRenderer{t: t, objects: []fyne.CanvasObject{t.rail, t.knob}}
	r.Layout(r.MinSize())
	return r
}

func (t *toggleSwitch) railColor() color.Color {
	if t.disabled {
		// Grisé : couleur atténuée quel que soit l'état.
		return color.NRGBA{R: 0x3A, G: 0x3D, B: 0x40, A: 0xFF}
	}
	if t.on {
		return colorActive
	}
	return toggleOffColor
}

// SetOn change l'état sans déclencher le callback (synchronisation externe).
func (t *toggleSwitch) SetOn(v bool) {
	if t.on == v {
		return
	}
	t.on = v
	t.Refresh()
}

// SetDisabled grise le switch et ignore les clics (état piloté ailleurs).
func (t *toggleSwitch) SetDisabled(v bool) {
	if t.disabled == v {
		return
	}
	t.disabled = v
	t.Refresh()
}

func (t *toggleSwitch) Tapped(_ *fyne.PointEvent) {
	if t.disabled {
		return
	}
	t.on = !t.on
	t.Refresh()
	if t.onChange != nil {
		t.onChange(t.on)
	}
}

func (t *toggleSwitch) Cursor() desktop.Cursor {
	if t.disabled {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

type toggleRenderer struct {
	t       *toggleSwitch
	objects []fyne.CanvasObject
}

func (r *toggleRenderer) Layout(_ fyne.Size) {
	r.t.rail.Resize(fyne.NewSize(toggleWidth, toggleHeight))
	r.t.rail.Move(fyne.NewPos(0, 0))

	d := float32(toggleHeight - 2*togglePadding)
	r.t.knob.Resize(fyne.NewSize(d, d))
	x := float32(togglePadding)
	if r.t.on {
		x = toggleWidth - d - togglePadding
	}
	r.t.knob.Move(fyne.NewPos(x, togglePadding))
}

func (r *toggleRenderer) MinSize() fyne.Size { return fyne.NewSize(toggleWidth, toggleHeight) }

func (r *toggleRenderer) Refresh() {
	r.t.rail.FillColor = r.t.railColor()
	r.Layout(r.MinSize())
	r.t.rail.Refresh()
	r.t.knob.Refresh()
}

func (r *toggleRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *toggleRenderer) Destroy()                     {}

// toggleRow assemble un libellé à gauche + un switch aligné à droite, pour les
// listes d'options. onChange reçoit le nouvel état. Renvoie aussi le switch pour
// pouvoir le synchroniser (SetOn) depuis l'extérieur.
func toggleRow(label string, initial bool, onChange func(bool)) (fyne.CanvasObject, *toggleSwitch) {
	sw := newToggleSwitch(initial, onChange)
	lbl := widget.NewLabel(label)
	row := container.NewBorder(nil, nil, nil, container.NewCenter(sw), lbl)
	return row, sw
}
