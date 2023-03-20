package fast

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

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

	for i := 0; i < 200_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)
		instr := vmState.Instr()
		if sym != nil {
			fmt.Printf("i: %4d  pc: 0x%x  instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, instr, sym.Name, sym.Size)
		} else {
			fmt.Printf("i: %4d  pc: 0x%x  instr: %08x\n", i, vmState.PC, instr)
		}
		if sym.Name == "runtime.throw" {
			break
		}
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)
		}
		fmt.Println()
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		t.Fatalf("failed with exit code %d", vmState.Exit)
	}
}
