package fast

import (
	"encoding/binary"
	"fmt"

	"github.com/protolambda/asterisc/rvgo/oracle"
)

// page size must be at least 32 bytes (one merkle node)
// memory merkleization will look the same regardless of page size past 32.
const (
	pageAddrSize = 10
	pageSize     = 1 << pageAddrSize
	pageAddrMask = pageSize - 1
	maxPageCount = 1 << (64 - pageAddrSize)
)

type VMState struct {
	PC uint64
	// sparse memory, pages of 1KB, keyed by page number: start memory address truncated by 10 bits.
	Memory    map[uint64]*[pageSize]byte
	Registers [32]uint64

	// 0xF14: mhartid  - riscv tests use this. Always hart 0, no parallelism supported
	CSR [4096]uint64 // 12 bit addressing space

	Exit   uint64
	Exited bool
	Heap   uint64 // for mmap to keep allocating new anon memory
}

func NewVMState() *VMState {
	return &VMState{
		Memory: make(map[uint64]*[1024]byte),
		Heap:   1 << 28, // 0.25 GiB of program code space
	}
}

func (state *VMState) Merkleize(so oracle.VMStateOracle) [32]byte {
	var zeroHashes [256][32]byte
	for i := 1; i < 256; i++ {
		zeroHashes[i] = so.Remember(zeroHashes[i-1], zeroHashes[i-1])
	}
	pageBranches := make(map[uint64]struct{})
	for pageKey := range state.Memory {
		for i := 0; i < 64-pageAddrSize; i++ {
			gindex := (1 << (64 - pageAddrSize - i)) | (pageKey >> i)
			pageBranches[gindex] = struct{}{}
		}
	}
	merkleize := func(stackDepth uint64, getItem func(index uint64) [32]byte) [32]byte {
		stack := make([][32]byte, stackDepth+1)
		for i := uint64(0); i < pageSize/32; i += 1 {
			for j := uint64(0); j < pageAddrSize; j++ {
				if i&(1<<j) == 0 {
					stack[j] = getItem(i)
					break
				} else {
					stack[j+1] = so.Remember(stack[j], getItem(i))
				}
			}
		}
		return stack[stackDepth]
	}
	uint64AsBytes32 := func(v uint64) (out [32]byte) {
		binary.LittleEndian.PutUint64(out[:8], v)
		return
	}
	merkleizePage := func(page *[1024]byte) [32]byte {
		return merkleize(pageAddrSize, func(index uint64) [32]byte {
			return *(*[32]byte)(page[index*32 : index*32+32])
		})
	}
	var merkleizeMemory func(gindex uint64, depth uint64) [32]byte
	merkleizeMemory = func(gindex uint64, depth uint64) [32]byte {
		if depth == 64-pageAddrSize {
			pageKey := gindex & ((1 << (64 - pageAddrSize)) - 1)
			return merkleizePage(state.Memory[pageKey])
		}
		left := gindex << 1
		right := left | 1
		var leftRoot, rightRoot [32]byte
		if _, ok := pageBranches[left]; ok {
			leftRoot = merkleizeMemory(left, depth+1)
		} else {
			leftRoot = zeroHashes[64-pageAddrSize-depth]
		}
		if _, ok := pageBranches[right]; ok {
			rightRoot = merkleizeMemory(right, depth+1)
		} else {
			rightRoot = zeroHashes[64-pageAddrSize-depth]
		}
		return so.Remember(leftRoot, rightRoot)
	}

	memoryRoot := merkleizeMemory(1, 0)
	registersRoot := merkleize(5, func(index uint64) [32]byte {
		return uint64AsBytes32(state.Registers[index])
	})
	csrRoot := merkleize(12, func(index uint64) [32]byte {
		return uint64AsBytes32(state.CSR[index])
	})
	return so.Remember(
		so.Remember(
			so.Remember(uint64AsBytes32(state.PC), memoryRoot), // 8, 9
			so.Remember(registersRoot, csrRoot),                // 10, 11
		),
		so.Remember(
			so.Remember(uint64AsBytes32(state.Exit), uint64AsBytes32(state.Heap)), // 12, 13
			zeroHashes[1], // 14, 15
		),
	)
}

func (state *VMState) loadOrCreatePage(pageIndex uint64) *[pageSize]byte {
	if pageIndex >= maxPageCount {
		panic("invalid page key")
	}
	p, ok := state.Memory[pageIndex]
	if ok {
		return p
	}
	p = &[pageSize]byte{}
	state.Memory[pageIndex] = p
	return p
}

func (state *VMState) loadMem(addr uint64, size uint64, signed bool) uint64 {
	if size > 8 {
		panic(fmt.Errorf("cannot load more than 8 bytes: %d", size))
	}
	var out [8]byte
	pageIndex := addr >> pageAddrSize
	if _, ok := state.Memory[pageIndex]; !ok { // if page does not exist, then it's a 0 value
		return 0
	}
	copy(out[:], state.Memory[pageIndex][addr&pageAddrMask:])
	end := addr + size - 1 // Can also wrap around total memory.
	endPage := end >> pageAddrSize
	if pageIndex != endPage { // if it spans across two pages.
		if _, ok := state.Memory[endPage]; ok { // only if page exists, 0 otherwise
			remaining := (end & pageAddrMask) + 1
			copy(out[size-remaining:], state.Memory[endPage][:remaining])
		}
	}
	v := binary.LittleEndian.Uint64(out[:]) & ((1 << (size * 8)) - 1)
	if signed && v&((1<<(size<<3))>>1) != 0 { // if the last bit is set, then extend it to the full 64 bits
		v |= 0xFFFF_FFFF_FFFF_FFFF << (size << 3)
	} // otherwise just leave it zeroed
	return v
}

func (state *VMState) storeMem(addr uint64, size uint64, value uint64) {
	if size > 8 {
		panic(fmt.Errorf("cannot store more than 8 bytes: %d", size))
	}
	var bytez [8]byte
	binary.LittleEndian.PutUint64(bytez[:], value)
	pageIndex := addr >> pageAddrSize
	if _, ok := state.Memory[pageIndex]; !ok { // create page if it does not exist
		state.Memory[pageIndex] = &[pageSize]byte{}
	}
	copy(state.Memory[pageIndex][addr&pageAddrMask:], bytez[:size])
	end := addr + size - 1
	endPage := end >> pageAddrSize
	if pageIndex != endPage { // if it spans across two pages. Can also wrap around total memory.
		if _, ok := state.Memory[endPage]; !ok { // create page if it does not exist
			state.Memory[endPage] = &[pageSize]byte{}
		}
		remaining := (end & pageAddrMask) + 1
		copy(state.Memory[endPage][:remaining], bytez[size-remaining:])
	}
}

func (state *VMState) writeRegister(reg uint64, v uint64) {
	fmt.Printf("rd write to %d value: 0x%x\n", reg, v)
	if reg == 0 { // reg 0 must stay 0
		// v is a HINT, but no hints are specified by standard spec, or used by us.
		return
	}
	if reg >= 32 {
		panic(fmt.Errorf("unknown register %d, cannot write %x", reg, v))
	}
	state.Registers[reg] = v
}
