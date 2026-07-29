package app

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseHM(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"09:00", 540, false},
		{"00:00", 0, false},
		{"23:59", 1439, false},
		{"14:30", 870, false},
		{" 8:05 ", 485, false}, // espaces tolérés
		{"24:00", 0, true},     // heure hors borne
		{"12:60", 0, true},     // minutes hors borne
		{"abc", 0, true},
		{"12", 0, true},
		{"12:00:00", 0, true},
	}
	for _, c := range cases {
		got, err := parseHM(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseHM(%q): erreur attendue, got %d", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseHM(%q): erreur inattendue %v", c.in, err)
		} else if got != c.want {
			t.Errorf("parseHM(%q) = %d, attendu %d", c.in, got, c.want)
		}
	}
}

func TestScheduleValid(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	cases := []struct {
		name string
		s    Schedule
		want bool
	}{
		{"normale", Schedule{Start: "09:00", End: "18:00", Days: lunVen}, true},
		{"fin avant debut", Schedule{Start: "18:00", End: "09:00", Days: lunVen}, false},
		{"fin egale debut", Schedule{Start: "09:00", End: "09:00", Days: lunVen}, false},
		{"aucun jour", Schedule{Start: "09:00", End: "18:00", Days: []int{}}, false},
		{"heure invalide", Schedule{Start: "25:00", End: "18:00", Days: lunVen}, false},
	}
	for _, c := range cases {
		if got := c.s.valid(); got != c.want {
			t.Errorf("%s: valid()=%v, attendu %v", c.name, got, c.want)
		}
	}
}

// dateAt construit un time.Time à un jour/heure donnés (semaine de référence).
// 2026-06-15 est un lundi.
func dateAt(weekday time.Weekday, hour, min int) time.Time {
	// lundi 15 juin 2026 = base ; on décale selon le weekday voulu.
	base := time.Date(2026, 6, 15, 0, 0, 0, 0, time.Local) // lundi
	offset := int(weekday) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return base.AddDate(0, 0, offset).Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
}

func TestScheduleActiveAt(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	s := Schedule{Start: "09:00", End: "18:00", Days: lunVen}

	cases := []struct {
		name string
		when time.Time
		want bool
	}{
		{"lundi 10h dans plage", dateAt(time.Monday, 10, 0), true},
		{"lundi 09h00 borne debut incluse", dateAt(time.Monday, 9, 0), true},
		{"lundi 18h00 borne fin exclue", dateAt(time.Monday, 18, 0), false},
		{"lundi 17h59 juste avant fin", dateAt(time.Monday, 17, 59), true},
		{"lundi 08h59 avant plage", dateAt(time.Monday, 8, 59), false},
		{"samedi 10h hors jours", dateAt(time.Saturday, 10, 0), false},
		{"dimanche 10h hors jours", dateAt(time.Sunday, 10, 0), false},
		{"vendredi 12h dans plage", dateAt(time.Friday, 12, 0), true},
	}
	for _, c := range cases {
		if got := s.activeAt(c.when); got != c.want {
			t.Errorf("%s: activeAt=%v, attendu %v", c.name, got, c.want)
		}
	}
}

func TestDisabledScheduleNeverActive(t *testing.T) {
	lunVen := []int{1, 2, 3, 4, 5}
	// Plage qui serait active (lundi 10h dans 09:00-18:00) mais désactivée.
	s := Schedule{Start: "09:00", End: "18:00", Days: lunVen, Disabled: true}
	if s.Enabled() {
		t.Errorf("Enabled() devrait être false quand Disabled=true")
	}
	if s.activeAt(dateAt(time.Monday, 10, 0)) {
		t.Errorf("une plage désactivée ne doit jamais être active")
	}
	if anyScheduleActiveAt([]Schedule{s}, dateAt(time.Monday, 10, 0)) {
		t.Errorf("anyScheduleActiveAt doit ignorer les plages désactivées")
	}
}

func TestScheduleBackwardCompatEnabledByDefault(t *testing.T) {
	// Un config.json existant n'a pas le champ "disabled". Après désérialisation,
	// Disabled vaut false → la plage doit être considérée comme active.
	jsonOld := `{"start":"09:00","end":"18:00","days":[1,2,3,4,5]}`
	var s Schedule
	if err := json.Unmarshal([]byte(jsonOld), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !s.Enabled() {
		t.Errorf("une plage sans champ disabled doit être active (rétrocompat)")
	}
}

func TestAnyScheduleActiveAt(t *testing.T) {
	matin := Schedule{Start: "09:00", End: "12:00", Days: []int{1, 2, 3, 4, 5}}
	aprem := Schedule{Start: "14:00", End: "18:00", Days: []int{1, 2, 3, 4, 5}}
	schedules := []Schedule{matin, aprem}

	if anyScheduleActiveAt(schedules, dateAt(time.Monday, 13, 0)) {
		t.Errorf("13h (pause déj) ne devrait être dans aucune plage")
	}
	if !anyScheduleActiveAt(schedules, dateAt(time.Monday, 10, 0)) {
		t.Errorf("10h devrait être dans la plage du matin")
	}
	if !anyScheduleActiveAt(schedules, dateAt(time.Monday, 15, 0)) {
		t.Errorf("15h devrait être dans la plage de l'après-midi")
	}
	if anyScheduleActiveAt(schedules, dateAt(time.Sunday, 10, 0)) {
		t.Errorf("dimanche ne devrait jamais être actif")
	}
	if anyScheduleActiveAt(nil, dateAt(time.Monday, 10, 0)) {
		t.Errorf("liste vide ne devrait jamais être active")
	}
}
