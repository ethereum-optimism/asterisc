package cmd

import (
	"debug/elf"
	"fmt"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

var (
	LoadELFPathFlag = &cli.PathFlag{
		Name:      "path",
		Usage:     "Path to RISC-V ELF file",
		TakesFile: true,
		Required:  true,
	}
	LoadELFOutFlag = &cli.PathFlag{
		Name:     "out",
		Usage:    "Output path to write JSON state to. State is dumped to stdout if set to -. Not written if empty.",
		Value:    "state.json",
		Required: false,
	}
)

func LoadELF(ctx *cli.Context) error {
	elfPath := ctx.Path(LoadELFPathFlag.Name)
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
	return writeJSON[*fast.VMState](ctx.Path(LoadELFOutFlag.Name), state)
}

var LoadELFCommand = &cli.Command{
	Name:        "load-elf",
	Usage:       "Load ELF file into Cannon JSON state",
	Description: "Load ELF file into Cannon JSON state, optionally patch out functions",
	Action:      LoadELF,
	Flags: []cli.Flag{
		LoadELFPathFlag,
		LoadELFOutFlag,
	},
}
