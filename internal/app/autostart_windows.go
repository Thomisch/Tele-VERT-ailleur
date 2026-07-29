//go:build windows

package app

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

// Clé Run de l'utilisateur courant : les valeurs y sont lancées à l'ouverture
// de session. On y stocke notre exe sous un nom dédié.
const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue   = "FuckTeamsStatus"
)

// exePath renvoie le chemin absolu de l'exécutable courant, entre guillemets
// pour gérer les espaces dans le chemin.
func exePath() string {
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	return `"` + p + `"`
}

// setAutoStart active ou désactive le lancement au démarrage en ajoutant ou
// retirant la valeur de registre.
func setAutoStart(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if enabled {
		return key.SetStringValue(runValue, exePath())
	}
	// Suppression : on ignore l'erreur "valeur inexistante".
	err = key.DeleteValue(runValue)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// isAutoStartEnabled indique si la valeur de registre est présente.
func isAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	_, _, err = key.GetStringValue(runValue)
	return err == nil
}

// reconcileAutoStart synchronise le registre avec l'état voulu par la config,
// et corrige le chemin si l'exe a été déplacé/renommé. Appelé au démarrage.
func reconcileAutoStart(want bool) {
	if !want {
		// L'utilisateur ne veut pas l'autostart : on s'assure qu'il est absent.
		if isAutoStartEnabled() {
			_ = setAutoStart(false)
		}
		return
	}

	// L'utilisateur le veut : on s'assure que l'entrée existe ET pointe vers
	// le chemin réel courant (auto-correction).
	wantPath := exePath()
	if wantPath == "" {
		return
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	current, _, err := key.GetStringValue(runValue)
	if err != nil || current != wantPath {
		_ = key.SetStringValue(runValue, wantPath)
	}
}
