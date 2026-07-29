package app

import (
	"os"
	"path/filepath"
	"testing"
)

// Un fichier config préfixé d'un BOM UTF-8 (ajouté par certains éditeurs
// Windows) doit quand même être chargé correctement.
func TestLoadHandlesUTF8BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	bom := []byte{0xEF, 0xBB, 0xBF}
	body := []byte(`{"scheduleEnabled": true, "schedules": [{"start":"08:00","end":"12:00","days":[1]}]}`)
	if err := os.WriteFile(path, append(bom, body...), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := loadConfigFrom(path)
	if !cfg.ScheduleEnabled {
		t.Errorf("ScheduleEnabled devrait être true malgré le BOM")
	}
	if len(cfg.Schedules) != 1 {
		t.Errorf("la plage devrait être chargée malgré le BOM, got %d", len(cfg.Schedules))
	}
}

// Reproduit le bug : un JSON avec scheduleEnabled=true et une plage doit être
// rechargé tel quel (et non écrasé par les défauts).
func TestLoadPreservesScheduleFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
  "keepAliveEnabled": true,
  "scheduleEnabled": true,
  "schedules": [ { "start": "14:00", "end": "17:30", "days": [1,2,3,4,5] } ],
  "humanize": true,
  "altTabEnabled": true,
  "altTabSeconds": 12
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := loadConfigFrom(path)
	if !cfg.ScheduleEnabled {
		t.Errorf("ScheduleEnabled devrait être true")
	}
	if len(cfg.Schedules) != 1 {
		t.Fatalf("attendu 1 plage, got %d", len(cfg.Schedules))
	}
	if cfg.Schedules[0].Start != "14:00" {
		t.Errorf("plage mal chargée: %+v", cfg.Schedules[0])
	}
	if !cfg.Humanize || !cfg.AltTabEnabled || cfg.AltTabSeconds != 12 {
		t.Errorf("champs avancés mal chargés: humanize=%v altTab=%v sec=%d",
			cfg.Humanize, cfg.AltTabEnabled, cfg.AltTabSeconds)
	}

	// Et un save() suivi d'un reload doit préserver (le save initial de main).
	if err := cfg.save(); err != nil {
		t.Fatal(err)
	}
	again := loadConfigFrom(path)
	if !again.ScheduleEnabled || len(again.Schedules) != 1 {
		t.Errorf("save+reload a perdu les données: enabled=%v n=%d", again.ScheduleEnabled, len(again.Schedules))
	}
}
