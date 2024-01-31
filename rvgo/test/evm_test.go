package test

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-chain-ops/srcmap"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

type dummyChain struct {
}

// Engine retrieves the chain's consensus engine.
func (d *dummyChain) Engine() consensus.Engine {
	return ethash.NewFullFaker()
}

// GetHeader returns the hash corresponding to their hash.
func (d *dummyChain) GetHeader(h common.Hash, n uint64) *types.Header {
	parentHash := common.Hash{0: 0xff}
	binary.BigEndian.PutUint64(parentHash[1:], n-1)
	return fakeHeader(n, parentHash)
}

func fakeHeader(n uint64, parentHash common.Hash) *types.Header {
	header := types.Header{
		Coinbase:   common.HexToAddress("0x00000000000000000000000000000000deadbeef"),
		Number:     big.NewInt(int64(n)),
		ParentHash: parentHash,
		Time:       1000,
		Nonce:      types.BlockNonce{0x1},
		Extra:      []byte{},
		Difficulty: big.NewInt(0),
		GasLimit:   100000,
	}
	return &header
}

func loadStepContractCode(t *testing.T) *Contract {
	dat, err := os.ReadFile("../../rvsol/out/Step.sol/Step.json")
	require.NoError(t, err)
	var outDat Contract
	err = json.Unmarshal(dat, &outDat)
	require.NoError(t, err)
	return &outDat
}

func loadPreimageOracleContractCode(t *testing.T) *Contract {
	dat, err := os.ReadFile("../../rvsol/out/PreimageOracle.sol/PreimageOracle.json")
	require.NoError(t, err)
	var outDat Contract
	err = json.Unmarshal(dat, &outDat)
	require.NoError(t, err)
	return &outDat
}

type Contract struct {
	DeployedBytecode struct {
		Object    hexutil.Bytes `json:"object"`
		SourceMap string        `json:"sourceMap"`
	} `json:"deployedBytecode"`
}

func (c *Contract) SourceMap(sourcePaths []string) (*srcmap.SourceMap, error) {
	return srcmap.ParseSourceMap(sourcePaths, c.DeployedBytecode.Object, c.DeployedBytecode.SourceMap)
}

type Contracts struct {
	RISCV  *Contract
	Oracle *Contract
}

type Addresses struct {
	RISCV        common.Address
	Oracle       common.Address
	Sender       common.Address
	FeeRecipient common.Address
}

func newEVMEnv(t *testing.T, contracts *Contracts, addrs *Addresses) *vm.EVM {
	chainCfg := params.MainnetChainConfig
	bc := &dummyChain{}
	header := bc.GetHeader(common.Hash{}, 100)
	db := rawdb.NewMemoryDatabase()
	statedb := state.NewDatabase(db)
	state, err := state.New(types.EmptyRootHash, statedb, nil)
	require.NoError(t, err)
	blockContext := core.NewEVMBlockContext(header, bc, nil, chainCfg, state)
	vmCfg := vm.Config{}

	env := vm.NewEVM(blockContext, vm.TxContext{}, state, chainCfg, vmCfg)
	env.StateDB.SetCode(addrs.RISCV, contracts.RISCV.DeployedBytecode.Object)
	env.StateDB.SetCode(addrs.Oracle, contracts.Oracle.DeployedBytecode.Object)
	env.StateDB.SetState(addrs.RISCV, common.Hash{}, common.BytesToHash(addrs.Oracle.Bytes())) // set storage slot pointing to preimage oracle

	rules := env.ChainConfig().Rules(header.Number, true, header.Time)
	env.StateDB.Prepare(rules, addrs.Sender, addrs.FeeRecipient, &addrs.RISCV, vm.ActivePrecompiles(rules), nil)
	return env
}

var testAddrs = &Addresses{
	RISCV:        common.HexToAddress("0x1337"),
	Oracle:       common.HexToAddress("0xf00d"),
	Sender:       common.HexToAddress("0x7070"),
	FeeRecipient: common.HexToAddress("0xbd69"),
}

func testContracts(t *testing.T) *Contracts {
	return &Contracts{
		RISCV:  loadStepContractCode(t),
		Oracle: loadPreimageOracleContractCode(t),
	}
}

func addTracer(t *testing.T, env *vm.EVM, addrs *Addresses, contracts *Contracts) {
	//env.Config.Tracer = logger.NewMarkdownLogger(&logger.Config{}, os.Stdout)

	a, err := contracts.RISCV.SourceMap([]string{"../../rvsol/src/Step.sol"})
	require.NoError(t, err)
	b, err := contracts.Oracle.SourceMap([]string{"../../rvsol/src/PreimageOracle.sol"})
	require.NoError(t, err)
	env.Config.Tracer = srcmap.NewSourceMapTracer(map[common.Address]*srcmap.SourceMap{
		addrs.RISCV:  a,
		addrs.Oracle: b,
	}, os.Stdout)
}

func stepEVM(t *testing.T, env *vm.EVM, wit *fast.StepWitness, addrs *Addresses, step uint64) (postState []byte, postHash common.Hash, gasUsed uint64) {
	startingGas := uint64(30_000_000)

	snap := env.StateDB.Snapshot()

	if wit.HasPreimage() {
		input, err := wit.EncodePreimageOracleInput()
		require.NoError(t, err)
		ret, leftOverGas, err := env.Call(vm.AccountRef(addrs.Sender), addrs.Oracle, input, startingGas, big.NewInt(0))
		require.NoError(t, err, "evm must not fail (ret: %x, gas: %d)", ret, startingGas-leftOverGas)
	}

	input := wit.EncodeStepInput()

	ret, leftOverGas, err := env.Call(vm.AccountRef(addrs.Sender), addrs.RISCV, input, startingGas, big.NewInt(0))
	require.NoError(t, err, "evm must not fail (ret: %x), at step %d", ret, step)
	gasUsed = startingGas - leftOverGas

	require.Len(t, ret, 32)
	postHash = *(*[32]byte)(ret)
	logs := env.StateDB.(interface {
		Logs() []*types.Log
	}).Logs()
	require.Equal(t, 1, len(logs), "expecting a log with post-state")
	postState = logs[0].Data

	stateHash, err := fast.StateWitness(postState).StateHash()
	require.NoError(t, err, "state hash could not be computed")
	require.Equal(t, stateHash, postHash, "logged state must be accurate")

	env.StateDB.RevertToSnapshot(snap)
	return
}
