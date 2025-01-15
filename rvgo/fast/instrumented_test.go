package fast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type MockPreimageOracle struct {
}

func (oracle *MockPreimageOracle) Hint(v []byte) {
}

func (oracle *MockPreimageOracle) GetPreimage(k [32]byte) []byte {
	return make([]byte, 32)
}

func (oracle *MockPreimageOracle) ReadPreimagePart(key [32]byte, offset uint64) ([32]byte, uint8, error) {
	return [32]byte{}, 32, nil
}

func TestReadPreimage(t *testing.T) {

	vmState := VMState{
		PC:        0,
		Memory:    NewMemory(),
		Registers: [32]uint64{},
		ExitCode:  0,
		Exited:    false,
		Heap:      0x7f_00_00_00_00_00,
	}

	// instruction ecall
	vmState.Memory.SetUnaligned(0, []byte{0x73})
	vmState.Registers[17] = 63
	vmState.Registers[10] = 5

	instState := NewInstrumentedState(&vmState, &MockPreimageOracle{}, nil, nil)

	_, err := instState.Step(true)
	require.NoError(t, err)
}
