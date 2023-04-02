package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

func keccak256(dat []byte) [32]byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(dat)
	return *(*[32]byte)(h.Sum(nil))
}

type PreimageTyp [32]byte

var (
	ProgramInputPreimageTyp  = PreimageTyp{31: 0}
	Keccak256HashPreimageTyp = PreimageTyp{31: 1}
)

type PreimageKey interface {
	Key() [32]byte
}

type ProgramInputKey uint64

func (k ProgramInputKey) Key() [32]byte {
	var out [64]byte // key, type
	binary.BigEndian.PutUint64(out[24:32], uint64(k))
	return keccak256(out[:])
}

type Keccak256PreimageKey [32]byte

func (k Keccak256PreimageKey) Key() [32]byte {
	var out [64]byte // key, type
	copy(out[:32], k[:])
	out[32] = 1
	return keccak256(out[:])
}

type PreimageOracle struct {
	F *os.File
}

func (p *PreimageOracle) Get(k PreimageKey) []byte {
	key := k.Key()
	if _, err := io.Copy(p.F, bytes.NewReader(key[:])); err != nil {
		panic(fmt.Errorf("failed to write key %x to preimage oracle: %w", key, err))
	}
	dat, err := io.ReadAll(p.F)
	if err != nil {
		panic(fmt.Errorf("failed to read preimage %s (%T) from oracle: %w", key, k, err))
	}
	return dat
}

type PreimageHinter struct {
	F *os.File
}

func (p *PreimageHinter) Hint(typ PreimageTyp, dat []byte) {
	if typ == ProgramInputPreimageTyp {
		panic("cannot hint new program inputs")
	}
	if _, err := io.Copy(p.F, bytes.NewReader(typ[:])); err != nil {
		panic(fmt.Errorf("failed to write type %x to preimage hinter: %w", typ, err))
	}
	var l [8]byte
	binary.BigEndian.PutUint64(l[:], uint64(len(dat)))
	if _, err := io.Copy(p.F, bytes.NewReader(l[:])); err != nil {
		panic(fmt.Errorf("failed to write length %d to preimage hinter: %w", len(dat), err))
	}
	if _, err := io.Copy(p.F, bytes.NewReader(dat)); err != nil {
		panic(fmt.Errorf("failed to write data to preimage hinter: %w", err))
	}
}

func main() {
	_, _ = os.Stdout.Write([]byte("hello world!\n"))
	fmt.Println("starting!") // uses sync pool
	preimageOracle := PreimageOracle{F: os.NewFile(3, "preimage-oracle")}
	mode := preimageOracle.Get(ProgramInputKey(0))
	fmt.Printf("operating in %x mode\n", mode)

	// 1 input that commits to 'prestate', 'input', and 'claim' would also be enough,
	// but this simplifies it, as the structure of the inputs is not part of the program to execute this way.
	preState := *(*[32]byte)(preimageOracle.Get(ProgramInputKey(1)))
	fmt.Printf("preState: %x\n", preState)

	input := *(*[32]byte)(preimageOracle.Get(ProgramInputKey(2)))
	fmt.Printf("input: %x\n", input)

	claim := *(*[32]byte)(preimageOracle.Get(ProgramInputKey(3)))
	fmt.Printf("claim: %x\n", claim)

	if bytes.Equal(mode, []byte{1}) { // determine what mode we are in, and determine pre-images if necessary.
		preimageHinter := PreimageHinter{F: os.NewFile(4, "preimage-hinter")}

		preimageHinter.Hint(Keccak256HashPreimageTyp, []byte("hello"))
		preimageHinter.Hint(Keccak256HashPreimageTyp, []byte("world"))
	}

	// take inputs, and compute new output
	x := string(preimageOracle.Get(Keccak256PreimageKey(preState)))
	y := string(preimageOracle.Get(Keccak256PreimageKey(input)))
	result := x + " " + y + "!"
	fmt.Println("result:", result)

	outputRoot := keccak256([]byte(result))

	// verify the claim
	if claim != outputRoot {
		os.Exit(1)
	}
	os.Exit(0)
}
