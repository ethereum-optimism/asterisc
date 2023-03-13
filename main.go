package main

import (
	"debug/elf"
	"log"

	"github.com/protolambda/asterisc/rvgo/fast"
	"github.com/protolambda/asterisc/rvgo/oracle"
	"github.com/protolambda/asterisc/rvgo/slow"
)

func main() {
	programELF, err := elf.Open("program.bin")
	if err != nil {
		log.Fatalf("failed to open program ELF: %v", err)
	}
	defer programELF.Close()
	vmState, err := fast.LoadELF(programELF)
	if err != nil {
		log.Fatalf("failed to load ELF into VM state: %v", err)
	}

	// run through agreed instruction steps the fast way
	instructionStep := 1000
	for i := 0; i < instructionStep; i++ {
		fast.Step(vmState)
	}

	so := oracle.NewStateOracle()
	pre := vmState.Merkleize(so)
	log.Printf("pre-state: %x", pre)

	// Now run through the disputed step.
	// And remember all state we access, so we can reproduce it without full state oracle.
	so.BuildAccessList(true)
	post := slow.Step(pre, so)

	log.Printf("post-state: %x", post)
	al := so.AccessList()
	evmInput := oracle.Input(al, pre)

	log.Printf("proof calldata: %x", evmInput)
}
