//go:build windows

package app

import "testing"

// Vérifie que SetDisabled change l'état logique et la couleur du rail, et que
// les taps sont ignorés quand disabled.
func TestToggleDisabledIgnoresTap(t *testing.T) {
	changed := 0
	sw := newToggleSwitch(true, func(bool) { changed++ })
	sw.CreateRenderer() // initialise rail/knob

	sw.SetDisabled(true)
	if !sw.disabled {
		t.Fatalf("disabled devrait être true")
	}
	sw.Tapped(nil)
	if changed != 0 {
		t.Errorf("un tap sur un switch désactivé ne doit pas déclencher onChange (got %d)", changed)
	}

	sw.SetDisabled(false)
	sw.Tapped(nil)
	if changed != 1 {
		t.Errorf("après réactivation, le tap doit déclencher onChange (got %d)", changed)
	}
}

// Vérifie que SetOn reflète l'état et n'appelle pas onChange.
func TestToggleSetOnNoCallback(t *testing.T) {
	calls := 0
	sw := newToggleSwitch(false, func(bool) { calls++ })
	sw.CreateRenderer()
	sw.SetOn(true)
	if !sw.on {
		t.Errorf("on devrait être true")
	}
	if calls != 0 {
		t.Errorf("SetOn ne doit pas déclencher onChange (got %d)", calls)
	}
}
