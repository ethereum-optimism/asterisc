package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	cannon "github.com/ethereum-optimism/optimism/cannon/cmd"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type WitnessOutput struct {
	Witness   hexutil.Bytes `json:"witness"`
	StateHash common.Hash   `json:"stateHash"`
	Step      uint64        `json:"step"`
	Exited    bool          `json:"exited"`
	ExitCode  uint8         `json:"exitCode"`
	PC        uint64        `json:"pc"`
}

func Witness(ctx *cli.Context) error {
	input := ctx.Path(cannon.WitnessInputFlag.Name)
	witnessOutput := ctx.Path(cannon.WitnessOutputFlag.Name)
	state, err := fast.LoadVMStateFromFile(input)
	if err != nil {
		return fmt.Errorf("invalid input state (%v): %w", input, err)
	}

	witness := state.EncodeWitness()
	stateHash, err := witness.StateHash()
	if err != nil {
		return fmt.Errorf("invalid Witness (%v): %w", witness, err)
	}

	if witnessOutput != "" {
		if err := os.WriteFile(witnessOutput, witness, OutFilePerm); err != nil {
			return fmt.Errorf("writing output to %v: %w", witnessOutput, err)
		}
	}
	output := &WitnessOutput{
		Witness:   []byte(witness),
		StateHash: stateHash,
		Step:      state.GetStep(),
		Exited:    state.Exited,
		ExitCode:  state.ExitCode,
		PC:        state.PC,
	}

	if err := jsonutil.WriteJSON(output, ioutil.ToStdOut()); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

var WitnessCommand = &cli.Command{
	Name:        "witness",
	Usage:       "Convert an Asterisc JSON/binary state into a binary witness",
	Description: "Convert an Asterisc JSON/binary state into a binary witness. Basic Data about the state is printed to stdout in JSON format",
	Action:      Witness,
	Flags: []cli.Flag{
		cannon.WitnessInputFlag,
		cannon.WitnessOutputFlag,
	},
}
