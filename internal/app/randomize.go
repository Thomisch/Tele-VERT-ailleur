package app

import (
	"math/rand"
	"time"
)

// altTabMinSeconds : plancher de sécurité pour l'intervalle Alt+Tab. En dessous,
// on ne pourrait plus revenir à l'application entre deux bascules.
const altTabMinSeconds = 5

// Bornes de randomisation de l'intervalle keep-alive quand Humanize est actif.
// On reste sous le seuil d'inactivité de Teams (~5 min) avec une marge.
const (
	humanizeMinInterval = 40 * time.Second
	humanizeMaxInterval = 95 * time.Second
)

// nextKeepAliveInterval renvoie le délai avant le prochain réveil. Si humanize
// est faux, intervalle fixe (keepAliveInterval). Sinon, valeur aléatoire dans
// [humanizeMinInterval, humanizeMaxInterval].
func nextKeepAliveInterval(humanize bool, rng *rand.Rand) time.Duration {
	if !humanize {
		return keepAliveInterval
	}
	span := int64(humanizeMaxInterval - humanizeMinInterval)
	return humanizeMinInterval + time.Duration(rng.Int63n(span+1))
}

// clampAltTabSeconds applique le plancher de sécurité.
func clampAltTabSeconds(s int) int {
	if s < altTabMinSeconds {
		return altTabMinSeconds
	}
	return s
}

// nextAltTabInterval renvoie le délai avant le prochain Alt+Tab : base réglée
// par l'utilisateur (clampée au plancher) avec une variation aléatoire de
// ±25 % pour éviter une régularité robotique, sans jamais passer sous le
// plancher.
func nextAltTabInterval(baseSeconds int, rng *rand.Rand) time.Duration {
	base := clampAltTabSeconds(baseSeconds)
	// Variation dans [-25%, +25%].
	jitterRange := float64(base) * 0.25
	jitter := (rng.Float64()*2 - 1) * jitterRange
	secs := float64(base) + jitter
	if secs < float64(altTabMinSeconds) {
		secs = float64(altTabMinSeconds)
	}
	return time.Duration(secs * float64(time.Second))
}

// humanMouseSteps génère une séquence de petits déplacements relatifs (dx, dy)
// imitant un mouvement humain : plusieurs micro-pas d'amplitude et direction
// variables, plutôt qu'un unique aller-retour d'1 px. Somme proche de zéro pour
// ne pas faire dériver le curseur.
func humanMouseSteps(rng *rand.Rand) [][2]int32 {
	n := 3 + rng.Intn(5) // 3 à 7 pas
	steps := make([][2]int32, 0, n+1)
	var sumX, sumY int32
	for i := 0; i < n; i++ {
		dx := int32(rng.Intn(9) - 4) // -4..+4
		dy := int32(rng.Intn(9) - 4)
		steps = append(steps, [2]int32{dx, dy})
		sumX += dx
		sumY += dy
	}
	// Pas final qui ramène le curseur près de son point de départ.
	steps = append(steps, [2]int32{-sumX, -sumY})
	return steps
}
