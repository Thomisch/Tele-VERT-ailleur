//go:build windows

package app

import (
	"errors"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	errInvalidNumber = errors.New("nombre invalide")
	errBelowFloor    = errors.New("minimum 5 secondes")
)

// buildAdvancedTab : options de furtivité (randomisation) et mode Alt+Tab.
func (u *ui) buildAdvancedTab() fyne.CanvasObject {
	intro := widget.NewLabelWithStyle(
		"Options pour rendre l'activité simulée moins détectable.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)
	intro.Wrapping = fyne.TextWrapWord

	// --- Randomisation (humanize) ---
	humanizeRow, _ := toggleRow("Imiter un comportement humain", u.cfg.Humanize, func(on bool) {
		u.cfg.Humanize = on
		_ = u.cfg.save()
	})
	humanizeHelp := helpLabel(
		"Varie l'intervalle de réveil (≈40–95 s) et les mouvements de souris, " +
			"au lieu d'un geste régulier facile à repérer.")

	// --- Mode Alt+Tab ---
	altTabHelp := helpLabel(
		"Bascule de fenêtre toutes les X secondes (cumulable avec le maintien). " +
			"Minimum 5 s pour pouvoir reprendre la main.")

	// Champ intervalle Alt+Tab.
	secEntry := widget.NewEntry()
	secEntry.SetText(strconv.Itoa(u.cfg.AltTabSeconds))
	secEntry.Validator = altTabSecondsValidator
	commitSeconds := func(s string) {
		if n, err := strconv.Atoi(s); err == nil {
			u.cfg.AltTabSeconds = clampAltTabSeconds(n)
			_ = u.cfg.save()
		}
	}
	secEntry.OnChanged = commitSeconds

	secBox := container.NewGridWrap(fyne.NewSize(64, secEntry.MinSize().Height), secEntry)
	intervalRow := container.NewHBox(
		widget.NewLabel("Intervalle"),
		secBox,
		widget.NewLabel("secondes"),
	)

	altTabRow, _ := toggleRow("Activer le mode Alt+Tab", u.cfg.AltTabEnabled, func(on bool) {
		u.cfg.AltTabEnabled = on
		_ = u.cfg.save()
	})

	return container.NewVScroll(container.NewPadded(container.NewVBox(
		intro,
		widget.NewSeparator(),
		humanizeRow,
		humanizeHelp,
		widget.NewSeparator(),
		altTabRow,
		altTabHelp,
		intervalRow,
	)))
}

// helpLabel : petit texte d'aide gris, retour à la ligne automatique.
func helpLabel(s string) *widget.Label {
	l := widget.NewLabel(s)
	l.Wrapping = fyne.TextWrapWord
	l.Importance = widget.LowImportance
	return l
}

// altTabSecondsValidator : entier >= plancher de sécurité.
func altTabSecondsValidator(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return errInvalidNumber
	}
	if n < altTabMinSeconds {
		return errBelowFloor
	}
	return nil
}
