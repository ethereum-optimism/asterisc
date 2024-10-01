package fast

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSerializeStateRoundTrip(t *testing.T) {
	// Construct a test case with populated fields
	mem := NewMemory()
	mem.AllocPage(5)
	p := mem.AllocPage(123)
	p.Data[2] = 0x01
	state := &VMState{
		Memory:          mem,
		PreimageKey:     common.Hash{0xFF},
		PreimageOffset:  5,
		PC:              6,
		ExitCode:        1,
		Exited:          true,
		Step:            0xdeadbeef,
		Heap:            0xc0ffee,
		LoadReservation: 40,
		Registers: [32]uint64{
			0xdeadbeef,
			0xdeadbeef,
			0xc0ffee,
			0xbeefbabe,
			0xdeadc0de,
			0xbadc0de,
			0xdeaddead,
		},
		LastHint:  hexutil.Bytes{1, 2, 3, 4, 5},
		Witness:   hexutil.Bytes{6, 7, 8, 9, 10},
		StateHash: common.Hash{0x12},
	}

	ser := new(bytes.Buffer)
	err := state.Serialize(ser)
	require.NoError(t, err, "must serialize state")
	state2 := &VMState{}
	err = state2.Deserialize(ser)
	require.NoError(t, err, "must deserialize state")
	require.Equal(t, state, state2, "must roundtrip state")
}
