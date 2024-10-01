package cmd

import (
	"debug/elf"
	"fmt"

	cannon "github.com/ethereum-optimism/optimism/cannon/cmd"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

func LoadELF(ctx *cli.Context) error {
	elfPath := ctx.Path(cannon.LoadELFPathFlag.Name)
	elfProgram, err := elf.Open(elfPath)
	if err != nil {
		return fmt.Errorf("failed to open ELF file %q: %w", elfPath, err)
	}
	if elfProgram.Machine != elf.EM_RISCV {
		return fmt.Errorf("ELF is not RISC-V, but got %q", elfProgram.Machine.String())
	}
	state, err := fast.LoadELF(elfProgram)
	if err != nil {
		return fmt.Errorf("failed to load ELF data into VM state: %w", err)
	}
	err = fast.PatchVM(elfProgram, state)
	if err != nil {
		return fmt.Errorf("failed to patch VM")
	}
	meta, err := MakeMetadata(elfProgram)
	if err != nil {
		return fmt.Errorf("failed to compute program metadata: %w", err)
	}
	// Must set witness and stateHash after initial state is prepared
	if err := state.SetWitnessAndStateHash(); err != nil {
		return fmt.Errorf("failed to set witness and stateHash: %w", err)
	}
	if err := jsonutil.WriteJSON[*Metadata](meta, ioutil.ToStdOutOrFileOrNoop(ctx.Path(cannon.LoadELFMetaFlag.Name), OutFilePerm)); err != nil {
		return fmt.Errorf("failed to output metadata: %w", err)
	}
	return jsonutil.WriteJSON[*fast.VMState](state, ioutil.ToStdOutOrFileOrNoop(ctx.Path(cannon.LoadELFOutFlag.Name), OutFilePerm))
}

var LoadELFCommand = &cli.Command{
	Name:        "load-elf",
	Usage:       "Load ELF file into Asterisc JSON state",
	Description: "Load ELF file into Asterisc JSON state, optionally patch out functions",
	Action:      LoadELF,
	Flags: []cli.Flag{
		cannon.LoadELFPathFlag,
		cannon.LoadELFOutFlag,
		cannon.LoadELFMetaFlag,
	},
}
