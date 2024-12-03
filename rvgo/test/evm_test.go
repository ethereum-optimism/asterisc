package test

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/holiman/uint256"
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
	"github.com/ethereum/go-ethereum/triedb"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
)

type dummyChain struct {
	startTime uint64
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

func loadRISCVContractCode(t require.TestingT) *Contract {
	dat, err := os.ReadFile("../../rvsol/out/RISCV.sol/RISCV.json")
	require.NoError(t, err)
	var outDat Contract
	err = json.Unmarshal(dat, &outDat)
	require.NoError(t, err)
	return &outDat
}

func loadPreimageOracleContractCode(t require.TestingT) *Contract {
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
	// Temporary hack until Cancun is activated on mainnet
	chainCfg := params.MainnetChainConfig
	offsetBlocks := uint64(1000)
	bc := &dummyChain{startTime: *chainCfg.CancunTime + offsetBlocks*12}
	header := bc.GetHeader(common.Hash{}, 19426587+offsetBlocks)
	header.Time = bc.startTime
	db := rawdb.NewMemoryDatabase()
	statedb, err := state.New(types.EmptyRootHash, state.NewDatabase(triedb.NewDatabase(db, nil), nil))
	require.NoError(t, err)
	blockContext := core.NewEVMBlockContext(header, bc, nil, chainCfg, statedb)
	vmCfg := vm.Config{}

	env := vm.NewEVM(blockContext, vm.TxContext{}, statedb, chainCfg, vmCfg)
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

func testContracts(t require.TestingT) *Contracts {
	return &Contracts{
		RISCV:  loadRISCVContractCode(t),
		Oracle: loadPreimageOracleContractCode(t),
	}
}

func stepEVM(t *testing.T, env *vm.EVM, wit *fast.StepWitness, addrs *Addresses, step uint64, revertCode []byte) (postState []byte, postHash common.Hash, gasUsed uint64) {
	startingGas := uint64(30_000_000)

	snap := env.StateDB.Snapshot()

	if wit.HasPreimage() {
		input, err := wit.EncodePreimageOracleInput(fast.LocalContext{})
		require.NoError(t, err)
		ret, leftOverGas, err := env.Call(vm.AccountRef(addrs.Sender), addrs.Oracle, input, startingGas, uint256.NewInt(0))
		require.NoError(t, err, "evm must not fail (ret: %x, gas: %d)", ret, startingGas-leftOverGas)
	}

	input, err := wit.EncodeStepInput(fast.LocalContext{})
	require.NoError(t, err)

	ret, leftOverGas, err := env.Call(vm.AccountRef(addrs.Sender), addrs.RISCV, input, startingGas, uint256.NewInt(0))
	if revertCode != nil {
		require.ErrorIs(t, err, vm.ErrExecutionReverted)
		require.Equal(t, ret, revertCode)
		return
	}
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
