//go:build windows

package app

import (
	"sync"
	"time"
)

// scheduleCheckInterval : fréquence à laquelle le scheduler réévalue les
// plages. 20 s garantit une bascule à moins d'une minute des bornes de plage.
const scheduleCheckInterval = 20 * time.Second

// Scheduler applique le mode planifié : quand il est actif, il force l'état du
// moteur selon les plages horaires (ON dans une plage, OFF en dehors). Quand il
// est inactif, il ne touche à rien (contrôle manuel pur).
type Scheduler struct {
	cfg    *Config
	engine *Engine

	mu      sync.Mutex
	enabled bool

	ticker *time.Ticker
	stop   chan struct{}

	// now permet d'injecter l'heure en test ; nil = time.Now en prod.
	now func() time.Time
}

func newScheduler(cfg *Config, engine *Engine) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		engine:  engine,
		enabled: cfg.ScheduleEnabled,
		now:     time.Now,
	}
}

// start lance la boucle d'évaluation et applique l'état immédiatement.
func (s *Scheduler) start() {
	s.stop = make(chan struct{})
	s.ticker = time.NewTicker(scheduleCheckInterval)
	s.apply() // état initial sans attendre le premier tick
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.apply()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *Scheduler) shutdown() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.stop != nil {
		close(s.stop)
	}
}

// SetEnabled bascule le mode planifié. En l'activant, on applique aussitôt
// l'état dicté par les plages. En le désactivant, on laisse le moteur dans son
// état courant (l'utilisateur reprend la main manuellement).
func (s *Scheduler) SetEnabled(v bool) {
	s.mu.Lock()
	s.enabled = v
	s.cfg.ScheduleEnabled = v
	s.mu.Unlock()
	_ = s.cfg.save()
	if v {
		s.apply()
	}
}

// IsEnabled renvoie l'état du mode planifié.
func (s *Scheduler) IsEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled
}

// SchedulesChanged signale que la liste des plages a été modifiée : si le mode
// est actif, on réapplique immédiatement.
func (s *Scheduler) SchedulesChanged() {
	if s.IsEnabled() {
		s.apply()
	}
}

// apply est le cœur : si le mode planifié est actif, force l'engine selon les
// plages. Sinon, ne fait rien.
func (s *Scheduler) apply() {
	s.mu.Lock()
	enabled := s.enabled
	schedules := s.cfg.Schedules
	s.mu.Unlock()

	if !enabled {
		return
	}
	shouldRun := anyScheduleActiveAt(schedules, s.now())
	s.engine.SetEnabled(shouldRun)
}
