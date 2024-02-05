package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	"github.com/pkg/profile"

	"github.com/ethereum-optimism/optimism/cannon/cmd"
	preimage "github.com/ethereum-optimism/optimism/op-preimage"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

var (
	RunInputFlag = &cli.PathFlag{
		Name:      "input",
		Usage:     "path of input JSON state. Stdin if left empty.",
		TakesFile: true,
		Value:     "state.json",
		Required:  true,
	}
	RunOutputFlag = &cli.PathFlag{
		Name:      "output",
		Usage:     "path of output JSON state. Not written if empty, use - to write to Stdout.",
		TakesFile: true,
		Value:     "out.json",
		Required:  false,
	}
	patternHelp    = "'never' (default), 'always', '=123' at exactly step 123, '%123' for every 123 steps"
	RunProofAtFlag = &cli.GenericFlag{
		Name:     "proof-at",
		Usage:    "step pattern to output proof at: " + patternHelp,
		Value:    new(cmd.StepMatcherFlag),
		Required: false,
	}
	RunProofFmtFlag = &cli.StringFlag{
		Name:     "proof-fmt",
		Usage:    "format for proof data output file names. Proof data is written to stdout if -.",
		Value:    "proof-%d.json",
		Required: false,
	}
	RunSnapshotAtFlag = &cli.GenericFlag{
		Name:     "snapshot-at",
		Usage:    "step pattern to output snapshots at: " + patternHelp,
		Value:    new(cmd.StepMatcherFlag),
		Required: false,
	}
	RunSnapshotFmtFlag = &cli.StringFlag{
		Name:     "snapshot-fmt",
		Usage:    "format for snapshot output file names.",
		Value:    "state-%d.json",
		Required: false,
	}
	RunStopAtFlag = &cli.GenericFlag{
		Name:     "stop-at",
		Usage:    "step pattern to stop at: " + patternHelp,
		Value:    new(cmd.StepMatcherFlag),
		Required: false,
	}
	RunStopAtPreimageTypeFlag = &cli.StringFlag{
		Name:     "stop-at-preimage-type",
		Usage:    "stop at the first preimage request matching this type (must be either 'any', 'local' or 'global')",
		Required: false,
	}
	RunStopAtPreimageKeyFlag = &cli.StringFlag{
		Name:     "stop-at-preimage-key",
		Usage:    "stop at the first step that requests the specified preimage key",
		Required: false,
	}
	RunMetaFlag = &cli.PathFlag{
		Name:     "meta",
		Usage:    "path to metadata file for symbol lookup for enhanced debugging info during execution.",
		Value:    "meta.json",
		Required: false,
	}
	RunInfoAtFlag = &cli.GenericFlag{
		Name:     "info-at",
		Usage:    "step pattern to print info at: " + patternHelp,
		Value:    cmd.MustStepMatcherFlag("%100000"),
		Required: false,
	}
	RunPProfCPU = &cli.BoolFlag{
		Name:  "pprof.cpu",
		Usage: "enable pprof cpu profiling",
	}
)

type Proof struct {
	Step uint64 `json:"step"`

	Pre  common.Hash `json:"pre"`
	Post common.Hash `json:"post"`

	StateData hexutil.Bytes `json:"state-data"`
	ProofData hexutil.Bytes `json:"proof-data"`

	OracleKey    hexutil.Bytes `json:"oracle-key,omitempty"`
	OracleValue  hexutil.Bytes `json:"oracle-value,omitempty"`
	OracleOffset uint64        `json:"oracle-offset,omitempty"`
}

type StepFn func(proof bool) (*fast.StepWitness, error)

func Guard(proc *os.ProcessState, fn StepFn) StepFn {
	return func(proof bool) (*fast.StepWitness, error) {
		wit, err := fn(proof)
		if err != nil {
			if proc.Exited() {
				return nil, fmt.Errorf("pre-image server exited with code %d, resulting in err %w", proc.ExitCode(), err)
			} else {
				return nil, err
			}
		}
		return wit, nil
	}
}

var _ fast.PreimageOracle = (*cmd.ProcessPreimageOracle)(nil)

func Run(ctx *cli.Context) error {
	if ctx.Bool(RunPProfCPU.Name) {
		defer profile.Start(profile.NoShutdownHook, profile.ProfilePath("."), profile.CPUProfile).Stop()
	}

	state, err := cmd.LoadJSON[fast.VMState](ctx.Path(RunInputFlag.Name))
	if err != nil {
		return err
	}

	l := cmd.Logger(os.Stderr, log.LvlInfo)
	outLog := &cmd.LoggingWriter{Name: "program std-out", Log: l}
	errLog := &cmd.LoggingWriter{Name: "program std-err", Log: l}

	stopAtPreimageType := ctx.String(RunStopAtPreimageTypeFlag.Name)
	if stopAtPreimageType != "" && stopAtPreimageType != "any" && stopAtPreimageType != "local" && stopAtPreimageType != "global" {
		return fmt.Errorf("invalid preimage type %q, must be either 'any', 'local' or 'global'", stopAtPreimageType)
	}
	stopAtPreimageKey := common.HexToHash(ctx.String(RunStopAtPreimageKeyFlag.Name))

	// split CLI args after first '--'
	args := ctx.Args().Slice()
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		args = []string{""}
	}

	po, err := cmd.NewProcessPreimageOracle(args[0], args[1:])
	if err != nil {
		return fmt.Errorf("failed to create pre-image oracle process: %w", err)
	}
	if err := po.Start(); err != nil {
		return fmt.Errorf("failed to start pre-image oracle server: %w", err)
	}
	defer func() {
		if err := po.Close(); err != nil {
			l.Error("failed to close pre-image server", "err", err)
		}
	}()

	stopAt := ctx.Generic(RunStopAtFlag.Name).(*cmd.StepMatcherFlag).Matcher()
	proofAt := ctx.Generic(RunProofAtFlag.Name).(*cmd.StepMatcherFlag).Matcher()
	snapshotAt := ctx.Generic(RunSnapshotAtFlag.Name).(*cmd.StepMatcherFlag).Matcher()
	infoAt := ctx.Generic(RunInfoAtFlag.Name).(*cmd.StepMatcherFlag).Matcher()

	us := fast.NewInstrumentedState(state, po, outLog, errLog)
	proofFmt := ctx.String(RunProofFmtFlag.Name)
	snapshotFmt := ctx.String(RunSnapshotFmtFlag.Name)

	stepFn := us.Step
	poCmd := po.GetCmd()
	if poCmd != nil {
		stepFn = Guard(poCmd.ProcessState, stepFn)
	}

	start := time.Now()
	startStep := state.Step

	for !state.Exited {
		if state.Step%100 == 0 { // don't do the ctx err check (includes lock) too often
			if err := ctx.Context.Err(); err != nil {
				return err
			}
		}

		step := state.Step

		if infoAt(state) {
			delta := time.Since(start)
			l.Info("processing",
				"step", step,
				"pc", cmd.HexU32(state.PC),
				"insn", cmd.HexU32(state.Instr()),
				"ips", float64(step-startStep)/(float64(delta)/float64(time.Second)),
				"pages", state.Memory.PageCount(),
				"mem", state.Memory.Usage(),
			)
		}

		if stopAt(state) {
			break
		}

		if snapshotAt(state) {
			if err := cmd.WriteJSON(fmt.Sprintf(snapshotFmt, step), state); err != nil {
				return fmt.Errorf("failed to write state snapshot: %w", err)
			}
		}

		prevPreimageOffset := state.PreimageOffset

		if proofAt(state) {
			preStateHash, err := state.EncodeWitness().StateHash()
			if err != nil {
				return fmt.Errorf("failed to hash prestate witness: %w", err)
			}
			witness, err := stepFn(true)
			if err != nil {
				return fmt.Errorf("failed at proof-gen step %d (PC: %08x): %w", step, state.PC, err)
			}
			postStateHash, err := state.EncodeWitness().StateHash()
			if err != nil {
				return fmt.Errorf("failed to hash poststate witness: %w", err)
			}
			proof := &Proof{
				Step:      step,
				Pre:       preStateHash,
				Post:      postStateHash,
				StateData: witness.State,
				ProofData: witness.MemProof,
			}
			if witness.HasPreimage() {
				proof.OracleKey = witness.PreimageKey[:]
				proof.OracleValue = witness.PreimageValue
				proof.OracleOffset = witness.PreimageOffset
			}
			if err := cmd.WriteJSON(fmt.Sprintf(proofFmt, step), proof); err != nil {
				return fmt.Errorf("failed to write proof data: %w", err)
			}
		} else {
			_, err = stepFn(false)
			if err != nil {
				return fmt.Errorf("failed at step %d (PC: %08x): %w", step, state.PC, err)
			}
		}

		if preimageRead := state.PreimageOffset > prevPreimageOffset; preimageRead {
			if stopAtPreimageType == "any" {
				break
			}
			if stopAtPreimageType != "" {
				keyType := byte(preimage.LocalKeyType)
				if stopAtPreimageType == "global" {
					keyType = byte(preimage.Keccak256KeyType)
				}
				if state.PreimageKey[0] == keyType {
					break
				}
			}
			if (stopAtPreimageKey != common.Hash{}) && state.PreimageKey == stopAtPreimageKey {
				break
			}
		}
	}

	if err := cmd.WriteJSON(ctx.Path(RunOutputFlag.Name), state); err != nil {
		return fmt.Errorf("failed to write state output: %w", err)
	}
	return nil
}

var RunCommand = &cli.Command{
	Name:        "run",
	Usage:       "Run VM step(s) and generate proof data to replicate onchain.",
	Description: "Run VM step(s) and generate proof data to replicate onchain. See flags to match when to output a proof, a snapshot, or to stop early.",
	Action:      Run,
	Flags: []cli.Flag{
		RunInputFlag,
		RunOutputFlag,
		RunProofAtFlag,
		RunProofFmtFlag,
		RunSnapshotAtFlag,
		RunSnapshotFmtFlag,
		RunStopAtFlag,
		RunStopAtPreimageTypeFlag,
		RunStopAtPreimageKeyFlag,
		RunMetaFlag,
		RunInfoAtFlag,
		RunPProfCPU,
	},
}
