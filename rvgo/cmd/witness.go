package cmd

import (
	"fmt"
	"os"

	cannon "github.com/ethereum-optimism/optimism/cannon/cmd"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

func Witness(ctx *cli.Context) error {
	input := ctx.Path(cannon.WitnessInputFlag.Name)
	output := ctx.Path(cannon.WitnessOutputFlag.Name)
	state, err := cannon.LoadJSON[fast.VMState](input)
	if err != nil {
		return fmt.Errorf("invalid input state (%v): %w", input, err)
	}
	witness := state.EncodeWitness()
	h, err := witness.StateHash()
	if err != nil {
		return fmt.Errorf("failed to compute witness hash: %w", err)
	}
	if output != "" {
		if err := os.WriteFile(output, witness, 0755); err != nil {
			return fmt.Errorf("writing output to %v: %w", output, err)
		}
	}
	fmt.Println(h.Hex())
	return nil
}

var WitnessCommand = &cli.Command{
	Name:        "witness",
	Usage:       "Convert a Asterisc JSON state into a binary witness",
	Description: "Convert a Asterisc JSON state into a binary witness. The hash of the witness is written to stdout",
	Action:      Witness,
	Flags: []cli.Flag{
		cannon.WitnessInputFlag,
		cannon.WitnessOutputFlag,
	},
}
