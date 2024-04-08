package cmd

import (
	"fmt"

	cannon "github.com/ethereum-optimism/optimism/cannon/cmd"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

type WitnessOutput struct {
	Witness   []byte   `json:"witness"`
	StateHash [32]byte `json:"stateHash"`
}

func Witness(ctx *cli.Context) error {
	input := ctx.Path(cannon.WitnessInputFlag.Name)
	output := ctx.Path(cannon.WitnessOutputFlag.Name)
	state, err := jsonutil.LoadJSON[fast.VMState](input)
	if err != nil {
		return fmt.Errorf("invalid input state (%v): %w", input, err)
	}
	witness := state.EncodeWitness()
	stateHash, err := witness.StateHash()
	if err != nil {
		return fmt.Errorf("failed to compute witness hash: %w", err)
	}
	witnessOutput := &WitnessOutput{
		Witness:   witness,
		StateHash: stateHash,
	}
	if err := jsonutil.WriteJSON(output, witnessOutput, OutFilePerm); err != nil {
		return fmt.Errorf("failed to write witness output %w", err)
	}
	fmt.Println(stateHash.Hex())
	return nil
}

var WitnessCommand = &cli.Command{
	Name:        "witness",
	Usage:       "Convert an Asterisc JSON state into a binary witness",
	Description: "Convert an Asterisc JSON state into a binary witness. The statehash is written to stdout",
	Action:      Witness,
	Flags: []cli.Flag{
		cannon.WitnessInputFlag,
		cannon.WitnessOutputFlag,
	},
}
