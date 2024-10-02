package disputegame

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum-optimism/asterisc/op-e2e/e2eutils/challenger"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/contracts/metrics"
	"github.com/ethereum-optimism/optimism/op-challenger/game/fault/trace/outputs"
	op_e2e_challenger "github.com/ethereum-optimism/optimism/op-e2e/e2eutils/challenger"
	op_e2e_disputegame "github.com/ethereum-optimism/optimism/op-e2e/e2eutils/disputegame"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/disputegame/preimage"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/transactions"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const (
	asteriscGameType uint32 = 2
)

type AsteriscFactoryHelper struct {
	op_e2e_disputegame.FactoryHelper
}

func NewAsteriscFactoryHelper(t *testing.T, ctx context.Context, system op_e2e_disputegame.DisputeSystem) *AsteriscFactoryHelper {
	factoryHelper := op_e2e_disputegame.NewFactoryHelper(t, ctx, system)
	return &AsteriscFactoryHelper{
		*factoryHelper,
	}
}

// PreimageHelper overrides op_e2e_disputegame.FactoryHelper PreimageHelper method
func (h *AsteriscFactoryHelper) PreimageHelper(ctx context.Context) *preimage.Helper {
	opts := &bind.CallOpts{Context: ctx}
	gameAddr, err := h.Factory.GameImpls(opts, asteriscGameType)
	h.Require.NoError(err)
	caller := batching.NewMultiCaller(h.Client.Client(), batching.DefaultBatchSize)
	game, err := contracts.NewFaultDisputeGameContract(ctx, metrics.NoopContractMetrics, gameAddr, caller)
	h.Require.NoError(err)
	vm, err := game.Vm(ctx)
	h.Require.NoError(err)
	oracle, err := vm.Oracle(ctx)
	h.Require.NoError(err)
	return preimage.NewHelper(h.T, h.PrivKey, h.Client, oracle)
}

func (h *AsteriscFactoryHelper) StartOutputAsteriscGameWithCorrectRoot(ctx context.Context, l2Node string, l2BlockNumber uint64, opts ...op_e2e_disputegame.GameOpt) *OutputAsteriscGameHelper {
	cfg := op_e2e_disputegame.NewGameCfg(opts...)
	h.WaitForBlock(l2Node, l2BlockNumber, cfg)
	output, err := h.System.RollupClient(l2Node).OutputAtBlock(ctx, l2BlockNumber)
	h.Require.NoErrorf(err, "Failed to get output at block %v", l2BlockNumber)
	return h.StartOutputAsteriscGame(ctx, l2Node, l2BlockNumber, common.Hash(output.OutputRoot), opts...)
}

func (h *AsteriscFactoryHelper) StartOutputAsteriscGame(ctx context.Context, l2Node string, l2BlockNumber uint64, rootClaim common.Hash, opts ...op_e2e_disputegame.GameOpt) *OutputAsteriscGameHelper {
	cfg := op_e2e_disputegame.NewGameCfg(opts...)
	logger := testlog.Logger(h.T, log.LevelInfo).New("role", "OutputAsteriscGameHelper")
	rollupClient := h.System.RollupClient(l2Node)
	l2Client := h.System.NodeClient(l2Node)

	extraData := h.CreateBisectionGameExtraData(l2Node, l2BlockNumber, cfg)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	tx, err := transactions.PadGasEstimate(h.Opts, 2, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return h.Factory.Create(opts, asteriscGameType, rootClaim, extraData)
	})
	h.Require.NoError(err, "create fault dispute game")
	rcpt, err := wait.ForReceiptOK(ctx, h.Client, tx.Hash())
	h.Require.NoError(err, "wait for create fault dispute game receipt to be OK")
	h.Require.Len(rcpt.Logs, 2, "should have emitted a single DisputeGameCreated event")
	createdEvent, err := h.Factory.ParseDisputeGameCreated(*rcpt.Logs[1])
	h.Require.NoError(err)
	game, err := contracts.NewFaultDisputeGameContract(ctx, metrics.NoopContractMetrics, createdEvent.DisputeProxy, batching.NewMultiCaller(h.Client.Client(), batching.DefaultBatchSize))
	h.Require.NoError(err)

	prestateBlock, poststateBlock, err := game.GetBlockRange(ctx)
	h.Require.NoError(err, "Failed to load starting block number")
	splitDepth, err := game.GetSplitDepth(ctx)
	h.Require.NoError(err, "Failed to load split depth")
	l1Head := h.GetL1Head(ctx, game)

	prestateProvider := outputs.NewPrestateProvider(rollupClient, prestateBlock)
	provider := outputs.NewTraceProvider(logger, prestateProvider, rollupClient, l2Client, l1Head, splitDepth, prestateBlock, poststateBlock)

	return &OutputAsteriscGameHelper{
		OutputCannonGameHelper: op_e2e_disputegame.OutputCannonGameHelper{
			OutputGameHelper: *op_e2e_disputegame.NewOutputGameHelper(h.T, h.Require, h.Client, h.Opts, h.PrivKey, game, h.FactoryAddr, createdEvent.DisputeProxy, provider, h.System, h.AllocType),
		},
	}
}

// GetL1Head overrides op_e2e_disputegame.FactoryHelper GetL1Head method
// Identical to FactoryHelper implementation
func (h *AsteriscFactoryHelper) GetL1Head(ctx context.Context, game contracts.FaultDisputeGameContract) eth.BlockID {
	return h.FactoryHelper.GetL1Head(ctx, game)
}

// CreateBisectionGameExtraData overrides op_e2e_disputegame.FactoryHelper CreateBisectionGameExtraData method
// Identical to FactoryHelper implementation
func (h *AsteriscFactoryHelper) CreateBisectionGameExtraData(l2Node string, l2BlockNumber uint64, cfg *op_e2e_disputegame.GameCfg) []byte {
	return h.FactoryHelper.CreateBisectionGameExtraData(l2Node, l2BlockNumber, cfg)
}

// WaitForBlock overrides op_e2e_disputegame.FactoryHelper WaitForBlock method
// Identical to FactoryHelper implementation
func (h *AsteriscFactoryHelper) WaitForBlock(l2Node string, l2BlockNumber uint64, cfg *op_e2e_disputegame.GameCfg) {
	h.FactoryHelper.WaitForBlock(l2Node, l2BlockNumber, cfg)
}

// StartChallenger overrides op_e2e_disputegame.FactoryHelper StartChallenger method
func (h *AsteriscFactoryHelper) StartChallenger(ctx context.Context, name string, options ...op_e2e_challenger.Option) *op_e2e_challenger.Helper {
	opts := []op_e2e_challenger.Option{
		op_e2e_challenger.WithFactoryAddress(h.FactoryAddr),
	}
	opts = append(opts, options...)
	c := challenger.NewChallenger(h.T, ctx, h.System, name, opts...)
	h.T.Cleanup(func() {
		_ = c.Close()
	})
	return c
}
