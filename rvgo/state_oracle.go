package rvgo

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

type VMStateOracle interface {
	// Get a merkle pair by key (merkle hash of pair)
	// When replicating a step, this just loads two values from calldata, and verifies it matches the key.
	Get(key [32]byte) (a, b [32]byte)
	// Remember the pair of left and right, and return the merkle hash of the pair.
	// When replicating a step, this is just a pure merkle hash function.
	Remember(left [32]byte, right [32]byte) [32]byte
}

type StateOracle struct {
	data map[[32]byte][2][32]byte

	accessList [][32]byte
}

var _ VMStateOracle = (*StateOracle)(nil)

func NewStateOracle() *StateOracle {
	return &StateOracle{data: make(map[[32]byte][2][32]byte)}
}

func (s *StateOracle) Get(key [32]byte) (a, b [32]byte) {
	s.accessList = append(s.accessList, key)
	ab, ok := s.data[key]
	if !ok {
		panic(fmt.Errorf("missing key %x", key))
	}
	return ab[0], ab[1]
}

func (s *StateOracle) Remember(left [32]byte, right [32]byte) [32]byte {
	key := crypto.Keccak256Hash(left[:], right[:])
	s.data[key] = [2][32]byte{left, right}
	return key
}

type Access struct {
	Key   [32]byte
	Value [2][32]byte
}

func (s *StateOracle) AccessList() (out []Access) {
	out = make([]Access, len(s.accessList))
	for i, k := range s.accessList {
		out[i] = Access{
			Key:   k,
			Value: s.data[k],
		}
	}
	return out
}

type AccessListOracle struct {
	AccessList []Access
	Index      uint64
}

func (al *AccessListOracle) Get(key [32]byte) (a, b [32]byte) {
	access := al.AccessList[al.Index]
	if access.Key != key {
		panic("key mismatch")
	}
	return access.Value[0], access.Value[1]
}

func (al *AccessListOracle) Remember(left [32]byte, right [32]byte) [32]byte {
	// nothing to remember, just return the hash
	return crypto.Keccak256Hash(left[:], right[:])
}

var _ VMStateOracle = (*AccessListOracle)(nil)
