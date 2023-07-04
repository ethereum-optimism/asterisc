package fast

import (
	"encoding/binary"
	"fmt"
	"io"
)

type PreimageOracle interface {
	Hint(v []byte) error
	GetPreimage(k [32]byte) ([]byte, error)
}

type InstrumentedState struct {
	state *VMState

	stdOut io.Writer
	stdErr io.Writer

	lastMemAccess   uint64
	memProofEnabled bool
	memProof        [(64 - 5 + 1) * 32]byte

	preimageOracle PreimageOracle

	// cached pre-image data, including 8 byte length prefix
	lastPreimage []byte
	// key for above preimage
	lastPreimageKey [32]byte
	// offset we last read from, or max uint64 if nothing is read this step
	lastPreimageOffset uint64
}

func NewInstrumentedState(state *VMState, po PreimageOracle, stdOut, stdErr io.Writer) *InstrumentedState {
	return &InstrumentedState{
		state:          state,
		stdOut:         stdOut,
		stdErr:         stdErr,
		preimageOracle: po,
	}
}

func (m *InstrumentedState) Step(proof bool) (wit *StepWitness, err error) {
	m.memProofEnabled = proof
	m.lastMemAccess = ^uint64(0)
	m.lastPreimageOffset = ^uint64(0)

	if proof {
		insnProof := m.state.Memory.MerkleProof(m.state.PC)
		wit = &StepWitness{
			State:    m.state.EncodeWitness(),
			MemProof: insnProof[:],
		}
	}
	err = m.riscvStep()
	if err != nil {
		return nil, err
	}

	if proof {
		wit.MemProof = append(wit.MemProof, m.memProof[:]...)
		if m.lastPreimageOffset != ^uint64(0) {
			wit.PreimageOffset = m.lastPreimageOffset
			wit.PreimageKey = m.lastPreimageKey
			wit.PreimageValue = m.lastPreimage
		}
	}
	return
}

func (m *InstrumentedState) readPreimage(key [32]byte, offset uint64) (dat [32]byte, datLen uint64, err error) {
	preimage := m.lastPreimage
	if key != m.lastPreimageKey {
		m.lastPreimageKey = key
		data, err := m.preimageOracle.GetPreimage(key)
		if err != nil {
			return [32]byte{}, 0, fmt.Errorf("failed to read preimage %x at offset %d: %w", key, offset, err)
		}
		// add the length prefix
		preimage = make([]byte, 0, 8+len(data))
		preimage = binary.BigEndian.AppendUint64(preimage, uint64(len(data)))
		preimage = append(preimage, data...)
		m.lastPreimage = preimage
	}
	m.lastPreimageOffset = offset
	datLen = uint64(copy(dat[:], preimage[offset:]))
	return
}

func (m *InstrumentedState) trackMemAccess(effAddr uint64) {
	if effAddr&31 != 0 {
		panic("effective memory access must be aligned to 32 bytes")
	}
	if m.memProofEnabled && m.lastMemAccess != effAddr {
		if m.lastMemAccess != ^uint64(0) {
			panic(fmt.Errorf("unexpected different mem access at %08x, already have access at %08x buffered", effAddr, m.lastMemAccess))
		}
		m.lastMemAccess = effAddr
		m.memProof = m.state.Memory.MerkleProof(effAddr)
	}
}
