package fast

import (
	"encoding/binary"
	"fmt"
	"io"
)

type PreimageOracle interface {
	Hint(v []byte)
	GetPreimage(k [32]byte) []byte
}

const memProofSize = (64 - 5 + 1) * 32

type InstrumentedState struct {
	state *VMState

	stdOut io.Writer
	stdErr io.Writer

	memProofEnabled bool
	memProofs       [][memProofSize]byte
	memAccess       []uint64

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
	m.memAccess = m.memAccess[:0]
	m.memProofs = m.memProofs[:0]
	m.lastPreimageOffset = ^uint64(0)

	if proof {
		wit = &StepWitness{
			State: m.state.EncodeWitness(), // we need the pre-state as wit-ness
		}
	}

	err = m.riscvStep()

	if proof {
		wit.MemProof = make([]byte, 0, len(m.memProofs)*memProofSize)
		for i := range m.memProofs {
			wit.MemProof = append(wit.MemProof, m.memProofs[i][:]...)
		}
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
	if preimage == nil || key != m.lastPreimageKey {
		m.lastPreimageKey = key
		data := m.preimageOracle.GetPreimage(key)
		// add the length prefix
		preimage = make([]byte, 0, 8+len(data))
		preimage = binary.BigEndian.AppendUint64(preimage, uint64(len(data)))
		preimage = append(preimage, data...)
		m.lastPreimage = preimage
	}
	m.lastPreimageOffset = offset
	if offset >= uint64(len(preimage)) {
		panic("Preimage offset out-of-bounds")
	}
	datLen = uint64(copy(dat[:], preimage[offset:]))
	return
}

// trackMemAccess remembers a merkle-branch of memory to the given address,
// and ensures it comes right after the last memory proof.
func (m *InstrumentedState) trackMemAccess(effAddr uint64, proofIndex uint8) {
	if !m.memProofEnabled {
		return
	}
	if effAddr&31 != 0 {
		panic("effective memory access must be aligned to 32 bytes")
	}
	if len(m.memProofs) != int(proofIndex) {
		panic(fmt.Errorf("mem access with unexpected proof index, got %d but expected %d", proofIndex, len(m.memProofs)))
	}
	m.memProofs = append(m.memProofs, m.state.Memory.MerkleProof(effAddr))
	m.memAccess = append(m.memAccess, effAddr)
}

// verifyMemChange verifies a memory change proof reused the last verified mem-proof at the same address
func (m *InstrumentedState) verifyMemChange(effAddr uint64, proofIndex uint8) {
	if !m.memProofEnabled {
		return
	}
	if int(proofIndex) >= len(m.memAccess) {
		panic(fmt.Errorf("mem change at %016x with proof index %d, but only aware of %d proofs", effAddr, proofIndex, len(m.memAccess)))
	}
	if effAddr != m.memAccess[proofIndex] {
		panic(fmt.Errorf("mem access at %016x with mismatching prior proof verification for address %016x", effAddr, m.memAccess[proofIndex]))
	}
}

func (m *InstrumentedState) LastPreimage() ([32]byte, []byte, uint64) {
	return m.lastPreimageKey, m.lastPreimage, m.lastPreimageOffset
}
