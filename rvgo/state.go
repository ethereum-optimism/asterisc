package rvgo

import "encoding/binary"

type VMState struct {
	PC uint64
	// sparse memory, pages of 1KB, keyed by page number: start memory address truncated by 10 bits.
	Memory    map[uint64]*[1024]byte
	Registers [32]uint64
	CSR       [2048]uint64
	Exit      uint64
}

func NewVMState() *VMState {
	return &VMState{
		Memory: make(map[uint64]*[1024]byte),
	}
}

// tree:
// ```
//
//	         1
//	    2          3
//	 4    5     6     7
//	8 9 10 11 12 13 14 15
//
// ```
const (
	pcGindex        = 8
	memoryGindex    = 9
	registersGindex = 10
	csrGindex       = 11
	exitGindex      = 12
)

func (state *VMState) Merkleize(so VMStateOracle) [32]byte {
	var zeroHashes [256][32]byte
	for i := 1; i < 256; i++ {
		zeroHashes[i] = so.Remember(zeroHashes[i-1], zeroHashes[i-1])
	}
	pageBranches := make(map[uint64]struct{})
	for pageKey := range state.Memory {
		for i := 0; i < 64-10; i++ {
			gindex := (1 << (64 - 10 - i)) | (pageKey >> i)
			pageBranches[gindex] = struct{}{}
		}
	}
	merkleize := func(stackDepth uint64, getItem func(index uint64) [32]byte) [32]byte {
		stack := make([][32]byte, stackDepth+1)
		for i := uint64(0); i < 1024/32; i += 1 {
			for j := uint64(0); j < 10; j++ {
				if i&(1<<j) == 0 {
					stack[j] = getItem(i)
					break
				} else {
					stack[j+1] = so.Remember(stack[j], getItem(i))
				}
			}
		}
		return stack[10+1]
	}
	uint64AsBytes32 := func(v uint64) (out [32]byte) {
		binary.BigEndian.PutUint64(out[:8], v)
		return
	}
	merkleizePage := func(page *[1024]byte) [32]byte {
		return merkleize(10, func(index uint64) [32]byte {
			return *(*[32]byte)(page[index*32 : index*32+32])
		})
	}
	var merkleizeMemory func(gindex uint64, depth uint64) [32]byte
	merkleizeMemory = func(gindex uint64, depth uint64) [32]byte {
		if depth == 64-10 {
			pageKey := gindex & ((1 << (64 - 10)) - 1)
			return merkleizePage(state.Memory[pageKey])
		}
		left := gindex << 1
		right := left | 1
		var leftRoot, rightRoot [32]byte
		if _, ok := pageBranches[left]; ok {
			leftRoot = merkleizeMemory(left, depth+1)
		} else {
			leftRoot = zeroHashes[64-10-depth]
		}
		if _, ok := pageBranches[right]; ok {
			rightRoot = merkleizeMemory(right, depth+1)
		} else {
			rightRoot = zeroHashes[64-10-depth]
		}
		return so.Remember(leftRoot, rightRoot)
	}

	memoryRoot := merkleizeMemory(1, 0)
	registersRoot := merkleize(5, func(index uint64) [32]byte {
		return uint64AsBytes32(state.Registers[index])
	})
	csrRoot := merkleize(11, func(index uint64) [32]byte {
		return uint64AsBytes32(state.CSR[index])
	})
	return so.Remember(
		so.Remember(
			so.Remember(uint64AsBytes32(state.PC), memoryRoot), // 8, 9
			so.Remember(registersRoot, csrRoot),                // 10, 11
		),
		so.Remember(
			so.Remember(uint64AsBytes32(state.Exit), zeroHashes[0]), // 12, 13
			zeroHashes[1], // 14, 15
		),
	)
}
