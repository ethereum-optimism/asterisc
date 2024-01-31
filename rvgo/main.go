package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum-optimism/asterisc/rvgo/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "asterisc"
	app.Usage = "RISC-V Fault Proof tool"
	app.Description = "RISC-V Fault Proof tool"
	app.Commands = []*cli.Command{
		cmd.LoadELFCommand,
		cmd.WitnessCommand,
	}
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-c
			cancel()
			fmt.Println("\r\nExiting...")
		}
	}()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		if errors.Is(err, ctx.Err()) {
			_, _ = fmt.Fprintf(os.Stderr, "command interrupted")
			os.Exit(130)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
			os.Exit(1)
		}
	}
}
