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
	Typ() PreimageTyp
	Hash() [32]byte
}

type ProgramInputKey uint64

func (k ProgramInputKey) Typ() PreimageTyp {
	return ProgramInputPreimageTyp
}

func (k ProgramInputKey) Hash() [32]byte {
	var out [32]byte
	binary.BigEndian.PutUint64(out[24:], uint64(k))
	return out
}

type Keccak256PreimageKey [32]byte

func (k Keccak256PreimageKey) Typ() PreimageTyp {
	return Keccak256HashPreimageTyp
}

func (k Keccak256PreimageKey) Hash() [32]byte {
	return k
}

type PreimageOracle struct {
	F *os.File
}

func (p *PreimageOracle) Get(k PreimageKey) []byte {
	typ := k.Typ()
	if _, err := io.Copy(p.F, bytes.NewReader(typ[:])); err != nil {
		panic(fmt.Errorf("failed to write type %x to preimage oracle: %w", typ, err))
	}
	hash := k.Hash()
	if _, err := io.Copy(p.F, bytes.NewReader(hash[:])); err != nil {
		panic(fmt.Errorf("failed to write hash %x to preimage oracle: %w", hash, err))
	}
	dat, err := io.ReadAll(p.F)
	if err != nil {
		panic(fmt.Errorf("failed to read preimage of (%x, %x) from oracle: %w", typ, hash, err))
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
	fmt.Println("starting!")
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
		preimageHinter.Hint(Keccak256HashPreimageTyp, []byte("hello world!"))
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
