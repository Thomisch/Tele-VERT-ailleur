package app

import (
	"math/rand"
	"testing"
	"time"
)

func newRng() *rand.Rand { return rand.New(rand.NewSource(42)) }

func TestNextKeepAliveIntervalFixed(t *testing.T) {
	rng := newRng()
	if got := nextKeepAliveInterval(false, rng); got != keepAliveInterval {
		t.Errorf("sans humanize : attendu %v, got %v", keepAliveInterval, got)
	}
}

func TestNextKeepAliveIntervalRandomBounded(t *testing.T) {
	rng := newRng()
	for i := 0; i < 1000; i++ {
		d := nextKeepAliveInterval(true, rng)
		if d < humanizeMinInterval || d > humanizeMaxInterval {
			t.Fatalf("intervalle hors bornes: %v (attendu [%v,%v])", d, humanizeMinInterval, humanizeMaxInterval)
		}
	}
}

func TestClampAltTabSeconds(t *testing.T) {
	cases := map[int]int{
		0:   altTabMinSeconds,
		3:   altTabMinSeconds,
		5:   5,
		10:  10,
		120: 120,
	}
	for in, want := range cases {
		if got := clampAltTabSeconds(in); got != want {
			t.Errorf("clampAltTabSeconds(%d)=%d, attendu %d", in, got, want)
		}
	}
}

func TestNextAltTabIntervalNeverBelowFloor(t *testing.T) {
	rng := newRng()
	floor := time.Duration(altTabMinSeconds) * time.Second
	// Même avec une base au plancher, la variation ne doit jamais descendre
	// sous le plancher de sécurité.
	for i := 0; i < 2000; i++ {
		d := nextAltTabInterval(5, rng)
		if d < floor {
			t.Fatalf("intervalle Alt+Tab sous le plancher: %v < %v", d, floor)
		}
	}
	// Une base sous le plancher est clampée d'abord.
	for i := 0; i < 2000; i++ {
		d := nextAltTabInterval(1, rng)
		if d < floor {
			t.Fatalf("base sous plancher non clampée: %v < %v", d, floor)
		}
	}
}

func TestHumanMouseStepsSumZero(t *testing.T) {
	rng := newRng()
	for i := 0; i < 200; i++ {
		steps := humanMouseSteps(rng)
		if len(steps) < 2 {
			t.Fatalf("trop peu de pas: %d", len(steps))
		}
		var sx, sy int32
		for _, s := range steps {
			sx += s[0]
			sy += s[1]
		}
		// Le dernier pas compense la somme : le curseur ne dérive pas.
		if sx != 0 || sy != 0 {
			t.Errorf("dérive du curseur: somme=(%d,%d), attendu (0,0)", sx, sy)
		}
	}
}
