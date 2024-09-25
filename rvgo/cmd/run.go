package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/profile"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	cannon "github.com/ethereum-optimism/optimism/cannon/cmd"
	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// Guard checks if the step is failed due to pre-image server error
func Guard(proc *os.ProcessState, fn StepFn) StepFn {
	return func(proof bool) (*fast.StepWitness, error) {
		wit, err := fn(proof)
		if err != nil {
			// proc is not nil when the preimage server process is terminated
			if proc != nil && proc.Exited() {
				return nil, fmt.Errorf("pre-image server exited with code %d, resulting in err %w", proc.ExitCode(), err)
			} else {
				return nil, err
			}
		}
		return wit, nil
	}
}

var _ fast.PreimageOracle = (*ProcessPreimageOracle)(nil)

var OutFilePerm = os.FileMode(0o755)

func Run(ctx *cli.Context) error {
	if ctx.Bool(cannon.RunPProfCPU.Name) {
		defer profile.Start(profile.NoShutdownHook, profile.ProfilePath("."), profile.CPUProfile).Stop()
	}

	state, err := jsonutil.LoadJSON[fast.VMState](ctx.Path(cannon.RunInputFlag.Name))
	if err != nil {
		return err
	}
	l := Logger(os.Stderr, slog.LevelInfo)
	outLog := &LoggingWriter{Name: "program std-out", Log: l}
	errLog := &LoggingWriter{Name: "program std-err", Log: l}

	stopAtAnyPreimage := false
	var stopAtPreimageKeyPrefix []byte
	stopAtPreimageOffset := uint64(0)
	if ctx.IsSet(cannon.RunStopAtPreimageFlag.Name) {
		val := ctx.String(cannon.RunStopAtPreimageFlag.Name)
		parts := strings.Split(val, "@")
		if len(parts) > 2 {
			return fmt.Errorf("invalid %v: %v", cannon.RunStopAtPreimageFlag.Name, val)
		}
		stopAtPreimageKeyPrefix = common.FromHex(parts[0])
		if len(parts) == 2 {
			x, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid preimage offset: %w", err)
			}
			stopAtPreimageOffset = x
		}
	} else {
		switch ctx.String(cannon.RunStopAtPreimageTypeFlag.Name) {
		case "local":
			stopAtPreimageKeyPrefix = []byte{byte(preimage.LocalKeyType)}
		case "keccak":
			stopAtPreimageKeyPrefix = []byte{byte(preimage.Keccak256KeyType)}
		case "sha256":
			stopAtPreimageKeyPrefix = []byte{byte(preimage.Sha256KeyType)}
		case "blob":
			stopAtPreimageKeyPrefix = []byte{byte(preimage.BlobKeyType)}
		case "precompile":
			stopAtPreimageKeyPrefix = []byte{byte(preimage.PrecompileKeyType)}
		case "any":
			stopAtAnyPreimage = true
		case "":
			// 0 preimage type is forbidden so will not stop at any preimage
		default:
			return fmt.Errorf("invalid preimage type %q", ctx.String(cannon.RunStopAtPreimageTypeFlag.Name))
		}
	}
	stopAtPreimageLargerThan := ctx.Int(cannon.RunStopAtPreimageLargerThanFlag.Name)

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

	po, err := NewProcessPreimageOracle(args[0], args[1:])
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

	stopAt := ctx.Generic(cannon.RunStopAtFlag.Name).(*cannon.StepMatcherFlag).Matcher()
	proofAt := ctx.Generic(cannon.RunProofAtFlag.Name).(*cannon.StepMatcherFlag).Matcher()
	snapshotAt := ctx.Generic(cannon.RunSnapshotAtFlag.Name).(*cannon.StepMatcherFlag).Matcher()
	infoAt := ctx.Generic(cannon.RunInfoAtFlag.Name).(*cannon.StepMatcherFlag).Matcher()

	var meta *Metadata
	if metaPath := ctx.Path(cannon.RunMetaFlag.Name); metaPath == "" {
		l.Info("no metadata file specified, defaulting to empty metadata")
		meta = &Metadata{Symbols: nil} // provide empty metadata by default
	} else {
		if m, err := jsonutil.LoadJSON[Metadata](metaPath); err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		} else {
			meta = m
		}
	}

	us := fast.NewInstrumentedState(state, po, outLog, errLog)
	proofFmt := ctx.String(cannon.RunProofFmtFlag.Name)
	snapshotFmt := ctx.String(cannon.RunSnapshotFmtFlag.Name)

	stepFn := us.Step
	if po.cmd != nil {
		stepFn = Guard(po.cmd.ProcessState, stepFn)
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
				"pc", HexU32(state.PC),
				"insn", HexU32(state.Instr()),
				"ips", float64(step-startStep)/(float64(delta)/float64(time.Second)),
				"pages", state.Memory.PageCount(),
				"mem", state.Memory.Usage(),
				"name", meta.LookupSymbol(state.PC),
			)
		}

		if stopAt(state) {
			break
		}

		if snapshotAt(state) {
			if err := jsonutil.WriteJSON(state, ioutil.ToStdOutOrFileOrNoop(fmt.Sprintf(snapshotFmt, step), OutFilePerm)); err != nil {
				return fmt.Errorf("failed to write state snapshot: %w", err)
			}
		}

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
			if err := jsonutil.WriteJSON(proof, ioutil.ToStdOutOrFileOrNoop(fmt.Sprintf(proofFmt, step), OutFilePerm)); err != nil {
				return fmt.Errorf("failed to write proof data: %w", err)
			}
		} else {
			_, err = stepFn(false)
			if err != nil {
				return fmt.Errorf("failed at step %d (PC: %08x): %w", step, state.PC, err)
			}
		}

		lastPreimageKey, lastPreimageValue, lastPreimageOffset := us.LastPreimage()
		if lastPreimageOffset != ^uint64(0) {
			if stopAtAnyPreimage {
				l.Info("Stopping at preimage read")
				break
			}
			if len(stopAtPreimageKeyPrefix) > 0 &&
				slices.Equal(lastPreimageKey[:len(stopAtPreimageKeyPrefix)], stopAtPreimageKeyPrefix) {
				if stopAtPreimageOffset == lastPreimageOffset {
					l.Info("Stopping at preimage read", "keyPrefix", common.Bytes2Hex(stopAtPreimageKeyPrefix), "offset", lastPreimageOffset)
					break
				}
			}
			if stopAtPreimageLargerThan != 0 && len(lastPreimageValue) > stopAtPreimageLargerThan {
				l.Info("Stopping at preimage read", "size", len(lastPreimageValue), "min", stopAtPreimageLargerThan)
				break
			}
		}
	}

	if err := state.SetWitnessAndStateHash(); err != nil {
		return fmt.Errorf("failed to set witness and stateHash: %w", err)
	}

	if err := jsonutil.WriteJSON(state, ioutil.ToStdOutOrFileOrNoop(ctx.Path(cannon.RunOutputFlag.Name), OutFilePerm)); err != nil {
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
		cannon.RunInputFlag,
		cannon.RunOutputFlag,
		cannon.RunProofAtFlag,
		cannon.RunProofFmtFlag,
		cannon.RunSnapshotAtFlag,
		cannon.RunSnapshotFmtFlag,
		cannon.RunStopAtFlag,
		cannon.RunStopAtPreimageFlag,
		cannon.RunStopAtPreimageTypeFlag,
		cannon.RunStopAtPreimageLargerThanFlag,
		cannon.RunMetaFlag,
		cannon.RunInfoAtFlag,
		cannon.RunPProfCPU,
	},
}
