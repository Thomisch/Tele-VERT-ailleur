//go:build windows

package app

import (
	"math/rand"
	"time"
)

// AltTabber fait des Alt+Tab périodiques quand l'option est activée. Boucle
// indépendante du keep-alive (cumulable). Lit son état en direct depuis la
// config, donc activer/désactiver l'option prend effet sans redémarrage.
type AltTabber struct {
	cfg  *Config
	rng  *rand.Rand
	stop chan struct{}
}

func newAltTabber(cfg *Config, seed int64) *AltTabber {
	return &AltTabber{
		cfg: cfg,
		rng: rand.New(rand.NewSource(seed)),
	}
}

// start lance la boucle. Le timer est reprogrammé à chaque cycle avec un
// intervalle aléatoire autour de la base réglée (plancher de sécurité 5 s).
func (a *AltTabber) start() {
	a.stop = make(chan struct{})
	go func() {
		timer := time.NewTimer(a.nextDelay())
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				if a.cfg.AltTabEnabled {
					pressAltTab()
				}
				timer.Reset(a.nextDelay())
			case <-a.stop:
				return
			}
		}
	}()
}

func (a *AltTabber) shutdown() {
	if a.stop != nil {
		close(a.stop)
	}
}

// nextDelay calcule le prochain intervalle d'après la config (clampé au
// plancher) avec variation aléatoire.
func (a *AltTabber) nextDelay() time.Duration {
	return nextAltTabInterval(a.cfg.AltTabSeconds, a.rng)
}
