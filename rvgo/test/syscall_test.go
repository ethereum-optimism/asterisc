package test

import (
	"encoding/binary"
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
			Registers:       [32]uint64{17: riscv.SysMmap, 10: addr, 11: length},
			Step:            step,
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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

		// Test other commands
		if cmd == 3 {
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		_, err := fastState.Step(true)
		require.Contains(t, err.Error(), "f0012")
		// TODO: Test EVM & slow VM
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
		runEVM(t, contracts, addrs, stepWitness, fastPost)
		runSlow(t, stepWitness, fastPost, nil)
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
