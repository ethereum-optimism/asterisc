package test

import (
	"fmt"
	"os"
	"testing"

	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/asterisc/rvgo/riscv"
	"github.com/ethereum-optimism/asterisc/rvgo/slow"
)

var syscallInsn = []byte{0x73}

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

func runEVM(t *testing.T, contracts *Contracts, addrs *Addresses, stepWitness *fast.StepWitness, fastPost fast.StateWitness) {
	env := newEVMEnv(t, contracts, addrs)
	evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0)
	require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
		"fast VM produced different state than EVM")
}

func runSlow(t *testing.T, stepWitness *fast.StepWitness, fastPost fast.StateWitness, po slow.PreimageOracle) {
	slowPostHash, err := slow.Step(stepWitness.EncodeStepInput(fast.LocalContext{}), po)
	require.NoError(t, err)
	fastPostHash, err := fastPost.StateHash()
	require.NoError(t, err)
	require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
}

func TestStateSyscallUnsupported(t *testing.T) {
	syscalls := []int{
		riscv.SysPrlimit64,
		riscv.SysFutex,
		riscv.SysNanosleep,
	}

	for _, syscall := range syscalls {
		t.Run(fmt.Sprintf("sys_%d", syscall), func(t *testing.T) {
			pc := uint64(0)
			state := &fast.VMState{
				PC:              pc,
				Heap:            0,
				ExitCode:        0,
				Exited:          false,
				Memory:          fast.NewMemory(),
				LoadReservation: 0,
				Registers:       [32]uint64{17: uint64(syscall)},
				Step:            0,
			}
			state.Memory.SetUnaligned(pc, syscallInsn)

			fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
			_, err := fastState.Step(true)
			var syscallErr *fast.UnsupportedSyscallErr
			require.ErrorAs(t, err, &syscallErr)

			// TODO: Test EVM & slow VM
		})
	}
}

func FuzzStateSyscallExit(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	syscalls := []int{riscv.SysExit, riscv.SysExitGroup}

	testExit := func(t *testing.T, syscall int, exitCode uint8, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: uint64(syscall), 10: uint64(exitCode)},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		preStateRegisters := state.Registers

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, exitCode, state.ExitCode) // ExitCode must be set
		require.Equal(t, true, state.Exited)       // Must be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, preStateRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
	}

	f.Fuzz(func(t *testing.T, exitCode uint8, pc uint64, step uint64) {
		for _, syscall := range syscalls {
			testExit(t, syscall, exitCode, pc, step)
		}
	})
}

func FuzzStateSyscallNoop(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	syscalls := []int{
		riscv.SysSchedGetaffinity,
		riscv.SysSchedYield,
		riscv.SysRtSigprocmask,
		riscv.SysSigaltstack,
		riscv.SysGettid,
		riscv.SysRtSigaction,
		riscv.SysMadvise,
		riscv.SysEpollCreate1,
		riscv.SysEpollCtl,
		riscv.SysPipe2,
		riscv.SysReadlinnkat,
		riscv.SysNewfstatat,
		riscv.SysNewuname,
		riscv.SysMunmap,
		riscv.SysGetRandom,
	}

	testNoop := func(t *testing.T, syscall int, arg uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: uint64(syscall), 10: arg},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = 0
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
	}

	f.Fuzz(func(t *testing.T, arg uint64, pc uint64, step uint64) {
		for _, syscall := range syscalls {
			testNoop(t, syscall, arg, pc, step)
		}
	})
}

func FuzzStateHintRead(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysRead, 10: riscv.FdHintRead, 11: addr, 12: count},
			Step:            step,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
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

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, preStatePreimageKey, state.PreimageKey)
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, oracle)
	})
}

func FuzzStatePreimageRead(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysRead, 10: riscv.FdPreimageRead, 11: addr, 12: count},
			Step:            step,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStatePreimageKey := state.PreimageKey
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		writeLen := count
		maxData := 32 - addr&31
		if writeLen > maxData {
			writeLen = maxData
		}
		leftPreimageLen := uint64(8+len(preimageData)) - preimageOffset
		if writeLen > leftPreimageLen {
			writeLen = leftPreimageLen
		}
		expectedRegisters[10] = writeLen
		expectedRegisters[11] = 0

		oracle := staticOracle(t, preimageData)
		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.True(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
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
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, preStatePreimageKey, state.PreimageKey)
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, oracle)
	})
}

func FuzzStateHintWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysWrite, 10: riscv.FdHintWrite, 11: addr, 12: count},
			Step:            step,
			PreimageKey:     preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey(),
			PreimageOffset:  preimageOffset,

			// This is only used by fast/vm.go. The reads a zeroed page-sized buffer when reading hint data from memory.
			// We pre-allocate a buffer for the read hint data to be copied into.
			LastHint: make(hexutil.Bytes, fast.PageSize),
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
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

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, preStatePreimageKey, state.PreimageKey)
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, oracle)
	})
}

func FuzzStatePreimageWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, preimageOffset uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		preimageData := []byte("hello world")
		if preimageOffset >= uint64(len(preimageData)) {
			t.SkipNow()
		}
		heap := uint64(0x7f_00_00_00_00_00)
		if addr < heap {
			// to avoid override code space
			addr = heap + addr%(0xff_ff_ff_ff_ff_ff_ff_ff-heap)
		}
		count := uint64(32) // preimage key is 32 bytes
		state := &fast.VMState{
			PC:              pc,
			Heap:            heap,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysWrite, 10: riscv.FdPreimageWrite, 11: addr, 12: count},
			Step:            step,
			PreimageOffset:  preimageOffset,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)

		// Set preimage key to addr
		preimageKey := preimage.Keccak256Key(crypto.Keccak256Hash(preimageData)).PreimageKey()
		state.Memory.SetUnaligned(addr, preimageKey[:])
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers

		maxData := 32 - (addr & 31)
		if maxData < count {
			count = maxData
		}
		expectedRegisters[10] = count
		expectedRegisters[11] = 0

		var expectedKey [32]byte
		// slice preimage key by count
		for i := uint64(0); i < count; i++ {
			expectedKey[i+32-count] = preimageKey[i]
		}

		oracle := staticOracle(t, preimageData)

		fastState := fast.NewInstrumentedState(state, oracle, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, heap, state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode)
		require.Equal(t, false, state.Exited)
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, uint64(0), state.PreimageOffset)
		require.Equal(t, expectedRegisters, state.Registers)
		require.Equal(t, expectedKey, state.PreimageKey)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, oracle)
	})
}