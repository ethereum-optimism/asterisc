package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-preimage"
)

type rawHint string

func (rh rawHint) Hint() string {
	return string(rh)
}

func main() {
	_, _ = os.Stdout.Write([]byte("hello world!\n"))
	fmt.Println("starting!") // uses sync pool

	po := preimage.NewOracleClient(preimage.ClientPreimageChannel())
	hinter := preimage.NewHintWriter(preimage.ClientHinterChannel())

	mode := po.Get(preimage.LocalIndexKey(0))
	fmt.Printf("operating in %x mode\n", mode)

	// 1 input that commits to 'prestate', 'input', and 'claim' would also be enough,
	// but this simplifies it, as the structure of the inputs is not part of the program to execute this way.
	preState := *(*[32]byte)(po.Get(preimage.LocalIndexKey(0)))
	fmt.Printf("preState: %x\n", preState)

	input := *(*[32]byte)(po.Get(preimage.LocalIndexKey(1)))
	fmt.Printf("input: %x\n", input)

	claim := *(*[32]byte)(po.Get(preimage.LocalIndexKey(2)))
	fmt.Printf("claim: %x\n", claim)

	hinter.Hint(rawHint("hello"))
	hinter.Hint(rawHint("world"))

	//// take inputs, and compute new output
	x := string(po.Get(preimage.Keccak256Key(preState)))
	fmt.Printf("x: %x\n", []byte(x))
	//y := string(po.Get(preimage.Keccak256Key(input)))
	//result := x + " " + y + "!"
	//fmt.Println("result:", result)
	//
	//outputRoot := crypto.Keccak256Hash([]byte(result))
	//
	//// verify the claim
	//if claim != outputRoot {
	//	os.Exit(1)
	//}
	os.Exit(0)
}
