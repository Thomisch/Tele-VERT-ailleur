//go:build windows

package app

import "testing"

// Vérifie que setSleepPrevention pose puis retire bien le flag anti-veille.
// SetThreadExecutionState renvoie l'ÉTAT PRÉCÉDENT des flags : on s'en sert
// pour observer ce que notre fonction a réellement appliqué.
func TestSleepPreventionTogglesFlag(t *testing.T) {
	// État de départ propre.
	setSleepPrevention(false)

	// On arme l'anti-veille via le vrai code de prod.
	setSleepPrevention(true)

	// Un appel direct ES_CONTINUOUS renvoie l'état courant (= ce que prod a
	// posé) et le laisse inchangé puisqu'on repose ES_CONTINUOUS seul... mais
	// cela RELÂCHERAIT le flag. Pour ne pas perturber, on lit l'état précédent
	// en re-posant exactement le même flag que la prod.
	prev, _, _ := procSetThreadExecState.Call(uintptr(esContinuous | esSystemRequired))
	if prev == 0 {
		t.Fatalf("SetThreadExecutionState a renvoyé 0 (échec d'appel)")
	}
	// L'état précédent doit contenir ES_CONTINUOUS | ES_SYSTEM_REQUIRED.
	const wantMask = uintptr(esContinuous | esSystemRequired)
	if prev&wantMask != wantMask {
		t.Errorf("flag anti-veille absent: prev=0x%X, attendu masque 0x%X", prev, wantMask)
	}

	// On relâche et on vérifie que le flag SYSTEM_REQUIRED disparaît.
	setSleepPrevention(false)
	prev2, _, _ := procSetThreadExecState.Call(uintptr(esContinuous))
	if prev2&uintptr(esSystemRequired) != 0 {
		t.Errorf("ES_SYSTEM_REQUIRED toujours présent après relâche: prev=0x%X", prev2)
	}

	// Nettoyage : s'assure qu'on ne laisse rien d'armé.
	setSleepPrevention(false)
}

// Vérifie l'idempotence : deux appels identiques consécutifs ne changent rien.
func TestSleepPreventionIdempotent(t *testing.T) {
	setSleepPrevention(false)
	setSleepPrevention(true)
	if !sleepBlocked {
		t.Fatalf("sleepBlocked devrait être true")
	}
	setSleepPrevention(true) // 2e fois : no-op
	if !sleepBlocked {
		t.Fatalf("sleepBlocked devrait rester true")
	}
	setSleepPrevention(false)
	if sleepBlocked {
		t.Fatalf("sleepBlocked devrait être false")
	}
}
