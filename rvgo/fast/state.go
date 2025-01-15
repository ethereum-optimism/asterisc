package fast

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum-optimism/optimism/op-service/serialize"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type VMState struct {
	Memory *Memory `json:"memory"`

	PreimageKey    common.Hash `json:"preimageKey"`
	PreimageOffset uint64      `json:"preimageOffset"`

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
	// and should only be read when len(LastHint) > 4 && uint32(LastHint[:4]) <= len(LastHint[4:])
	LastHint hexutil.Bytes `json:"lastHint,omitempty"`

	// VMState must hold these values because if not, we must ask FPVM again to
	// compute these values.
	Witness   []byte      `json:"witness,omitempty"`
	StateHash common.Hash `json:"stateHash,omitempty"`
}

func NewVMState() *VMState {
	return &VMState{
		Memory: NewMemory(),
		Heap:   1 << 28, // 0.25 GiB of program code space
	}
}

func (state *VMState) SetWitnessAndStateHash() error {
	witness := state.EncodeWitness()
	state.Witness = witness
	stateHash, err := witness.StateHash()
	if err != nil {
		return fmt.Errorf("failed to compute stateHash: %w", err)
	}
	state.StateHash = stateHash
	return nil
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

// Serialize writes the state in a simple binary format which can be read again using Deserialize
// The format is a simple concatenation of fields, with prefixed item count for repeating items and using big endian
// encoding for numbers.
//
// Memory                      As per Memory.Serialize
// PreimageKey                 [32]byte
// PreimageOffset              uint64
// PC						   uint64
// ExitCode                    uint8
// Exited                      bool - 0 for false, 1 for true
// Step                        uint64
// Heap                        uint64
// LoadReservation			   uint64
// Registers. 				   [32]uint64
// len(LastHint)			   uint64 (0 when LastHint is nil)
// LastHint 				   []byte
// len(Witness)				   uint64 (0 when Witness is nil)
// Witness					   []byte
// StateHash				   [32]byte
func (s *VMState) Serialize(out io.Writer) error {
	bout := serialize.NewBinaryWriter(out)

	if err := s.Memory.Serialize(out); err != nil {
		return err
	}
	if err := bout.WriteHash(s.PreimageKey); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.PreimageOffset); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.PC); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.ExitCode); err != nil {
		return err
	}
	if err := bout.WriteBool(s.Exited); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.Step); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.Heap); err != nil {
		return err
	}
	if err := bout.WriteUInt(s.LoadReservation); err != nil {
		return err
	}
	for _, r := range s.Registers {
		if err := bout.WriteUInt(r); err != nil {
			return err
		}
	}
	if err := bout.WriteBytes(s.LastHint); err != nil {
		return err
	}
	if err := bout.WriteBytes(s.Witness); err != nil {
		return err
	}
	if err := bout.WriteHash(s.StateHash); err != nil {
		return err
	}

	return nil
}

func (s *VMState) Deserialize(in io.Reader) error {
	bin := serialize.NewBinaryReader(in)
	s.Memory = NewMemory()
	if err := s.Memory.Deserialize(in); err != nil {
		return err
	}
	if err := bin.ReadHash(&s.PreimageKey); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.PreimageOffset); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.PC); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.ExitCode); err != nil {
		return err
	}
	if err := bin.ReadBool(&s.Exited); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.Step); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.Heap); err != nil {
		return err
	}
	if err := bin.ReadUInt(&s.LoadReservation); err != nil {
		return err
	}
	for i := range s.Registers {
		if err := bin.ReadUInt(&s.Registers[i]); err != nil {
			return err
		}
	}
	if err := bin.ReadBytes((*[]byte)(&s.LastHint)); err != nil {
		return err
	}
	if err := bin.ReadBytes((*[]byte)(&s.Witness)); err != nil {
		return err
	}
	if err := bin.ReadHash(&s.StateHash); err != nil {
		return err
	}

	return nil
}

func LoadVMStateFromFile(path string) (*VMState, error) {
	if !serialize.IsBinaryFile(path) {
		return jsonutil.LoadJSON[VMState](path)
	}
	return serialize.LoadSerializedBinary[VMState](path)
}
