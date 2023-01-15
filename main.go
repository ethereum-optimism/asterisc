package main

import (
	"fmt"

	"github.com/protolambda/asterisc/rvgo/fast"
	"github.com/protolambda/asterisc/rvgo/oracle"
	"github.com/protolambda/asterisc/rvgo/slow"
)

func main() {
	state := fast.NewVMState()
	// TODO load program into memory

	// run through agreed instruction steps the fast way
	instructionStep := 1000
	for i := 0; i < instructionStep; i++ {
		fast.Step(state)
	}

	instructionSubStep := 10

	so := oracle.NewStateOracle()
	pre := slow.VMSubState{StateRoot: state.Merkleize(so)} // oracle will remember merkleized state
	// run through agreed instruction sub-steps
	for i := 0; i < instructionSubStep; i++ {
		pre = slow.SubStep(pre, so)
	}

	// Now run through the sub-step of dispute.
	// And remember all state we access, so we can reproduce it without full state oracle.
	so.BuildAccessList(true)
	post := slow.SubStep(pre, so)
	al := so.AccessList()
	fmt.Println("produced post sub-step post-state with access list!", post, al)

	// replicate the step in Go with only the access list data, and copy of pre-state
	so2 := &oracle.AccessListOracle{AccessList: al}
	post2 := slow.SubStep(pre, so2)
	if post != post2 {
		panic("failed to replicate step")
	}

	// TODO: replicate the sub-step in EVM
	// TODO encode pre-state and access-list as EVM function call args
	// TODO run EVM
	// TODO compare result
}
