package main

import (
	"fmt"

	"github.com/protolambda/asterisc/rvgo"
)

func main() {
	state := rvgo.NewVMState()
	// TODO load program into memory

	// TODO run N instructions the fast way

	// TODO capture instruction sub-steps:
	so := rvgo.NewStateOracle()
	pre := rvgo.VMScratchpad{StateRoot: state.Merkleize(so)}
	post := rvgo.Step(pre, so)
	al := so.AccessList()

	fmt.Println(post, al)

	// replicate the step in Go
	so2 := &rvgo.AccessListOracle{AccessList: al}
	post2 := rvgo.Step(pre, so2)
	if post != post2 {
		panic("failed to replicate step")
	}

	// TODO: replicate the step in EVM
	// TODO encode pre-state and access-list as EVM function call args
	// TODO run EVM
	// TODO compare result
}
