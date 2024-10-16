package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ABI types
var (
	asteriscMemoryProof, _ = abi.NewType("tuple", "AsteriscMemoryProof", []abi.ArgumentMarshaling{
		{Name: "memRoot", Type: "bytes32"},
		{Name: "proof", Type: "bytes"},
	})
	asteriscMemoryProofArgs = abi.Arguments{
		{Name: "encodedAsteriscMemoryProof", Type: asteriscMemoryProof},
	}
)

func DiffTestUtils() {
	args := os.Args[2:]
	variant := args[0]

	// This command requires arguments
	if len(args) == 0 {
		panic("Error: No arguments provided")
	}

	switch variant {
	case "asteriscMemoryProof":
		// <pc, insn, [memAddr, memValue]>
		if len(args) != 3 && len(args) != 5 {
			panic("Error: asteriscMemoryProofWithProof requires 2 or 4 arguments")
		}
		mem := fast.NewMemory()
		pc, err := strconv.ParseUint(args[1], 10, 64)
		checkErr(err, "Error decoding addr")
		insn, err := strconv.ParseUint(args[2], 10, 32)
		checkErr(err, "Error decoding insn")
		instBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(instBytes, uint32(insn))
		mem.SetAligned(uint64(pc), instBytes)

		// proof size: 64-5+1=60 (a 64-bit mem-address branch to 32 byte leaf, incl leaf itself), all 32 bytes
		// 60 * 32 = 1920
		var insnProof, memProof [1920]byte
		if len(args) == 5 {
			memAddr, err := strconv.ParseUint(args[3], 10, 64)
			checkErr(err, "Error decoding memAddr")
			memValue, err := hex.DecodeString(strings.TrimPrefix(args[4], "0x"))
			checkErr(err, "Error decoding memValue")
			mem.SetAligned(uint64(memAddr), memValue)
			memProof = mem.MerkleProof(uint64(memAddr))
		}
		insnProof = mem.MerkleProof(uint64(pc))

		output := struct {
			MemRoot common.Hash
			Proof   []byte
		}{
			MemRoot: mem.MerkleRoot(),
			Proof:   append(insnProof[:], memProof[:]...),
		}
		packed, err := asteriscMemoryProofArgs.Pack(&output)
		checkErr(err, "Error encoding output")
		fmt.Print(hexutil.Encode(packed[32:]))
	default:
		panic(fmt.Errorf("unknown command: %s", args[0]))
	}
}
