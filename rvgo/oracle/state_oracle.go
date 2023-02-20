package oracle

import (
	"fmt"
	"math/bits"
	"strings"

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

type Differ interface {
	Diff(a [32]byte, b [32]byte, gindex uint64)
}

type StateOracle struct {
	data    map[[32]byte][2][32]byte
	reverse map[[2][32]byte][32]byte

	accessList      [][32]byte
	buildAccessList bool
}

func (s *StateOracle) Dump(stateRoot [32]byte) string {
	var zeroHashes [256][32]byte
	for i := 1; i < 256; i++ {
		zeroHashes[i] = crypto.Keccak256Hash(zeroHashes[i-1][:], zeroHashes[i-1][:])
	}

	lookup := func(root [32]byte) ([2][32]byte, bool) {
		ab, ok := s.data[root]
		if ok {
			if root == stateRoot {
				return ab, true
			}
			for _, k := range s.accessList {
				if root == k {
					return ab, true
				}
			}
			return [2][32]byte{}, false
		} else {
			return [2][32]byte{}, false
		}
	}

	var out strings.Builder
	var vizMemory func(memRoot [32]byte, addrOffset uint64, depth uint64)
	vizMemory = func(memRoot [32]byte, addrOffset uint64, depth uint64) {
		if depth == 64-5 {
			out.WriteString(fmt.Sprintf("value %016x: %x\n", addrOffset, memRoot[:]))
			return
		}
		//for i, h := range zeroHashes {
		//	if h == memRoot {
		//		out.WriteString(fmt.Sprintf("ignore %d\n", i))
		//		return
		//	}
		//}
		if zeroHashes[64-5-depth+1] == memRoot {
			out.WriteString(fmt.Sprintf("range %016x-%016x: 000...000\n", addrOffset, addrOffset+(1<<(64-depth))))
			return
		}
		ab, ok := lookup(memRoot)
		if !ok {
			out.WriteString(fmt.Sprintf("unknown %016x (depth %d): %x\n", addrOffset, depth, memRoot))
			return
		}
		vizMemory(ab[0], addrOffset, depth+1)
		vizMemory(ab[1], addrOffset+(1<<(64-depth-1)), depth+1)
	}
	vizRootMem := func(memRoot [32]byte) {
		vizMemory(memRoot, 0, 0)
	}
	vizUint64 := func(root [32]byte) {
		out.WriteString(fmt.Sprintf("value: %x    (full: %x)\n", root[:8], root[:]))
	}
	var vizRegs func(root [32]byte, num uint8, depth uint8)
	vizRegs = func(root [32]byte, num uint8, depth uint8) {
		if depth == 5 {
			out.WriteString(fmt.Sprintf("%2d: %016x   (full: %x)\n", num, root[:8], root))
			return
		}
		ab, ok := lookup(root)
		if !ok {
			out.WriteString(fmt.Sprintf("missing registers %d - %d\n", num, num+(32>>depth)))
			return
		}
		vizRegs(ab[0], num, depth+1)
		vizRegs(ab[1], num+(32>>(depth+1)), depth+1)
	}
	//vizRootRegs := func(root [32]byte) {
	//	vizRegs(root, 0, 0)
	//}
	var vizMaybe func(name string, gindex uint64, root [32]byte, fn func(memRoot [32]byte))
	vizMaybe = func(name string, gindex uint64, root [32]byte, fn func(memRoot [32]byte)) {
		if gindex == 1 {
			out.WriteString(" --- START ")
			out.WriteString(name)
			out.WriteString(" --- \n")
			fn(root)
			out.WriteString(" --- END ")
			out.WriteString(name)
			out.WriteString(" --- \n")
			return
		}
		ab, ok := lookup(root)
		if !ok {
			out.WriteString(fmt.Sprintf("no %s (%b): %x\n", name, gindex, root))
			return
		}
		l := bits.Len64(gindex)
		x := uint64(1) << (l - 1)
		if gindex&(x>>1) != 0 {
			vizMaybe(name, (gindex^x)|(x>>1), ab[1], fn)
		} else {
			vizMaybe(name, (gindex^x)|(x>>1), ab[0], fn)
		}
	}
	vizMaybe("pc", 8, stateRoot, vizUint64)
	vizMaybe("memory", 9, stateRoot, vizRootMem)
	//vizMaybe("registers", 10, stateRoot, vizRootRegs)
	vizMaybe("heap", 13, stateRoot, vizUint64)
	return out.String()
}

func (s *StateOracle) Diff(a [32]byte, b [32]byte, gindex uint64) {
	if a == b {
		return
	}
	vA, okA := s.data[a]
	vB, okB := s.data[b]
	if okA {
		if okB {
			s.Diff(vA[0], vB[0], gindex<<1)
			s.Diff(vA[1], vB[1], (gindex<<1)|1)
		} else {
			fmt.Printf("%b: a = (%x, %x), b missing\n", gindex, vA[0], vA[1])
		}
	} else {
		if okB {
			fmt.Printf("%b: b = (%x, %x), a missing\n", gindex, vB[0], vB[1])
		} else {
			if a != b {
				fmt.Printf("%b: different:\n  a = %x\n  b = %x\n", gindex, a, b)
			}
		}
	}
}

var _ VMStateOracle = (*StateOracle)(nil)

func NewStateOracle() *StateOracle {
	return &StateOracle{
		data:            make(map[[32]byte][2][32]byte),
		reverse:         make(map[[2][32]byte][32]byte),
		buildAccessList: false,
	}
}

func (s *StateOracle) BuildAccessList(build bool) {
	s.buildAccessList = build
	s.accessList = [][32]byte{}
}

func (s *StateOracle) Get(key [32]byte) (a, b [32]byte) {
	if s.buildAccessList {
		s.accessList = append(s.accessList, key)
	}
	ab, ok := s.data[key]
	if !ok {
		panic(fmt.Errorf("missing key %x", key))
	}
	return ab[0], ab[1]
}

func (s *StateOracle) Remember(left [32]byte, right [32]byte) [32]byte {
	// cache is faster than hashing again
	value := [2][32]byte{left, right}
	if key, ok := s.reverse[value]; ok {
		//fmt.Printf("%x %x -> %x\n", left[:], right[:], key[:])
		return key
	}
	key := crypto.Keccak256Hash(left[:], right[:])
	//fmt.Printf("%x %x -> %x\n", left[:], right[:], key[:])
	s.data[key] = value
	s.reverse[value] = key
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
