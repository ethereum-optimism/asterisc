package disputegame

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"io"
	"math/big"
	"path/filepath"

	contractMetrics "github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/asterisc"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/split"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/utils"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/types"
	preimage "github.com/ethereum-optimism/optimism/op-preimage"

	"github.com/ethereum-optimism/asterisc/op-e2e/e2eutils/challenger"
	"github.com/ethereum-optimism/asterisc/rvgo/bindings"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/outputs"
	"github.com/ethereum-optimism/optimism/op-challenger/metrics"
	op_e2e_challenger "github.com/ethereum-optimism/optimism/op-e2e/e2eutils/challenger"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching/rpcblock"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	op_e2e_disputegame "github.com/ethereum-optimism/optimism/op-e2e/e2eutils/disputegame"
)

type OutputAsteriscGameHelper struct {
	op_e2e_disputegame.OutputCannonGameHelper
}

// StartChallenger overrides op_e2e_disputegame.OutputCannonGameHelper StartChallenger method
func (g *OutputAsteriscGameHelper) StartChallenger(
	ctx context.Context,
	name string,
	options ...op_e2e_challenger.Option,
) *op_e2e_challenger.Helper {
	opts := []op_e2e_challenger.Option{
		challenger.WithAsterisc(g.T, g.System.RollupCfg(), g.System.L2Genesis()),
		op_e2e_challenger.WithFactoryAddress(g.FactoryAddr),
		op_e2e_challenger.WithGameAddress(g.Addr),
	}
	opts = append(opts, options...)
	c := challenger.NewChallenger(g.T, ctx, g.System, name, opts...)
	g.T.Cleanup(func() {
		_ = c.Close()
	})
	return c
}

// CreateHonestActor overrides op_e2e_disputegame.OutputCannonGameHelper CreateHonestActor method
func (g *OutputAsteriscGameHelper) CreateHonestActor(ctx context.Context, l2Node string, options ...op_e2e_challenger.Option) *op_e2e_disputegame.OutputHonestHelper {
	opts := g.defaultChallengerOptions()
	opts = append(opts, options...)
	cfg := challenger.NewChallengerConfig(g.T, g.System, l2Node, opts...)

	logger := testlog.Logger(g.T, log.LevelInfo).New("role", "HonestHelper", "game", g.Addr)
	l2Client := g.System.NodeClient(l2Node)

	prestateBlock, poststateBlock, err := g.Game.GetBlockRange(ctx)
	g.Require.NoError(err, "Failed to load block range")
	dir := filepath.Join(cfg.Datadir, "honest")
	splitDepth := g.SplitDepth(ctx)
	rollupClient := g.System.RollupClient(l2Node)
	prestateProvider := outputs.NewPrestateProvider(rollupClient, prestateBlock)
	l1Head := g.GetL1Head(ctx)
	accessor, err := outputs.NewOutputAsteriscTraceAccessor(
		logger, metrics.NoopMetrics, cfg, l2Client, prestateProvider, cfg.AsteriscAbsolutePreState, rollupClient, dir, l1Head, splitDepth, prestateBlock, poststateBlock)
	g.Require.NoError(err, "Failed to create output asterisc trace accessor")
	return op_e2e_disputegame.NewOutputHonestHelper(g.T, g.Require, &g.OutputGameHelper, g.Game, accessor)
}

// CreateStepLargePreimageLoadCheck overrides op_e2e_disputegame.OutputCannonGameHelper CreateStepLargePreimageLoadCheck method
// Identical to OutputCannonGameHelper implementation
func (g *OutputAsteriscGameHelper) CreateStepLargePreimageLoadCheck(ctx context.Context, sender common.Address) op_e2e_disputegame.PreimageLoadCheck {
	return g.OutputCannonGameHelper.CreateStepLargePreimageLoadCheck(ctx, sender)
}

// CreateStepPreimageLoadCheck overrides op_e2e_disputegame.OutputCannonGameHelper CreateStepPreimageLoadCheck method
// Identical to OutputCannonGameHelper implementation
func (g *OutputAsteriscGameHelper) CreateStepPreimageLoadCheck(ctx context.Context) op_e2e_disputegame.PreimageLoadCheck {
	return g.OutputCannonGameHelper.CreateStepPreimageLoadCheck(ctx)
}

// ChallengeToPreimageLoad overrides op_e2e_disputegame.OutputCannonGameHelper ChallengeToPreimageLoad method
// ChallengeToPreimageLoad challenges the supplied execution root claim by inducing a step that requires a preimage to be loaded
// It does this by:
// 1. Identifying the first state transition that loads a global preimage
// 2. Descending the execution game tree to reach the step that loads the preimage
// 3. Asserting that the preimage was indeed loaded by an honest challenger (assuming the preimage is not preloaded)
// This expects an odd execution game depth in order for the honest challenger to step on our leaf claim
func (g *OutputAsteriscGameHelper) ChallengeToPreimageLoad(ctx context.Context, outputRootClaim *op_e2e_disputegame.ClaimHelper, challengerKey *ecdsa.PrivateKey, preimage utils.PreimageOpt, preimageCheck op_e2e_disputegame.PreimageLoadCheck, preloadPreimage bool) {
	// Identifying the first state transition that loads a global preimage
	provider, _ := g.createAsteriscTraceProvider(ctx, "sequencer", outputRootClaim, op_e2e_challenger.WithPrivKey(challengerKey))
	targetTraceIndex, err := provider.FindStep(ctx, 0, preimage)
	g.Require.NoError(err)

	splitDepth := g.SplitDepth(ctx)
	execDepth := g.ExecDepth(ctx)
	g.Require.NotEqual(outputRootClaim.Position.TraceIndex(execDepth).Uint64(), targetTraceIndex, "cannot move to defend a terminal trace index")
	g.Require.EqualValues(splitDepth+1, outputRootClaim.Depth(), "supplied claim must be the root of an execution game")
	g.Require.EqualValues(execDepth%2, 1, "execution game depth must be odd") // since we're challenging the execution root claim

	if preloadPreimage {
		_, _, preimageData, err := provider.GetStepData(ctx, types.NewPosition(execDepth, big.NewInt(int64(targetTraceIndex))))
		g.Require.NoError(err)
		g.UploadPreimage(ctx, preimageData, challengerKey)
		g.WaitForPreimageInOracle(ctx, preimageData)
	}

	// Descending the execution game tree to reach the step that loads the preimage
	bisectTraceIndex := func(claim *op_e2e_disputegame.ClaimHelper) *op_e2e_disputegame.ClaimHelper {
		execClaimPosition, err := claim.Position.RelativeToAncestorAtDepth(splitDepth + 1)
		g.Require.NoError(err)

		claimTraceIndex := execClaimPosition.TraceIndex(execDepth).Uint64()
		g.T.Logf("Bisecting: Into targetTraceIndex %v: claimIndex=%v at depth=%v. claimPosition=%v execClaimPosition=%v claimTraceIndex=%v",
			targetTraceIndex, claim.Index, claim.Depth(), claim.Position, execClaimPosition, claimTraceIndex)

		// We always want to position ourselves such that the challenger generates proofs for the targetTraceIndex as prestate
		if execClaimPosition.Depth() == execDepth-1 {
			if execClaimPosition.TraceIndex(execDepth).Uint64() == targetTraceIndex {
				newPosition := execClaimPosition.Attack()
				correct, err := provider.Get(ctx, newPosition)
				g.Require.NoError(err)
				g.T.Logf("Bisecting: Attack correctly for step at newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				return claim.Attack(ctx, correct)
			} else if execClaimPosition.TraceIndex(execDepth).Uint64() > targetTraceIndex {
				g.T.Logf("Bisecting: Attack incorrectly for step")
				return claim.Attack(ctx, common.Hash{0xdd})
			} else if execClaimPosition.TraceIndex(execDepth).Uint64()+1 == targetTraceIndex {
				g.T.Logf("Bisecting: Defend incorrectly for step")
				return claim.Defend(ctx, common.Hash{0xcc})
			} else {
				newPosition := execClaimPosition.Defend()
				correct, err := provider.Get(ctx, newPosition)
				g.Require.NoError(err)
				g.T.Logf("Bisecting: Defend correctly for step at newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				return claim.Defend(ctx, correct)
			}
		}

		// Attack or Defend depending on whether the claim we're responding to is to the left or right of the trace index
		// Induce the honest challenger to attack or defend depending on whether our new position will be to the left or right of the trace index
		if execClaimPosition.TraceIndex(execDepth).Uint64() < targetTraceIndex && claim.Depth() != splitDepth+1 {
			newPosition := execClaimPosition.Defend()
			if newPosition.TraceIndex(execDepth).Uint64() < targetTraceIndex {
				g.T.Logf("Bisecting: Defend correct. newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				correct, err := provider.Get(ctx, newPosition)
				g.Require.NoError(err)
				return claim.Defend(ctx, correct)
			} else {
				g.T.Logf("Bisecting: Defend incorrect. newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				return claim.Defend(ctx, common.Hash{0xaa})
			}
		} else {
			newPosition := execClaimPosition.Attack()
			if newPosition.TraceIndex(execDepth).Uint64() < targetTraceIndex {
				g.T.Logf("Bisecting: Attack correct. newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				correct, err := provider.Get(ctx, newPosition)
				g.Require.NoError(err)
				return claim.Attack(ctx, correct)
			} else {
				g.T.Logf("Bisecting: Attack incorrect. newPosition=%v execIndexAtDepth=%v", newPosition, newPosition.TraceIndex(execDepth))
				return claim.Attack(ctx, common.Hash{0xbb})
			}
		}
	}

	g.LogGameData(ctx)
	// Initial bisect to put us on defense
	mover := bisectTraceIndex(outputRootClaim)
	leafClaim := g.DefendClaim(ctx, mover, bisectTraceIndex, op_e2e_disputegame.WithoutWaitingForStep())

	// Validate that the preimage was loaded correctly
	g.Require.NoError(preimageCheck(provider, targetTraceIndex))

	// Now the preimage is available wait for the step call to succeed.
	leafClaim.WaitForCountered(ctx)
	g.LogGameData(ctx)
}

// VerifyPreimage overrides op_e2e_disputegame.OutputCannonGameHelper VerifyPreimage method
func (g *OutputAsteriscGameHelper) VerifyPreimage(ctx context.Context, outputRootClaim *op_e2e_disputegame.ClaimHelper, preimageKey preimage.Key) {
	execDepth := g.ExecDepth(ctx)

	// Identifying the first state transition that loads a global preimage
	provider, localContext := g.createAsteriscTraceProvider(ctx, "sequencer", outputRootClaim, op_e2e_challenger.WithPrivKey(op_e2e_disputegame.TestKey))
	start := uint64(0)
	found := false
	for offset := uint32(0); ; offset += 4 {
		preimageOpt := utils.PreimageLoad(preimageKey, offset)
		g.T.Logf("Searching for step with key %x and offset %v", preimageKey.PreimageKey(), offset)
		targetTraceIndex, err := provider.FindStep(ctx, start, preimageOpt)
		if errors.Is(err, io.EOF) {
			// Did not find any more reads
			g.Require.True(found, "Should have found at least one preimage read")
			g.T.Logf("Searching for step with key %x and offset %v did not find another read", preimageKey.PreimageKey(), offset)
			return
		}
		g.Require.NoError(err, "Failed to find step that loads requested preimage")
		start = targetTraceIndex
		found = true

		g.T.Logf("Target trace index: %v", targetTraceIndex)
		pos := types.NewPosition(execDepth, new(big.Int).SetUint64(targetTraceIndex))
		g.Require.Equal(targetTraceIndex, pos.TraceIndex(execDepth).Uint64())

		prestate, proof, oracleData, err := provider.GetStepData(ctx, pos)
		g.Require.NoError(err, "Failed to get step data")
		g.Require.NotNil(oracleData, "Should have had required preimage oracle data")
		g.Require.Equal(common.Hash(preimageKey.PreimageKey()).Bytes(), oracleData.OracleKey, "Must have correct preimage key")

		tx, err := g.GameBindings.AddLocalData(g.Opts,
			oracleData.GetIdent(),
			big.NewInt(outputRootClaim.Index),
			new(big.Int).SetUint64(uint64(oracleData.OracleOffset)))
		g.Require.NoError(err)
		_, err = wait.ForReceiptOK(ctx, g.Client, tx.Hash())
		g.Require.NoError(err)

		expectedPostState, err := provider.Get(ctx, pos)
		g.Require.NoError(err, "Failed to get expected post state")

		callOpts := &bind.CallOpts{Context: ctx}
		vmAddr, err := g.GameBindings.Vm(callOpts)
		g.Require.NoError(err, "Failed to get VM address")

		abi, err := bindings.RISCVMetaData.GetAbi()
		g.Require.NoError(err, "Failed to load RISCV ABI")
		caller := batching.NewMultiCaller(g.Client.Client(), batching.DefaultBatchSize)
		result, err := caller.SingleCall(ctx, rpcblock.Latest, &batching.ContractCall{
			Abi:    abi,
			Addr:   vmAddr,
			Method: "step",
			Args: []interface{}{
				prestate, proof, localContext,
			},
			From: g.Addr,
		})
		g.Require.NoError(err, "Failed to call step")
		actualPostState := result.GetBytes32(0)
		g.Require.Equal(expectedPostState, common.Hash(actualPostState))
	}
}

func (g *OutputAsteriscGameHelper) createAsteriscTraceProvider(ctx context.Context, l2Node string, outputRootClaim *op_e2e_disputegame.ClaimHelper, options ...op_e2e_challenger.Option) (*asterisc.AsteriscTraceProviderForTest, common.Hash) {
	splitDepth := g.SplitDepth(ctx)
	g.Require.EqualValues(outputRootClaim.Depth(), splitDepth+1, "outputRootClaim must be the root of an execution game")

	logger := testlog.Logger(g.T, log.LevelInfo).New("role", "AsteriscTraceProvider", "game", g.Addr)
	opt := g.defaultChallengerOptions()
	opt = append(opt, options...)
	cfg := challenger.NewChallengerConfig(g.T, g.System, l2Node, opt...)

	caller := batching.NewMultiCaller(g.System.NodeClient("l1").Client(), batching.DefaultBatchSize)
	l2Client := g.System.NodeClient(l2Node)
	contract, err := contracts.NewFaultDisputeGameContract(ctx, contractMetrics.NoopContractMetrics, g.Addr, caller)
	g.Require.NoError(err)

	prestateBlock, poststateBlock, err := contract.GetBlockRange(ctx)
	g.Require.NoError(err, "Failed to load block range")
	rollupClient := g.System.RollupClient(l2Node)
	prestateProvider := outputs.NewPrestateProvider(rollupClient, prestateBlock)
	l1Head := g.GetL1Head(ctx)
	outputProvider := outputs.NewTraceProvider(logger, prestateProvider, rollupClient, l2Client, l1Head, splitDepth, prestateBlock, poststateBlock)

	var localContext common.Hash
	selector := split.NewSplitProviderSelector(outputProvider, splitDepth, func(ctx context.Context, depth types.Depth, pre types.Claim, post types.Claim) (types.TraceProvider, error) {
		agreed, disputed, err := outputs.FetchProposals(ctx, outputProvider, pre, post)
		g.Require.NoError(err)
		g.T.Logf("Using trace between blocks %v and %v\n", agreed.L2BlockNumber, disputed.L2BlockNumber)
		localInputs, err := utils.FetchLocalInputsFromProposals(ctx, l1Head.Hash, l2Client, agreed, disputed)
		g.Require.NoError(err, "Failed to fetch local inputs")
		localContext = outputs.CreateLocalContext(pre, post)
		dir := filepath.Join(cfg.Datadir, "asterisc-trace")
		subdir := filepath.Join(dir, localContext.Hex())
		return asterisc.NewTraceProviderForTest(logger, metrics.NoopMetrics, cfg, localInputs, subdir, g.MaxDepth(ctx)-splitDepth-1), nil
	})

	claims, err := contract.GetAllClaims(ctx, rpcblock.Latest)
	g.Require.NoError(err)
	game := types.NewGameState(claims, g.MaxDepth(ctx))

	provider, err := selector(ctx, game, game.Claims()[outputRootClaim.ParentIndex], outputRootClaim.Position)
	g.Require.NoError(err)
	translatingProvider := provider.(*trace.TranslatingProvider)
	return translatingProvider.Original().(*asterisc.AsteriscTraceProviderForTest), localContext
}

func (g *OutputAsteriscGameHelper) defaultChallengerOptions() []op_e2e_challenger.Option {
	return []op_e2e_challenger.Option{
		challenger.WithAsterisc(g.T, g.System.RollupCfg(), g.System.L2Genesis()),
		op_e2e_challenger.WithFactoryAddress(g.FactoryAddr),
		op_e2e_challenger.WithGameAddress(g.Addr),
	}
}
