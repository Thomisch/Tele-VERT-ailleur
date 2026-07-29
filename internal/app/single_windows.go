//go:build windows

package app

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Instance unique via objets noyau nommés :
//   - un mutex nommé détecte qu'une instance tourne déjà ;
//   - un event nommé sert de signal "montre ta fenêtre" : la 1re instance
//     l'attend en boucle, les lancements suivants le déclenchent puis quittent.
const (
	mutexName = "Global\\FuckTeamsStatus_SingleInstance_Mutex"
	eventName = "Global\\FuckTeamsStatus_Show_Event"
)

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
	procCreateEvent = kernel32.NewProc("CreateEventW")
	procOpenEvent   = kernel32.NewProc("OpenEventW")
	procSetEvent    = kernel32.NewProc("SetEvent")
)

// acquireSingleInstance tente de devenir l'instance maîtresse.
// Renvoie isPrimary=true si c'est la première instance, false sinon.
func acquireSingleInstance() (isPrimary bool) {
	name, _ := windows.UTF16PtrFromString(mutexName)
	// CreateMutex renvoie un handle ; GetLastError == ERROR_ALREADY_EXISTS si
	// le mutex existait déjà (donc une autre instance tourne).
	_, _, callErr := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	// procCreateMutex.Call place GetLastError dans callErr (type syscall.Errno).
	return callErr != windows.ERROR_ALREADY_EXISTS
}

// signalShowExisting déclenche l'event "montre ta fenêtre" pour réveiller
// l'instance déjà en cours. Appelé par une instance secondaire avant de quitter.
func signalShowExisting() {
	name, _ := windows.UTF16PtrFromString(eventName)
	const eventModifyState = 0x0002
	h, _, _ := procOpenEvent.Call(eventModifyState, 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return
	}
	defer windows.CloseHandle(windows.Handle(h))
	_, _, _ = procSetEvent.Call(h)
}

// watchShowRequests crée l'event nommé et appelle onShow chaque fois qu'une
// autre instance le déclenche. À lancer dans une goroutine par l'instance
// maîtresse.
func watchShowRequests(onShow func()) {
	name, _ := windows.UTF16PtrFromString(eventName)
	// Event à réinitialisation automatique, non signalé au départ.
	h, _, _ := procCreateEvent.Call(0, 0, 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return
	}
	handle := windows.Handle(h)
	for {
		// Attente bloquante jusqu'au signal.
		s, _ := windows.WaitForSingleObject(handle, windows.INFINITE)
		if s == windows.WAIT_OBJECT_0 {
			onShow()
		}
	}
}
