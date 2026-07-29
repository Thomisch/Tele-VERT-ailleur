//go:build windows

package app

import "time"

// Run est le point d'entrée applicatif : instancie la config, le moteur, le
// scheduler et la GUI, puis bloque jusqu'à fermeture. Appelé par le main du
// package cmd. FuckTeamsStatus maintient la présence Teams en vert via une
// micro-activité d'entrée (F15 + jiggle souris), avec fenêtre Fyne et tray.
func Run() {
	// Instance unique : si une instance tourne déjà, on lui demande d'afficher
	// sa fenêtre puis on quitte immédiatement (pas de doublon caché).
	if !acquireSingleInstance() {
		signalShowExisting()
		return
	}

	cfg := loadConfig()

	// Écrit la config dès le démarrage : garantit qu'un config.json existe et
	// reste éditable même avant tout changement d'état (premier lancement).
	_ = cfg.save()

	// Synchronise le démarrage auto Windows avec la config (auto-correction
	// du chemin si l'exe a été déplacé).
	reconcileAutoStart(cfg.LaunchAtStartup)

	engine := newEngine(cfg.KeepAliveEnabled, func() bool { return cfg.Humanize }, time.Now().UnixNano())
	engine.start()
	defer engine.shutdown()
	// Relâche l'anti-veille à la fermeture pour ne pas laisser le système
	// bloqué éveillé après l'arrêt de l'appli.
	defer setSleepPrevention(false)

	// Mode Alt+Tab : boucle indépendante, cumulable avec le keep-alive.
	altTab := newAltTabber(cfg, time.Now().UnixNano())
	altTab.start()
	defer altTab.shutdown()

	// Quand le moteur change d'état, on persiste le toggle et on réapplique la
	// politique anti-veille (qui suit l'état effectif du maintien).
	engine.OnChange(func(enabled bool) {
		cfg.KeepAliveEnabled = enabled
		_ = cfg.save()
		applySleepPolicy(cfg, engine)
	})

	// État anti-veille initial cohérent avec la config au démarrage.
	applySleepPolicy(cfg, engine)

	// Mode planifié (Temps 2) : pilote l'engine selon les plages horaires.
	scheduler := newScheduler(cfg, engine)
	scheduler.start()
	defer scheduler.shutdown()

	// Lance la GUI Fyne (fenêtre + tray). Bloque jusqu'à fermeture.
	runUI(cfg, engine, scheduler)
}

// applySleepPolicy arme l'anti-veille si l'utilisateur l'a demandée ET que le
// maintien tourne effectivement. C'est le seul endroit qui décide de l'état
// anti-veille, pour éviter toute incohérence.
func applySleepPolicy(cfg *Config, engine *Engine) {
	setSleepPrevention(cfg.PreventSleep && engine.IsEnabled())
}
