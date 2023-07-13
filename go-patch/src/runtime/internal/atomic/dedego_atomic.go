package atomic

// DEDEGO-FILE-START

import (
	"unsafe"
)

var com chan<- AtomicElem
var linked bool
var counter uint64

// TODO: protect counter

type AtomicElem struct {
	Index uint64
	Addr  uint64
}

func DedegoAtomicLink(c chan<- AtomicElem) {
	com = c
	linked = true
}

func DedegoAtomicUnlink() {
	com = nil
	linked = false
}

func DedegoAtomic64(addr *uint64) {
	if linked {
		counter += 1
		com <- AtomicElem{Index: counter, Addr: uint64(uintptr(unsafe.Pointer(addr)))}
	}
}

func DedegoAtomic32(addr *uint32) {
	if linked {
		counter += 1
		com <- AtomicElem{Index: counter, Addr: uint64(uintptr(unsafe.Pointer(addr)))}
	}
}

func DedegoAtomicUIntPtr(addr *uintptr) {
	if linked {
		counter += 1
		com <- AtomicElem{Index: counter, Addr: uint64(uintptr(unsafe.Pointer(addr)))}
	}
}

func DedegoAtomicPtr(addr unsafe.Pointer) {
	if linked {
		counter += 1
		com <- AtomicElem{Index: counter, Addr: uint64(uintptr(addr))}
	}
}

// DEDEGO-FILE-END
