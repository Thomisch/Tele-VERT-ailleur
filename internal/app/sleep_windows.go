//go:build windows

package app

import (
	"sync"
)

// Anti-veille via SetThreadExecutionState : on déclare à Windows que le système
// doit rester éveillé tant que notre flag continu est posé. Méthode officielle
// (utilisée par les lecteurs vidéo), fiable et instantanément réversible.
//
// On empêche la veille SYSTÈME uniquement (pas l'extinction de l'écran) : on
// n'utilise donc pas ES_DISPLAY_REQUIRED.
const (
	esContinuous     = 0x80000000 // l'état reste actif jusqu'à nouvel appel
	esSystemRequired = 0x00000001 // empêche la veille système
)

var (
	procSetThreadExecState = kernel32.NewProc("SetThreadExecutionState")

	sleepMu      sync.Mutex
	sleepBlocked bool
)

// setSleepPrevention pose ou retire le flag anti-veille. Idempotent : appeler
// deux fois avec la même valeur ne fait rien.
func setSleepPrevention(block bool) {
	sleepMu.Lock()
	defer sleepMu.Unlock()

	if block == sleepBlocked {
		return
	}
	if block {
		// ES_CONTINUOUS | ES_SYSTEM_REQUIRED : maintient le système éveillé.
		procSetThreadExecState.Call(uintptr(esContinuous | esSystemRequired))
	} else {
		// ES_CONTINUOUS seul : relâche la demande, retour au comportement normal.
		procSetThreadExecState.Call(uintptr(esContinuous))
	}
	sleepBlocked = block
}
