package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
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
	// TODO for fuzzing: avoid creating a new evm environment each time
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

// The following provides a single entrypoint to fuzz all syscalls.
// go test -run NOTAREALTEST -v -fuzztime 8m -fuzz=FuzzEverything ./rvgo/test --parallel 15
func FuzzEverything(f *testing.F) {
	f.Fuzz(func(t *testing.T, randomFunc uint64, addrSeed uint64, resSeed uint64, seed uint64, fdSeed uint64, heap uint64, exitCode uint8, pcSeed uint64, stepSeed uint64, randSeed int64) {

		// it seems that inputs can be even more random from golang rand vs fuzzing inputs
		reservation := randomUint64FromUint(resSeed)
		addr := randomUint64FromUint(addrSeed)
		fd := randomUint64FromUint(fdSeed) % (math.MaxUint8)
		pc := randomUint64FromUint(pcSeed)
		step := randomUint64FromUint(stepSeed)

		switch randomFunc {
		case riscv.SysExit:
			{
				StateSyscallExit(t, riscv.SysExit, exitCode, heap, reservation, pc, step)
			}
		case riscv.SysExitGroup:
			{
				StateSyscallExit(t, riscv.SysExitGroup, exitCode, heap, reservation, pc, step)
			}
		case riscv.SysBrk:
			{
				StateSyscallBrk(t, riscv.SysBrk, heap, reservation, pc, step, exitCode)
			}
		case riscv.SysMmap:
			{
				StateSyscallMmap(t, addr, seed, heap, fd, uint64(randSeed), pc, step, reservation, exitCode)
			}
		case riscv.SysFcntl:
			{
				outputErrorCode := uint64(0xFFFF_FFFF_FFFF_FFFF)
				cmd := seed
				switch cmd {
				case 1:
					{
						if fd <= 6 {
							// for any cmd=1 where fd <= 6, expect out = 0 and err = 0
							StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, 0, 0)
						} else {
							// for any cmd=1, fd >= 7, expect 0xFF..FF and err = 0x4d
							StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, outputErrorCode, 0x4d)
						}
					}
				case 3:
					{
						switch fd {
						case riscv.FdStdin, riscv.FdHintRead, riscv.FdPreimageRead:
							{
								StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, 0, 0)
							}
						case riscv.FdStdout, riscv.FdStderr, riscv.FdHintWrite, riscv.FdPreimageWrite:
							{
								StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, 1, 0)
							}
						default:
							{
								StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, outputErrorCode, 0x4d)
							}
						}
					}
				default:
					{ // any other input should produce an error with respective error code (including cmd=2)
						StateSyscallFcntl(t, fd, cmd, pc, heap, reservation, step, outputErrorCode, 0x16)
					}
				}
			}
		case riscv.SysOpenat:
			{
				StateSyscallOpenat(t, pc, step, heap, reservation)
			}
		case riscv.SysClockGettime:
			{
				StateSyscallClockGettime(t, addr, pc, step, heap, reservation)
			}
		case riscv.SysClone:
			{
				StateSyscallClone(t, pc, step, heap, reservation)
			}
		case riscv.SysGetrlimit:
			{
				rlimit := seed
				if rlimit == 7 {
					StateSyscallGetrlimit(t, addr, pc, step)
				} else {
					StatesyscallgetrlimitError(t, rlimit, addr, pc, step)
				}
			}
		case riscv.SysRead:
			{
				count := reservation

				randBytes, err := randomBytes(randSeed, fd)
				require.NoError(t, err)

				switch fd {
				case riscv.FdStdin:
					{
						StateSyscallRead(t, riscv.FdStdin, addr, count, pc, step, 0, 0)
					}
				case riscv.FdHintRead:
					{
						StateHintRead(t, addr, count, seed, pc, step, randBytes)
					}
				case riscv.FdPreimageRead:
					{
						StatePreimageRead(t, addr, count, seed, randBytes, pc, step)
					}
				default:
					{
						StateSyscallRead(t, fd, addr, count, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)
					}

				}
			}
		case riscv.SysWrite:
			{
				opcode := seed

				randBytes, err := randomBytes(randSeed, fd)
				require.NoError(t, err)

				switch opcode {
				case riscv.FdStdout, riscv.FdStderr:
					{
						count := heap
						StateSyscallWrite(t, opcode, addr, count, pc, step, count, 0)
					}
				case riscv.FdHintWrite:
					{
						StateHintWrite(t, addr, seed, heap, randBytes, pc, step, randSeed)
					}
				case riscv.FdPreimageWrite:
					{
						preimageOffset := reservation

						StatePreimageWrite(t, addr, heap, preimageOffset, randBytes, pc, step)
					}
				default:
					{
						count := heap
						StateSyscallWrite(t, opcode, addr, count, pc, step, 0xFFFF_FFFF_FFFF_FFFF, 0x4d)
					}
				}
			}
		case riscv.SysPrlimit64, riscv.SysFutex, riscv.SysNanosleep:
			{
				// Tests unsupported syscalls
				unsupportedSyscalls := []int{
					riscv.SysPrlimit64,
					riscv.SysFutex,
					riscv.SysNanosleep,
				}
				// index should be between [0, len(unsupportedSyscalls), exclusive of end
				index := int(randomUint64FromUint(resSeed)) % len(unsupportedSyscalls)
				if index < 0 {
					index = -index
				}
				StateSyscallUnsupported(t, unsupportedSyscalls[index], heap, reservation)
			}
		default:
			{
				// if not encompassed in cases above, the syscall should results in a noop
				StateSyscallNoop(t, randomFunc, seed, pc, step)
			}
		}
	})
}

func StateSyscallExit(t *testing.T, syscall int, exitCode uint8, heap uint64, reservation uint64, pc uint64, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
		Registers:       [32]uint64{17: uint64(syscall), 10: uint64(exitCode)},
		Step:            step,
	}
	state.Memory.SetUnaligned(pc, syscallInsn)
	preStateRoot := state.Memory.MerkleRoot()
	preStateRegisters := state.Registers
	preStateProof := state.Memory.MerkleProof(pc)

	fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
	stepWitness, err := fastState.Step(true)
	require.NoError(t, err)
	require.False(t, stepWitness.HasPreimage())

	require.Equal(t, pc+4, state.PC) // PC must advance
	require.Equal(t, heap, state.Heap)
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, exitCode, state.ExitCode)
	require.Equal(t, true, state.Exited)
	require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
	require.Equal(t, preStateProof, state.Memory.MerkleProof(pc))
	require.Equal(t, step+1, state.Step) // Step must advance
	require.Equal(t, preStateRegisters, state.Registers)

	fastPost := state.EncodeWitness()
	// TODO: The impact of these two checks is intermediary results are not checked – only the final result
	runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
	runSlow(t, stepWitness, fastPost, nil, nil)
}

func StateSyscallBrk(t *testing.T, syscall int, heap uint64, reservation uint64, pc uint64, step uint64, exitCode uint8) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        exitCode,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
		Registers:       [32]uint64{17: uint64(syscall)},
		Step:            step,
	}
	state.Memory.SetUnaligned(pc, syscallInsn)
	preStateRoot := state.Memory.MerkleRoot()
	expectedRegisters := state.Registers
	expectedRegisters[10] = 1 << 30
	expectedRegisters[11] = 0
	preStateProof := state.Memory.MerkleProof(pc)

	fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
	stepWitness, err := fastState.Step(true)
	require.NoError(t, err)
	require.False(t, stepWitness.HasPreimage())

	require.Equal(t, pc+4, state.PC) // PC must advance
	require.Equal(t, heap, state.Heap)
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, exitCode, state.ExitCode)
	require.Equal(t, false, state.Exited) // Must not be exited
	require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
	require.Equal(t, preStateProof, state.Memory.MerkleProof(pc))
	require.Equal(t, step+1, state.Step) // Step must advance
	require.Equal(t, expectedRegisters, state.Registers)

	fastPost := state.EncodeWitness()
	runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
	runSlow(t, stepWitness, fastPost, nil, nil)
}

func StateSyscallMmap(t *testing.T, addr uint64, length uint64, heap uint64, flag uint64, fd uint64, pc uint64, step uint64, reservation uint64, exitCode uint8) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        exitCode,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
		Registers: [32]uint64{
			17: riscv.SysMmap,
			10: addr,
			11: length,
			13: flag, // MAP_ANONYMOUS flag
			14: fd,
		},
		Step: step,
	}
	state.Memory.SetUnaligned(pc, syscallInsn)
	preStateRoot := state.Memory.MerkleRoot()
	expectedRegisters := state.Registers

	// true if flag&0x20 == 0 or fd != mask
	shouldSkipZero := false

	u64Mask := uint64(0xFFFF_FFFF_FFFF_FFFF)
	if (flag&0x20 == 0) || fd != u64Mask {
		expectedRegisters[10] = u64Mask
		expectedRegisters[11] = 0x4d
		shouldSkipZero = true
	} else {
		expectedRegisters[10] = addr // although this should be unchanged from before
		expectedRegisters[11] = 0    // error code is set to zero
	}

	fastState := fast.NewInstrumentedState(state, nil, os.Stdout, os.Stderr)
	stepWitness, err := fastState.Step(true)
	require.NoError(t, err)
	require.False(t, stepWitness.HasPreimage())

	require.Equal(t, pc+4, state.PC) // PC must advance
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, exitCode, state.ExitCode) // ExitCode must be set
	require.Equal(t, false, state.Exited)      // Must not be exited
	require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
	require.Equal(t, step+1, state.Step) // Step must advance

	newHeap := heap

	// these checks should be skipped if conditions on L#560 are met
	if !shouldSkipZero && addr == 0 {
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
}

func StateSyscallFcntl(t *testing.T, fd, cmd, pc, heap, reservation, step, out, errCode uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
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
	require.Equal(t, heap, state.Heap)
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

func StateSyscallOpenat(t *testing.T, pc, step, heap, reservation uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
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
	require.Equal(t, heap, state.Heap)
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
	require.Equal(t, false, state.Exited)      // Must not be exited
	require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
	require.Equal(t, step+1, state.Step) // Step must advance
	require.Equal(t, expectedRegisters, state.Registers)

	fastPost := state.EncodeWitness()
	runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
	runSlow(t, stepWitness, fastPost, nil, nil)

}

func StateSyscallClockGettime(t *testing.T, addr, pc, step, heap, reservation uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
		Registers:       [32]uint64{17: riscv.SysClockGettime, 11: addr},
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
	binary.LittleEndian.PutUint64(bytes[:], 1337)

	postMemory.SetUnaligned(addr, bytes[:])
	postMemory.SetUnaligned(addr+8, []byte{42, 0, 0, 0, 0, 0, 0, 0})

	require.Equal(t, pc+4, state.PC) // PC must advance
	require.Equal(t, heap, state.Heap)
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
	require.Equal(t, false, state.Exited)      // Must not be exited
	require.Equal(t, state.Memory.MerkleRoot(), postMemory.MerkleRoot())
	require.Equal(t, step+1, state.Step) // Step must advance
	require.Equal(t, expectedRegisters, state.Registers)

	fastPost := state.EncodeWitness()
	runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
	runSlow(t, stepWitness, fastPost, nil, nil)
}

func StateSyscallClone(t *testing.T, pc, step, heap, reservation uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
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
	require.Equal(t, heap, state.Heap)
	require.Equal(t, reservation, state.LoadReservation)
	require.Equal(t, uint8(0), state.ExitCode) // ExitCode must be set
	require.Equal(t, false, state.Exited)      // Must not be exited
	require.Equal(t, preStateRoot, state.Memory.MerkleRoot())
	require.Equal(t, step+1, state.Step) // Step must advance
	require.Equal(t, expectedRegisters, state.Registers)

	fastPost := state.EncodeWitness()
	runEVM(t, contracts, addrs, stepWitness, fastPost, nil)
	runSlow(t, stepWitness, fastPost, nil, nil)
}

func StateSyscallGetrlimit(t *testing.T, addr, pc, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

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

func StatesyscallgetrlimitError(t *testing.T, res, addr, pc, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

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

func StateSyscallNoop(t *testing.T, syscall uint64, arg uint64, pc uint64, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	state := &fast.VMState{
		PC:              pc,
		Heap:            0,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: 0,
		Registers:       [32]uint64{17: syscall, 10: arg},
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

func StateSyscallRead(t *testing.T, fd, addr, count, pc, step, ret, errCode uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

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

func StateHintRead(t *testing.T, addr uint64, count uint64, preimageOffset uint64, pc uint64, step uint64, preimageData []byte) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
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
}

func StatePreimageRead(t *testing.T, addr uint64, count uint64, preimageOffset uint64, preimageData []byte, pc uint64, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
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
		// TODO: add more nuanced check here such that if first 7 bytes are zeroed, only check NotEqual
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
}

func StateSyscallWrite(t *testing.T, fd, addr, count, pc, step, ret, errCode uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

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

func StateHintWrite(t *testing.T, addr uint64, count uint64, preimageOffset uint64, preimageData []byte, pc uint64, step uint64, randSeed int64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
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
		Registers:       [32]uint64{17: riscv.SysWrite, 10: riscv.FdHintWrite, 12: count},
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
}

func StatePreimageWrite(t *testing.T, addr uint64, heap uint64, preimageOffset uint64, preimageData []byte, pc uint64, step uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc = pc & 0xFF_FF_FF_FF_FF_FF_FF_FC // align PC
	if preimageOffset >= uint64(len(preimageData)) {
		t.SkipNow()
	}
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
}

func StateSyscallUnsupported(t *testing.T, syscall int, heap uint64, reservation uint64) {
	contracts := testContracts(t)
	addrs := testAddrs

	pc := uint64(0)
	state := &fast.VMState{
		PC:              pc,
		Heap:            heap,
		ExitCode:        0,
		Exited:          false,
		Memory:          fast.NewMemory(),
		LoadReservation: reservation,
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

}

func randomBytes(seed int64, length uint64) ([]byte, error) {
	r := rand.New(rand.NewSource(seed))
	randBytes := make([]byte, length)
	if _, err := r.Read(randBytes); err != nil {
		return nil, err
	}
	return randBytes, nil
}
func randomUint64FromUint(seed uint64) uint64 {
	r := rand.New(rand.NewSource(int64(seed)))
	return r.Uint64()
}
