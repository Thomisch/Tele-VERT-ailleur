package app

import (
	"os"
	"path/filepath"
	"testing"
)

// Vérifie le cycle complet save -> reload de la config, y compris les champs
// du Temps 2 (schedules), pour garantir que rien n'est perdu sur disque.
func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	c := defaultConfig()
	c.path = path
	c.KeepAliveEnabled = false
	c.LaunchAtStartup = true
	c.ScheduleEnabled = true
	c.Schedules = []Schedule{{Start: "09:00", End: "12:00", Days: []int{1, 2, 3, 4, 5}}}

	if err := c.save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("fichier non écrit: %v", err)
	}

	// Recharge depuis le même chemin via le vrai code de chargement.
	reloaded := loadConfigFrom(path)

	if reloaded.KeepAliveEnabled != false {
		t.Errorf("KeepAliveEnabled: attendu false, got %v", reloaded.KeepAliveEnabled)
	}
	if reloaded.LaunchAtStartup != true {
		t.Errorf("LaunchAtStartup: attendu true, got %v", reloaded.LaunchAtStartup)
	}
	if !reloaded.ScheduleEnabled {
		t.Errorf("ScheduleEnabled: attendu true")
	}
	if len(reloaded.Schedules) != 1 || reloaded.Schedules[0].Start != "09:00" {
		t.Errorf("Schedules mal restaurées: %+v", reloaded.Schedules)
	}
}

// Vérifie que l'Engine notifie ses abonnés et persiste via le callback, comme
// le câblage réel de main().
func TestEngineNotifiesAndPersists(t *testing.T) {
	dir := t.TempDir()
	c := defaultConfig()
	c.path = filepath.Join(dir, "config.json")
	c.KeepAliveEnabled = true

	e := newEngine(c.KeepAliveEnabled, nil, 1)

	var notified []bool
	e.OnChange(func(v bool) { notified = append(notified, v) })
	e.OnChange(func(v bool) {
		c.KeepAliveEnabled = v
		_ = c.save()
	})

	e.SetEnabled(false) // changement
	e.SetEnabled(false) // pas de changement -> pas de notif
	e.SetEnabled(true)  // changement

	if len(notified) != 2 {
		t.Fatalf("attendu 2 notifications, got %d (%v)", len(notified), notified)
	}
	if notified[0] != false || notified[1] != true {
		t.Errorf("ordre des notifications inattendu: %v", notified)
	}
	if !c.KeepAliveEnabled {
		t.Errorf("config non synchronisée: attendu true")
	}
}
