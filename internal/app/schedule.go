package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseHM convertit "HH:MM" en minutes depuis minuit. Renvoie une erreur si le
// format est invalide (heures 0-23, minutes 0-59).
func parseHM(s string) (int, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("format attendu HH:MM, reçu %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("heure invalide dans %q", s)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("minutes invalides dans %q", s)
	}
	return h*60 + m, nil
}

// valid indique si une plage est exploitable (heures parsables, au moins un
// jour coché, fin strictement après début). Les plages invalides sont ignorées
// par le scheduler plutôt que de bloquer.
func (s Schedule) valid() bool {
	start, err1 := parseHM(s.Start)
	end, err2 := parseHM(s.End)
	if err1 != nil || err2 != nil {
		return false
	}
	if end <= start {
		return false // on ne gère pas les plages traversant minuit
	}
	return len(s.Days) > 0
}

// containsDay indique si la plage couvre ce jour de la semaine.
func (s Schedule) containsDay(d time.Weekday) bool {
	for _, day := range s.Days {
		if time.Weekday(day) == d {
			return true
		}
	}
	return false
}

// activeAt indique si CETTE plage est active au moment t. Intervalle
// semi-ouvert [start, end) : 18:00 exactement n'est plus dans 09:00-18:00.
// Une plage désactivée n'est jamais active.
func (s Schedule) activeAt(t time.Time) bool {
	if s.Disabled {
		return false
	}
	if !s.valid() {
		return false
	}
	if !s.containsDay(t.Weekday()) {
		return false
	}
	start, _ := parseHM(s.Start)
	end, _ := parseHM(s.End)
	now := t.Hour()*60 + t.Minute()
	return now >= start && now < end
}

// anyScheduleActiveAt indique si AU MOINS une plage de la liste est active au
// moment t. C'est la décision centrale du mode planifié.
func anyScheduleActiveAt(schedules []Schedule, t time.Time) bool {
	for _, s := range schedules {
		if s.activeAt(t) {
			return true
		}
	}
	return false
}
