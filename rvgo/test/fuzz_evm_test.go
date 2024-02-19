package test

import (
	"encoding/binary"
	"os"
	"testing"

	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/asterisc/rvgo/slow"
)

var syscallInsn = uint64(0x73_00_00_00_00_00_00_00)

func staticOracle(t *testing.T, preimageData []byte) *testOracle {
	return &testOracle{
		hint: func(v []byte) {},
		getPreimage: func(k [32]byte) []byte {
			if k != preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey() {
				t.Fatalf("invalid preimage request for %x", k)
			}
			return preimageData
		},
	}
}

func FuzzStateHintRead(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64) {
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		pc := uint64(0)
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: fast.SysRead, 10: fast.FdHintRead, 11: addr, 12: count},
			Step:            0,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, syscallInsn)
		state.Memory.SetUnaligned(pc, buf)
		preStatePreimageKey := state.PreimageKey
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = count
		expectedRegisters[11] = 0

		oracle := staticOracle(t, preimageData)

		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, uint64(4), state.PC)
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, uint64(1), state.Step)
		require.Equal(t, preStatePreimageKey, state.PreimageKey)
		require.Equal(t, expectedRegisters, state.Registers)

		env := newEVMEnv(t, contracts, addrs)
		evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0)
		fastPost := state.EncodeWitness()
		require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
			"fast VM produced different state than EVM")

		slowPostHash, err := slow.Step(stepWitness.EncodeStepInput(fast.LocalContext{}), oracle)
		require.NoError(t, err)
		fastPostHash, err := fastPost.StateHash()
		require.NoError(t, err)
		require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
	})
}

func FuzzStatePreimageRead(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64) {
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		pc := uint64(0)
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: fast.SysRead, 10: fast.FdPreimageRead, 11: addr, 12: count},
			Step:            0,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, syscallInsn)
		state.Memory.SetUnaligned(pc, buf)
		preStatePreimageKey := state.PreimageKey
		preStateRoot := state.Memory.MerkleRoot()
		writeLen := count
		if writeLen > 4 {
			writeLen = 4
		}
		if preimageOffset+writeLen > uint64(8+len(preimageData)) {
			writeLen = uint64(8+len(preimageData)) - preimageOffset
		}
		oracle := staticOracle(t, preimageData)

		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.True(t, stepWitness.HasPreimage())

		require.Equal(t, uint64(4), state.PC)
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		if writeLen > 0 {
			// Memory may be unchanged if we're writing the first zero-valued 7 bytes of the pre-image.
			//require.NotEqual(t, preStateRoot, state.Memory.MerkleRoot())
			require.Greater(t, state.PreimageOffset, preimageOffset)
		} else {
			require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
			require.Equal(t, state.PreimageOffset, preimageOffset)
		}
		require.Equal(t, uint64(1), state.Step)
		require.Equal(t, preStatePreimageKey, state.PreimageKey)

		env := newEVMEnv(t, contracts, addrs)
		evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0)
		fastPost := state.EncodeWitness()
		require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
			"fast VM produced different state than EVM")

		slowPostHash, err := slow.Step(stepWitness.EncodeStepInput(fast.LocalContext{}), oracle)
		require.NoError(t, err)
		fastPostHash, err := fastPost.StateHash()
		require.NoError(t, err)
		require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
	})
}

func FuzzStateHintWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64) {
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		pc := uint64(0)
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: fast.SysWrite, 10: fast.FdHintWrite, 11: addr, 12: count},
			Step:            0,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,

			// This is only used by mips.go. The reads a zeroed page-sized buffer when reading hint data from memory.
			// We pre-allocate a buffer for the read hint data to be copied into.
			LastHint: make(hexutil.Bytes, fast.PageSize),
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, syscallInsn)
		state.Memory.SetUnaligned(pc, buf)
		preStatePreimageKey := state.PreimageKey
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = count
		expectedRegisters[11] = 0

		oracle := staticOracle(t, preimageData)

		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, uint64(4), state.PC)
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, uint64(1), state.Step)
		require.Equal(t, preStatePreimageKey, state.PreimageKey)
		require.Equal(t, expectedRegisters, state.Registers)

		env := newEVMEnv(t, contracts, addrs)
		evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0)
		fastPost := state.EncodeWitness()
		require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
			"fast VM produced different state than EVM")

		slowPostHash, err := slow.Step(stepWitness.EncodeStepInput(fast.LocalContext{}), oracle)
		require.NoError(t, err)
		fastPostHash, err := fastPost.StateHash()
		require.NoError(t, err)
		require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
	})
}

func FuzzStatePreimageWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64) {
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		pc := uint64(0)
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: fast.SysWrite, 10: fast.FdPreimageWrite, 11: addr, 12: count},
			Step:            0,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,
		}
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, syscallInsn)
		state.Memory.SetUnaligned(pc, buf)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		maxData := 32 - (addr & 31)
		if maxData < count {
			count = maxData
		}
		expectedRegisters[10] = count
		expectedRegisters[11] = 0

		oracle := staticOracle(t, preimageData)

		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, uint64(4), state.PC)
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, uint64(1), state.Step)
		require.Equal(t, uint64(0), state.PreimageOffset)
		require.Equal(t, expectedRegisters, state.Registers)

		env := newEVMEnv(t, contracts, addrs)
		evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0)
		fastPost := state.EncodeWitness()
		require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
			"fast VM produced different state than EVM")

		slowPostHash, err := slow.Step(stepWitness.EncodeStepInput(fast.LocalContext{}), oracle)
		require.NoError(t, err)
		fastPostHash, err := fastPost.StateHash()
		require.NoError(t, err)
		require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
	})
}
