package fast

import (
	"encoding/binary"
)

// page size must be at least 32 bytes (one merkle node)
// memory merkleization will look the same regardless of page size past 32.
const (
	pageAddrSize = 10
	pageKeySize  = 64 - pageAddrSize
	pageSize     = 1 << pageAddrSize
	pageAddrMask = pageSize - 1
	maxPageCount = 1 << pageKeySize
)

type VMState struct {
	Memory *Memory `json:"memory"`

	PreimageKey    [32]byte `json:"preimageKey"`
	PreimageOffset uint64   `json:"preimageOffset"`

	PC uint64 `json:"pc"`

	//0xF14: mhartid  - riscv tests use this. Always hart 0, no parallelism supported
	//CSR [4096]uint64 // 12 bit addressing space

	ExitCode uint8 `json:"exit"`
	Exited   bool  `json:"exited"`

	Step uint64 `json:"step"`

	Heap uint64 `json:"heap"` // for mmap to keep allocating new anon memory

	LoadReservation uint64 `json:"loadReservation"`

	Registers [32]uint64 `json:"registers"`
}

func NewVMState() *VMState {
	return &VMState{
		Memory: NewMemory(),
		Heap:   1 << 28, // 0.25 GiB of program code space
	}
}

func (state *VMState) EncodeWitness() []byte {
	out := make([]byte, 0)
	memRoot := state.Memory.MerkleRoot()
	out = append(out, memRoot[:]...)
	out = append(out, state.PreimageKey[:]...)
	out = binary.BigEndian.AppendUint64(out, state.PreimageOffset)
	out = binary.BigEndian.AppendUint64(out, state.PC)
	out = append(out, state.ExitCode)
	if state.Exited {
		out = append(out, 1)
	} else {
		out = append(out, 0)
	}
	out = binary.BigEndian.AppendUint64(out, state.Step)
	out = binary.BigEndian.AppendUint64(out, state.Heap)
	out = binary.BigEndian.AppendUint64(out, state.LoadReservation)
	for _, r := range state.Registers {
		out = binary.BigEndian.AppendUint64(out, r)
	}
	return out
}

func (state *VMState) Instr() uint32 {
	var out [4]byte
	state.Memory.GetUnaligned(state.PC, out[:])
	return binary.LittleEndian.Uint32(out[:])
}
