//go:build windows

package app

import (
	"math/rand"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32      = windows.NewLazySystemDLL("user32.dll")
	procSendIn  = user32.NewProc("SendInput")
	procGetCurs = user32.NewProc("GetCursorPos")
)

// Constantes Win32 pour SendInput.
const (
	inputMouse    = 0
	inputKeyboard = 1

	mouseeventfMove = 0x0001

	keyeventfKeyup = 0x0002

	vkF15  = 0x7E // touche F15 : existe mais sans effet visible
	vkMenu = 0x12 // ALT
	vkTab  = 0x09 // TAB
)

// Structures SendInput. On utilise la plus grande (mouse) comme buffer
// commun car keybdInput est plus petit que mouseInput.
type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type keybdInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
	_           uint32 // padding pour atteindre la taille de mouseInput
}

type input struct {
	inputType uint32
	_         uint32 // padding d'alignement (union sur 8 octets)
	mi        mouseInput
}

type point struct {
	x int32
	y int32
}

func sendInputs(inputs []input) {
	if len(inputs) == 0 {
		return
	}
	procSendIn.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
}

// pressF15 envoie un appui + relâchement de la touche F15.
func pressF15() {
	mk := func(up bool) input {
		var flags uint32
		if up {
			flags = keyeventfKeyup
		}
		kb := keybdInput{wVk: vkF15, dwFlags: flags}
		var in input
		in.inputType = inputKeyboard
		// On copie le keybdInput dans la zone mémoire partagée mi.
		*(*keybdInput)(unsafe.Pointer(&in.mi)) = kb
		return in
	}
	sendInputs([]input{mk(false), mk(true)})
}

// moveMouseRel envoie un déplacement relatif du curseur.
func moveMouseRel(dx, dy int32) {
	var in input
	in.inputType = inputMouse
	in.mi = mouseInput{dx: dx, dy: dy, dwFlags: mouseeventfMove}
	sendInputs([]input{in})
}

// jiggleMouse fait un micro déplacement relatif d'un pixel puis revient,
// pour que le retour curseur soit invisible à l'œil.
func jiggleMouse() {
	moveMouseRel(1, 0)
	moveMouseRel(-1, 0)
}

// jiggleMouseHuman exécute une séquence de petits déplacements variés avec de
// brefs délais entre eux, pour imiter un mouvement humain plutôt qu'un
// aller-retour robotique. La somme des pas est ~nulle (pas de dérive).
func jiggleMouseHuman(rng *rand.Rand) {
	for _, step := range humanMouseSteps(rng) {
		moveMouseRel(step[0], step[1])
		time.Sleep(time.Duration(8+rng.Intn(25)) * time.Millisecond)
	}
}

// pressAltTab envoie un Alt+Tab (maintien d'ALT, appui/relâche de TAB, relâche
// d'ALT) pour basculer vers la fenêtre suivante.
func pressAltTab() {
	key := func(vk uint16, up bool) input {
		var flags uint32
		if up {
			flags = keyeventfKeyup
		}
		kb := keybdInput{wVk: vk, dwFlags: flags}
		var in input
		in.inputType = inputKeyboard
		*(*keybdInput)(unsafe.Pointer(&in.mi)) = kb
		return in
	}
	sendInputs([]input{key(vkMenu, false)}) // ALT down
	sendInputs([]input{key(vkTab, false)})  // TAB down
	sendInputs([]input{key(vkTab, true)})   // TAB up
	sendInputs([]input{key(vkMenu, true)})  // ALT up
}

// keepAlive déclenche le maintien. Si humanize est vrai, le mouvement souris
// est varié ; sinon, micro aller-retour discret. La touche F15 est toujours
// envoyée (invisible).
func keepAlive(humanize bool, rng *rand.Rand) {
	pressF15()
	if humanize {
		jiggleMouseHuman(rng)
	} else {
		jiggleMouse()
	}
}
