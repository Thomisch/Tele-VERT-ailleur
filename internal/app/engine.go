package app

import (
	"math/rand"
	"sync"
	"time"
)

// keepAliveInterval : fréquence des réveils en mode non randomisé. 60 s est
// largement sous le seuil d'inactivité de Teams (~5 min).
const keepAliveInterval = 60 * time.Second

// Engine pilote la boucle de maintien de présence. L'état "enabled" est la
// source de vérité partagée entre le tray, la GUI et le scheduler. Les abonnés
// (listeners) sont notifiés à chaque changement d'état pour rester synchronisés.
type Engine struct {
	mu        sync.Mutex
	enabled   bool
	listeners []func(bool)
	stop      chan struct{}

	// humanize indique si l'intervalle et les mouvements sont randomisés.
	// Fonction pour refléter en direct le réglage de la config.
	humanize func() bool

	rng *rand.Rand
}

// newEngine crée le moteur. humanize peut être nil (équivaut à "jamais").
func newEngine(initial bool, humanize func() bool, seed int64) *Engine {
	if humanize == nil {
		humanize = func() bool { return false }
	}
	return &Engine{
		enabled:  initial,
		humanize: humanize,
		rng:      rand.New(rand.NewSource(seed)),
	}
}

// fire déclenche un réveil en respectant le réglage humanize courant.
func (e *Engine) fire() {
	keepAlive(e.humanize(), e.rng)
}

// start lance la boucle de fond. L'intervalle est reprogrammé à chaque cycle
// via un timer (fixe ou aléatoire selon humanize).
func (e *Engine) start() {
	e.stop = make(chan struct{})

	// Premier coup immédiat si déjà actif, pour ne pas attendre l'intervalle.
	if e.IsEnabled() {
		e.fire()
	}

	go func() {
		timer := time.NewTimer(nextKeepAliveInterval(e.humanize(), e.rng))
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				if e.IsEnabled() {
					e.fire()
				}
				timer.Reset(nextKeepAliveInterval(e.humanize(), e.rng))
			case <-e.stop:
				return
			}
		}
	}()
}

// shutdown arrête proprement la boucle.
func (e *Engine) shutdown() {
	if e.stop != nil {
		close(e.stop)
	}
}

// IsEnabled renvoie l'état courant.
func (e *Engine) IsEnabled() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.enabled
}

// SetEnabled change l'état, notifie les abonnés, et réveille immédiatement le
// système si on vient d'activer. fireImmediate=false évite le réveil instantané
// (utile quand l'appel vient du scheduler à chaque tick).
func (e *Engine) SetEnabled(v bool) {
	e.mu.Lock()
	changed := e.enabled != v
	e.enabled = v
	listeners := make([]func(bool), len(e.listeners))
	copy(listeners, e.listeners)
	e.mu.Unlock()

	if !changed {
		return
	}
	if v {
		e.fire() // réveil immédiat à l'activation
	}
	for _, l := range listeners {
		l(v)
	}
}

// Toggle inverse l'état et renvoie la nouvelle valeur.
func (e *Engine) Toggle() bool {
	e.SetEnabled(!e.IsEnabled())
	return e.IsEnabled()
}

// OnChange enregistre un abonné notifié à chaque changement d'état.
func (e *Engine) OnChange(fn func(bool)) {
	e.mu.Lock()
	e.listeners = append(e.listeners, fn)
	e.mu.Unlock()
}
