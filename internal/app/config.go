package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Schedule représente une plage horaire récurrente (Temps 2).
// Start/End au format "HH:MM" (24h). Days : 0=dimanche … 6=samedi.
//
// Disabled (et non "Enabled") est volontaire : avec omitempty, une plage issue
// d'un ancien config.json sans ce champ vaut Disabled=false, donc reste ACTIVE
// par défaut (rétrocompatibilité). Côté UI on expose l'inverse via Enabled().
type Schedule struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Days     []int  `json:"days"`
	Disabled bool   `json:"disabled,omitempty"`
}

// Enabled indique si la plage est active (inverse de Disabled).
func (s Schedule) Enabled() bool { return !s.Disabled }

// Config est l'état persistant complet de l'application.
type Config struct {
	// KeepAliveEnabled : état du toggle manuel (Temps 1).
	KeepAliveEnabled bool `json:"keepAliveEnabled"`
	// LaunchAtStartup : lancer l'appli au démarrage de Windows.
	LaunchAtStartup bool `json:"launchAtStartup"`
	// PreventSleep : empêcher la mise en veille système tant que le maintien
	// tourne effectivement.
	PreventSleep bool `json:"preventSleep"`
	// ScheduleEnabled : mode planifié (Temps 2). Indépendant du toggle manuel.
	ScheduleEnabled bool `json:"scheduleEnabled"`
	// Schedules : lots de plages horaires (Temps 2).
	Schedules []Schedule `json:"schedules"`

	// Humanize : randomise l'intervalle de réveil et les mouvements souris pour
	// imiter un comportement humain (au lieu d'un geste robotique régulier).
	Humanize bool `json:"humanize"`

	// AltTabEnabled : en plus du keep-alive, fait des Alt+Tab périodiques.
	AltTabEnabled bool `json:"altTabEnabled"`
	// AltTabSeconds : intervalle de base entre deux Alt+Tab, en secondes
	// (plancher de sécurité à altTabMinSeconds).
	AltTabSeconds int `json:"altTabSeconds"`

	mu   sync.Mutex `json:"-"`
	path string     `json:"-"`
}

// defaultConfig renvoie l'état d'un premier lancement : keep-alive actif,
// reste neutre.
func defaultConfig() *Config {
	return &Config{
		KeepAliveEnabled: true,
		LaunchAtStartup:  false,
		ScheduleEnabled:  false,
		Schedules:        []Schedule{},
		Humanize:         false,
		AltTabEnabled:    false,
		AltTabSeconds:    10, // défaut raisonnable, au-dessus du plancher 5s
	}
}

// configPath renvoie %APPDATA%\FuckTeamsStatus\config.json (avec repli sur le
// dossier de l'exe si APPDATA est absent).
func configPath() string {
	dir, err := os.UserConfigDir() // %APPDATA% sur Windows
	if err != nil || dir == "" {
		if exe, e := os.Executable(); e == nil {
			dir = filepath.Dir(exe)
		} else {
			dir = "."
		}
	}
	return filepath.Join(dir, "FuckTeamsStatus", "config.json")
}

// loadConfig lit la config depuis l'emplacement standard.
func loadConfig() *Config {
	return loadConfigFrom(configPath())
}

// loadConfigFrom lit la config depuis un chemin donné ; en cas d'absence ou
// d'erreur de lecture, renvoie les valeurs par défaut (jamais d'échec bloquant
// au démarrage). Extrait pour être testable.
func loadConfigFrom(path string) *Config {
	cfg := defaultConfig()
	cfg.path = path

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg // premier lancement ou illisible : défauts
	}
	// Retire un éventuel BOM UTF-8 en tête : un éditeur Windows peut en ajouter
	// un, et encoding/json refuse de parser un document qui commence par un BOM.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	// On décode dans une struct temporaire pour ne pas écraser mu/path.
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		return cfg // fichier corrompu : on repart sur les défauts
	}
	cfg.KeepAliveEnabled = loaded.KeepAliveEnabled
	cfg.LaunchAtStartup = loaded.LaunchAtStartup
	cfg.PreventSleep = loaded.PreventSleep
	cfg.ScheduleEnabled = loaded.ScheduleEnabled
	if loaded.Schedules != nil {
		cfg.Schedules = loaded.Schedules
	}
	cfg.Humanize = loaded.Humanize
	cfg.AltTabEnabled = loaded.AltTabEnabled
	if loaded.AltTabSeconds > 0 {
		cfg.AltTabSeconds = loaded.AltTabSeconds // garde le défaut si absent/0
	}
	return cfg
}

// save écrit la config sur disque de façon atomique (écriture dans un .tmp
// puis renommage) pour éviter un fichier à moitié écrit.
func (c *Config) save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}
