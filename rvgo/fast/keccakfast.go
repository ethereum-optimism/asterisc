package fast

import (
	"reflect"
	_ "unsafe" // we use go:linkname

	"golang.org/x/crypto/sha3"
)

type keccakState struct {
	a [25]uint64 // main state of the hash
	// and other fields, unimportant
}

//go:noescape
//go:linkname keccakReset golang.org/x/crypto/sha3.(*state).Reset
func keccakReset(st *keccakState)

//go:noescape
//go:linkname keccakWrite golang.org/x/crypto/sha3.(*state).Write
func keccakWrite(st *keccakState, p []byte) (n int, err error)

//go:noescape
//go:linkname keccakRead golang.org/x/crypto/sha3.(*state).Read
func keccakRead(st *keccakState, out []byte) (n int, err error)

// example of how to get access to a hasher where the call arguments do not escape to the heap
var hasher = (*keccakState)(reflect.ValueOf(sha3.NewLegacyKeccak256()).UnsafePointer())

func hashPair(left, right [32]byte) (out [32]byte) {
	keccakReset(hasher)
	_, _ = keccakWrite(hasher, left[:])
	_, _ = keccakWrite(hasher, right[:])
	_, _ = keccakRead(hasher, out[:])
	return
}

func hash(data [64]byte) (out [32]byte) {
	keccakReset(hasher)
	_, _ = keccakWrite(hasher, data[:])
	_, _ = keccakRead(hasher, out[:])
	return
}
