package fast

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

	// LastHint is optional metadata, and not part of the VM state itself.
	// It is used to remember the last pre-image hint,
	// so a VM can start from any state without fetching prior pre-images,
	// and instead just repeat the last hint on setup,
	// to make sure pre-image requests can be served.
	// The first 4 bytes are a uin32 length prefix.
	// Warning: the hint MAY NOT BE COMPLETE. I.e. this is buffered,
	// and should only be read when len(LastHint) > 4 && uint32(LastHint[:4]) >= len(LastHint[4:])
	LastHint hexutil.Bytes `json:"lastHint,omitempty"`
}

func NewVMState() *VMState {
	return &VMState{
		Memory: NewMemory(),
		Heap:   1 << 28, // 0.25 GiB of program code space
	}
}

func (state *VMState) GetStep() uint64 { return state.Step }

func (state *VMState) EncodeWitness() StateWitness {
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

type StateWitness []byte

const (
	VMStatusValid      = 0
	VMStatusInvalid    = 1
	VMStatusPanic      = 2
	VMStatusUnfinished = 3
)

func (sw StateWitness) StateHash() (common.Hash, error) {
	offset := 32 + 32 + 8 + 8 // mem-root, preimage-key, preimage-offset, PC
	if len(sw) <= offset+1 {
		return common.Hash{}, fmt.Errorf("state must at least be %d bytes, but got %d", offset, len(sw))
	}

	hash := crypto.Keccak256Hash(sw)
	exitCode := sw[offset]
	exited := sw[offset+1]
	status := vmStatus(exited == 1, exitCode)
	hash[0] = status
	return hash, nil
}

func vmStatus(exited bool, exitCode uint8) uint8 {
	if !exited {
		return VMStatusUnfinished
	}
	switch exitCode {
	case 0:
		return VMStatusValid
	case 1:
		return VMStatusInvalid
	default:
		return VMStatusPanic
	}
}
