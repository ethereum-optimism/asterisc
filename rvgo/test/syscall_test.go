package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"testing"

	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/asterisc/rvgo/riscv"
	"github.com/ethereum-optimism/asterisc/rvgo/slow"
)

var syscallInsn = []byte{0x73, 0x00, 0x00, 0x00}

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

type hintTrackingOracle struct {
	hints [][]byte
}

func (t *hintTrackingOracle) Hint(v []byte) {
	t.hints = append(t.hints, v)
}

func (t *hintTrackingOracle) GetPreimage(k [32]byte) []byte {
	return nil
}

func runEVM(t *testing.T, contracts *Contracts, addrs *Addresses, stepWitness *fast.StepWitness, fastPost fast.StateWitness, revertCode []byte) {
	env := newEVMEnv(t, contracts, addrs)
	evmPost, _, _ := stepEVM(t, env, stepWitness, addrs, 0, revertCode)
	require.Equal(t, hexutil.Bytes(fastPost).String(), hexutil.Bytes(evmPost).String(),
		"fast VM produced different state than EVM")
}

func runSlow(t *testing.T, stepWitness *fast.StepWitness, fastPost fast.StateWitness, po slow.PreimageOracle, expectedErr interface{}) {
	input, err := stepWitness.EncodeStepInput(fast.LocalContext{})
	require.NoError(t, err)
	slowPostHash, err := slow.Step(input, po)
	if expectedErr != nil {
		require.ErrorAs(t, err, expectedErr)
	} else {
		require.NoError(t, err)
		fastPostHash, err := fastPost.StateHash()
		require.NoError(t, err)
		require.Equal(t, fastPostHash, slowPostHash, "fast VM produced different state than slow VM")
	}

}

func errCodeToByte32(errCode uint64) []byte {
	return binary.BigEndian.AppendUint64(make([]byte, 24), errCode)
}

func TestStateSyscallUnsupported(t *testing.T) {
	contracts := testContracts(t)
	addrs := testAddrs
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
			stepWitness, err := fastState.Step(true)
			var fastSyscallErr *fast.UnsupportedSyscallErr
			require.ErrorAs(t, err, &fastSyscallErr)

			runEVM(t, contracts, addrs, stepWitness, nil, errCodeToByte32(riscv.ErrInvalidSyscall))

			var slowSyscallErr *slow.UnsupportedSyscallErr
			runSlow(t, stepWitness, nil, nil, &slowSyscallErr)
		})
	}
}

func TestEVMSysWriteHint(t *testing.T) {
	contracts := testContracts(t)
	addrs := testAddrs

	cases := []struct {
		name          string
		memOffset     int      // Where the hint data is stored in memory
		hintData      []byte   // Hint data stored in memory at memOffset
		bytesToWrite  int      // How many bytes of hintData to write
		lastHint      []byte   // The buffer that stores lastHint in the state
		expectedHints [][]byte // The hints we expect to be processed
	}{
		{
			name:      "write 1 full hint at beginning of page",
			memOffset: 4096,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 10,
			lastHint:     nil,
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB},
			},
		},
		{
			name:      "write 1 full hint across page boundary",
			memOffset: 4092,
			hintData: []byte{
				0, 0, 0, 8, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 12,
			lastHint:     nil,
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB},
			},
		},
		{
			name:      "write 2 full hints",
			memOffset: 5012,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
				0, 0, 0, 8, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 22,
			lastHint:     nil,
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB},
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB},
			},
		},
		{
			name:      "write a single partial hint",
			memOffset: 4092,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite:  8,
			lastHint:      nil,
			expectedHints: nil,
		},
		{
			name:      "write 1 full, 1 partial hint",
			memOffset: 5012,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
				0, 0, 0, 8, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 16,
			lastHint:     nil,
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB},
			},
		},
		{
			name:      "write a single partial hint to large capacity lastHint buffer",
			memOffset: 4092,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite:  8,
			lastHint:      make([]byte, 0, 4096),
			expectedHints: nil,
		},
		{
			name:      "write full hint to large capacity lastHint buffer",
			memOffset: 5012,
			hintData: []byte{
				0, 0, 0, 6, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 10,
			lastHint:     make([]byte, 0, 4096),
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB},
			},
		},
		{
			name:      "write multiple hints to large capacity lastHint buffer",
			memOffset: 4092,
			hintData: []byte{
				0, 0, 0, 8, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, // Hint data
				0, 0, 0, 8, // Length prefix
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, // Hint data
			},
			bytesToWrite: 24,
			lastHint:     make([]byte, 0, 4096),
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC},
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB},
			},
		},
		{
			name:      "write remaining hint data to non-empty lastHint buffer",
			memOffset: 4092,
			hintData: []byte{
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, // Hint data
			},
			bytesToWrite: 8,
			lastHint:     []byte{0, 0, 0, 8},
			expectedHints: [][]byte{
				{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC},
			},
		},
		{
			name:      "write partial hint data to non-empty lastHint buffer",
			memOffset: 4092,
			hintData: []byte{
				0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xCC, 0xCC, // Hint data
			},
			bytesToWrite:  4,
			lastHint:      []byte{0, 0, 0, 8},
			expectedHints: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			oracle := hintTrackingOracle{}

			state := &fast.VMState{
				PC:        0,
				Memory:    fast.NewMemory(),
				Registers: [32]uint64{17: riscv.SysWrite, 10: riscv.FdHintWrite, 11: uint64(tt.memOffset), 12: uint64(tt.bytesToWrite)},
				LastHint:  tt.lastHint,
			}

			err := state.Memory.SetMemoryRange(uint64(tt.memOffset), bytes.NewReader(tt.hintData))
			require.NoError(t, err)
			state.Memory.SetUnaligned(0, syscallInsn)

			fastState := fast.NewInstrumentedState(state, &oracle, os.Stdout, os.Stderr)
			stepWitness, err := fastState.Step(true)
			require.NoError(t, err)
			require.Equal(t, tt.expectedHints, oracle.hints)

			fastPost := state.EncodeWitness()
			runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	f.Fuzz(func(t *testing.T, exitCode uint8, pc uint64, step uint64) {
		for _, syscall := range syscalls {
			testExit(t, syscall, exitCode, pc, step)
		}
	})
}

func FuzzStateSyscallBrk(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysBrk},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = 1 << 30
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	})
}

func FuzzStateSyscallMmap(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, isZeroAddr bool, addr uint64, length uint64, heap uint64, pc uint64, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		if isZeroAddr {
			addr = 0
		}
		state := &fast.VMState{
			PC:              pc,
			Heap:            heap,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers: [32]uint64{
				17: riscv.SysMmap,
				10: addr,
				11: length,
				13: 32,                    // MAP_ANONYMOUS flag
				14: 0xFFFF_FFFF_FFFF_FFFF, // fd == -1 (u64 mask)
			},
			Step: step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance

		newHeap := heap
		if addr == 0 {
			expectedRegisters[10] = heap
			align := length & fast.PageAddrMask
			if align != 0 {
				length = length + fast.PageSize - align
			}
			newHeap = heap + length
		}
		require.Equal(t, expectedRegisters, state.Registers)
		require.Equal(t, newHeap, state.Heap)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	})
}

func FuzzStateSyscallFcntl(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	testFcntl := func(t *testing.T, fd, cmd, pc, step, out, errCode uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysFcntl, 10: fd, 11: cmd},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = out
		expectedRegisters[11] = errCode

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	f.Fuzz(func(t *testing.T, fd uint64, cmd uint64, pc uint64, step uint64) {
		// Test F_GETFL for O_RDONLY fds
		for _, fd := range []uint64{0, 3, 5} {
			testFcntl(t, fd, 3, pc, step, 0, 0)
		}
		// Test F_GETFL for O_WRONLY fds
		for _, fd := range []uint64{1, 2, 4, 6} {
			testFcntl(t, fd, 3, pc, step, 1, 0)
		}
		// Test F_GETFL for unsupported fds
		// Add 7 to fd to ensure fd > 6
		testFcntl(t, fd+7, 3, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)

		// Test F_GETFD
		for _, fd := range []uint64{0, 1, 2, 3, 4, 5, 6} {
			testFcntl(t, fd, 1, pc, step, 0, 0)
		}

		// Test F_GETFD for unsupported fds
		// Add 7 to fd to ensure fd > 6
		testFcntl(t, fd+7, 1, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)

		// Test other commands
		if cmd == 3 || cmd == 1 {
			// Set arbitrary commands if cmd is F_GETFL
			cmd = 4
		}
		testFcntl(t, fd, cmd, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x16)
	})
}

func FuzzStateSyscallOpenat(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysOpenat},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = 0xFFFF_FFFF_FFFF_FFFF
		expectedRegisters[11] = 0xd

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	})
}

func FuzzStateSyscallClockGettime(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysClockGettime, 11: addr},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		expectedRegisters := state.Registers
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		postMemory := fast.NewMemory()
		postMemory.SetUnaligned(pc, syscallInsn)
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], 1337)
		postMemory.SetUnaligned(addr, bytes[:])
		postMemory.SetUnaligned(addr+8, []byte{42, 0, 0, 0, 0, 0, 0, 0})

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, state.Memory.MerkleRoot(), postMemory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	})
}

func FuzzStateSyscallClone(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysClone},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = 1
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	})
}

func FuzzStateSyscallGetrlimit(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	testGetrlimit := func(t *testing.T, addr, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysGetrlimit, 10: 7, 11: addr},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		expectedRegisters := state.Registers
		expectedRegisters[10] = 0
		expectedRegisters[11] = 0

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		postMemory := fast.NewMemory()
		postMemory.SetUnaligned(pc, syscallInsn)
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], 1024)
		postMemory.SetUnaligned(addr, bytes[:])
		postMemory.SetUnaligned(addr+8, bytes[:])

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, state.Memory.MerkleRoot(), postMemory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	testGetrlimitErr := func(t *testing.T, res, addr, pc, step uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		addr = addr &^ 31
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysGetrlimit, 10: res, 11: addr},
			Step:            0,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		var fastSyscallErr *fast.UnrecognizedResourceErr
		require.ErrorAs(t, err, &fastSyscallErr)

		runEVM(t, contracts, addrs, stepWitness, nil, errCodeToByte32(riscv.ErrUnrecognizedResource))

		var slowSyscallErr *slow.UnrecognizedResourceErr
		runSlow(t, stepWitness, nil, nil, &slowSyscallErr)
	}

	f.Fuzz(func(t *testing.T, res, addr, pc, step uint64) {
		// Test RLIMIT_NOFILE
		testGetrlimit(t, addr, pc, step)

		// Test other resources
		if res == 7 {
			// Set arbitrary resource if res is RLIMIT_NOFILE
			res = 8
		}
		testGetrlimitErr(t, res, addr, pc, step)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	f.Fuzz(func(t *testing.T, arg uint64, pc uint64, step uint64) {
		for _, syscall := range syscalls {
			testNoop(t, syscall, arg, pc, step)
		}
	})
}

func FuzzStateSyscallRead(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	testRead := func(t *testing.T, fd, addr, count, pc, step, ret, errCode uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysRead, 10: fd, 11: addr, 12: count},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = ret
		expectedRegisters[11] = errCode

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	f.Fuzz(func(t *testing.T, fd, addr, count, pc, step uint64) {
		// Test stdin
		testRead(t, riscv.FdStdin, addr, count, pc, step, 0, 0)

		// Test EBADF err
		if fd == riscv.FdStdin || fd == riscv.FdHintRead || fd == riscv.FdPreimageRead {
			// Ensure unsupported fd
			fd += 1
		}
		testRead(t, fd, addr, count, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, oracle, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, oracle, nil)
	})
}

func FuzzStateSyscallWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	testWrite := func(t *testing.T, fd, addr, count, pc, step, ret, errCode uint64) {
		pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
		state := &fast.VMState{
			PC:              pc,
			Heap:            0,
			ExitCode:        0,
			Exited:          false,
			Memory:          fast.NewMemory(),
			LoadReservation: 0,
			Registers:       [32]uint64{17: riscv.SysWrite, 10: fd, 11: addr, 12: count},
			Step:            step,
		}
		state.Memory.SetUnaligned(pc, syscallInsn)
		preStateRoot := state.Memory.MerkleRoot()
		expectedRegisters := state.Registers
		expectedRegisters[10] = ret
		expectedRegisters[11] = errCode

		fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
		stepWitness, err := fastState.Step(true)
		require.NoError(t, err)
		require.False(t, stepWitness.HasPreimage())

		require.Equal(t, pc+4, state.PC) // PC must advance
		require.Equal(t, uint64(0), state.Heap)
		require.Equal(t, uint64(0), state.LoadReservation)
		require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
		require.Equal(t, false, state.Exited)      // Must not be exited
		require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
		require.Equal(t, step+1, state.Step) // Step must advance
		require.Equal(t, expectedRegisters, state.Registers)

		fastPost := state.EncodeWitness()
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, nil, nil)
	}

	f.Fuzz(func(t *testing.T, fd, addr, count, pc, step uint64) {
		// Test stdout
		testWrite(t, riscv.FdStdout, addr, count, pc, step, count, 0)

		// Test stderr
		testWrite(t, riscv.FdStderr, addr, count, pc, step, count, 0)

		// Test EBADF err
		if fd == riscv.FdStdout || fd == riscv.FdStderr || fd == riscv.FdHintWrite || fd == riscv.FdPreimageWrite {
			// Ensure unsupported fd
			fd += 6
		}
		testWrite(t, fd, addr, count, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)
	})
}

func FuzzStateHintWrite(f *testing.F) {
	contracts := testContracts(f)
	addrs := testAddrs

	f.Fuzz(func(t *testing.T, addr uint64, count uint64, preimageOffset uint64, pc uint64, step uint64, randSeed int64) {
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

			LastHint: nil,
		}
		// Set random data at the target memory range
		randBytes, err := randomBytes(randSeed, count)
		require.NoError(t, err)
		err = state.Memory.SetMemoryRange(addr, bytes.NewReader(randBytes))
		require.NoError(t, err)

		// Set syscall instruction
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, oracle, nil)
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

		var expectedKey common.Hash
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
		runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
		runSlow(t, stepWitness, fastPost, oracle, nil)
	})
}

func randomBytes(seed int64, length uint64) ([]byte, error) {
	r := rand.New(rand.NewSource(seed))
	randBytes := make([]byte, length)
	if _, err := r.Read(randBytes); err != nil {
		return nil, err
	}
	return randBytes, nil
}
