package fast

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/protolambda/asterisc/rvgo/fast"
)

func TestFastSimple(t *testing.T) {
	programELF, err := elf.Open("../tests/go-tests/bin/simple")
	require.NoError(t, err)
	defer programELF.Close()

	vmState, err := fast.LoadELF(programELF)
	require.NoError(t, err, "must load test suite ELF binary")

	err = fast.PatchVM(programELF, vmState)
	require.NoError(t, err, "must patch Go runtime.gcenable")

	symbols, err := fast.Symbols(programELF)
	require.NoError(t, err)

	// TODO: by hashing together the type and key, we can reduce it to a single map?

	inputs := make(map[[32]byte][]byte)
	addInput := func(i uint64, val []byte) {
		var k [32]byte
		binary.BigEndian.PutUint64(k[24:], i)
		inputs[k] = val
	}

	preImages := make(map[[32]byte][]byte)
	addPreimage := func(img []byte) {
		preImages[crypto.Keccak256Hash(img)] = img
	}

	addInput(0, []byte{0})                                // onchain mode, no pre-image hints fd
	addInput(1, crypto.Keccak256([]byte("hello")))        // pre-state
	addInput(2, crypto.Keccak256([]byte("world")))        // input
	addInput(3, crypto.Keccak256([]byte("hello world!"))) // claim to verify
	addPreimage([]byte("hello"))                          // pre-state pre-image
	addPreimage([]byte("world"))                          // input pre-image

	vmState.PreimageOracle = func(typ [32]byte, key [32]byte) ([]byte, error) {
		switch typ {
		case [32]byte{}: // input
			if v, ok := inputs[key]; ok {
				return v, nil
			} else {
				return nil, fmt.Errorf("unknown input %x", key)
			}
		case [32]byte{31: 1}: // keccak256 hash
			if v, ok := preImages[key]; ok {
				return v, nil
			} else {
				return nil, fmt.Errorf("unknown pre-image %x", key)
			}
		default:
			return nil, fmt.Errorf("unknown pre-image type %x", typ)
		}
	}

	//lastSym := ""
	for i := 0; i < 300_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)
		if sym != nil {
			//if sym.Name != lastSym {
			//	lastSym = sym.Name
			instr := vmState.Instr()
			fmt.Printf("i: %4d  pc: 0x%x  instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, instr, sym.Name, sym.Size)
			if sym.Name == "runtime.throw" {
				throwArg := vmState.Registers[10]
				throwArgLen := vmState.Registers[11]
				if throwArgLen > 1000 {
					throwArgLen = 1000
				}
				x := vmState.GetMemRange(throwArg, throwArgLen)
				dat, _ := io.ReadAll(x)
				if utf8.Valid(dat) {
					fmt.Printf("THROW! %q\n", string(dat))
				} else {
					fmt.Printf("THROW! %016x: %x\n", throwArg, dat)
				}

				x = vmState.GetMemRange(0x1949e0, 0x1949e0+512)
				dat, _ = io.ReadAll(x)
				fmt.Printf("panic data buffer: %x\n", dat)
				break
			}
			if sym.Name == "runtime.recordForPanic" {
				recordArg := vmState.Registers[10]
				x := vmState.GetMemRange(recordArg, 32)
				dat, _ := io.ReadAll(x)
				fmt.Printf("recordForPanic! %016x: %x\n", recordArg, dat)
			}
			//}
		}
		//else {
		//fmt.Printf("i: %4d  pc: 0x%x  instr: %08x\n", i, vmState.PC, instr)
		//}
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)

		}
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		t.Fatalf("failed with exit code %d", vmState.Exit)
	}
}

func TestFastMinimal(t *testing.T) {
	programELF, err := elf.Open("../tests/go-tests/bin/minimal")
	require.NoError(t, err)
	defer programELF.Close()

	vmState, err := fast.LoadELF(programELF)
	require.NoError(t, err, "must load test suite ELF binary")

	err = fast.PatchVM(programELF, vmState)
	require.NoError(t, err, "must patch VM")

	symbols, err := fast.Symbols(programELF)
	require.NoError(t, err)

	vmState.PreimageOracle = func(typ [32]byte, key [32]byte) ([]byte, error) {
		return nil, fmt.Errorf("unknown key %x", key)
	}

	//lastSym := ""
	for i := 0; i < 300_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)
		if sym != nil {
			//if sym.Name != lastSym {
			//	lastSym = sym.Name
			instr := vmState.Instr()
			fmt.Printf("i: %4d  pc: 0x%x  offset: %03x instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, vmState.PC-sym.Value, instr, sym.Name, sym.Size)
			if sym.Name == "runtime.throw" {
				throwArg := vmState.Registers[10]
				throwArgLen := vmState.Registers[11]
				if throwArgLen > 1000 {
					throwArgLen = 1000
				}
				x := vmState.GetMemRange(throwArg, throwArgLen)
				dat, _ := io.ReadAll(x)
				if utf8.Valid(dat) {
					fmt.Printf("THROW! %q\n", string(dat))
				} else {
					fmt.Printf("THROW! %016x: %x\n", throwArg, dat)
				}

				//x = vmState.GetMemRange(0x1949e0, 0x1949e0+512)
				//dat, _ = io.ReadAll(x)
				//fmt.Printf("panic data buffer: %x\n", dat)
				break
			}
			//}
		}
		//else {
		//fmt.Printf("i: %4d  pc: 0x%x  instr: %08x\n", i, vmState.PC, instr)
		//}
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)

		}
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		t.Fatalf("failed with exit code %d", vmState.Exit)
	}
}
