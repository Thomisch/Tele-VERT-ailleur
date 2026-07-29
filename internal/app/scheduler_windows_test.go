//go:build windows

package app

import (
	"path/filepath"
	"testing"
	"time"
)

// newTestScheduler crée un scheduler avec une horloge injectée et une config
// jetable, sans démarrer la boucle (on appelle apply() manuellement).
func newTestScheduler(t *testing.T, scheduleEnabled bool, schedules []Schedule, now time.Time) (*Scheduler, *Engine) {
	t.Helper()
	cfg := defaultConfig()
	cfg.path = filepath.Join(t.TempDir(), "config.json")
	cfg.ScheduleEnabled = scheduleEnabled
	cfg.Schedules = schedules

	engine := newEngine(false, nil, 1)
	s := newScheduler(cfg, engine)
	s.now = func() time.Time { return now }
	return s, engine
}

func TestSchedulerForcesOnDuringRange(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	schedules := []Schedule{{Start: "09:00", End: "18:00", Days: lunVen}}

	// Lundi 10h, mode planifié actif : le moteur doit être forcé ON.
	s, engine := newTestScheduler(t, true, schedules, dateAt(time.Monday, 10, 0))
	s.apply()
	if !engine.IsEnabled() {
		t.Errorf("dans la plage : le moteur devrait être ON")
	}
}

func TestSchedulerForcesOffOutsideRange(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	schedules := []Schedule{{Start: "09:00", End: "18:00", Days: lunVen}}

	// Lundi 20h : hors plage, le moteur doit être forcé OFF.
	s, engine := newTestScheduler(t, true, schedules, dateAt(time.Monday, 20, 0))
	engine.SetEnabled(true) // on simule un état ON résiduel
	s.apply()
	if engine.IsEnabled() {
		t.Errorf("hors plage : le moteur devrait être OFF")
	}
}

func TestSchedulerWeekendOff(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	schedules := []Schedule{{Start: "09:00", End: "18:00", Days: lunVen}}

	// Samedi 10h : jour non couvert, moteur OFF.
	s, engine := newTestScheduler(t, true, schedules, dateAt(time.Saturday, 10, 0))
	engine.SetEnabled(true)
	s.apply()
	if engine.IsEnabled() {
		t.Errorf("samedi : le moteur devrait être OFF")
	}
}

func TestSchedulerDisabledDoesNotTouchEngine(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	schedules := []Schedule{{Start: "09:00", End: "18:00", Days: lunVen}}

	// Mode planifié OFF : apply() ne doit PAS modifier l'état manuel, même hors
	// plage. Ici on est lundi 20h (hors plage) mais le moteur reste ON.
	s, engine := newTestScheduler(t, false, schedules, dateAt(time.Monday, 20, 0))
	engine.SetEnabled(true)
	s.apply()
	if !engine.IsEnabled() {
		t.Errorf("mode planifié OFF : le moteur ne doit pas être touché (reste ON)")
	}
}
