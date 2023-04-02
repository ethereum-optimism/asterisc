package fast

import (
	"debug/elf"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/params"

	"github.com/protolambda/asterisc/rvgo/oracle"
	"github.com/protolambda/asterisc/rvgo/slow"

	"github.com/stretchr/testify/require"

	"github.com/protolambda/asterisc/rvgo/fast"
)

func forEachTestSuite(t *testing.T, path string, callItem func(t *testing.T, path string)) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("missing tests: %s", path)
	} else {
		require.NoError(t, err, "failed to stat path")
	}
	items, err := os.ReadDir(path)
	require.NoError(t, err, "failed to read dir items")
	require.NotEmpty(t, items, "expected at least one test suite binary")

	for _, item := range items {
		if !item.IsDir() && !strings.HasSuffix(item.Name(), ".dump") {
			t.Run(item.Name(), func(t *testing.T) {
				callItem(t, filepath.Join(path, item.Name()))
			})
		}
	}
}

func runFastTestSuite(t *testing.T, path string) {
	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	for i := 0; i < 10_000; i++ {
		//fmt.Printf("pc: 0x%x\n", vmState.PC)
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)
		}
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		testCaseNum := vmState.Exit >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

func runSlowTestSuite(t *testing.T, path string) {
	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	so := oracle.NewStateOracle()
	pre := vmState.Merkleize(so)

	maxAccessListLen := 0

	for i := 0; i < 10_000; i++ {
		so.BuildAccessList(true)
		//t.Logf("next step - pc: 0x%x\n", vmState.PC)

		post, err := slow.Step(pre, so, os.Stdout, os.Stderr)
		require.NoError(t, err)

		al := so.AccessList()
		alo := &oracle.AccessListOracle{AccessList: al}

		// Now run the same in fast mode
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)
		}

		fastRoot := vmState.Merkleize(so)
		if post != fastRoot {
			so.Diff(post, fastRoot, 1)
			t.Fatalf("slow state %x must match fast state %x", post, fastRoot)
		}

		post2, err := slow.Step(pre, alo, os.Stdout, os.Stderr)
		require.NoError(t, err)
		if post2 != fastRoot {
			so.Diff(post2, fastRoot, 1)
			t.Fatalf("access-list slow state %x must match fast state %x", post2, fastRoot)
		}
		if len(al) > maxAccessListLen {
			maxAccessListLen = len(al)
		}

		pre = post

		if vmState.Exited {
			break
		}
	}

	t.Logf("max access-list length: %d", maxAccessListLen)

	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		testCaseNum := vmState.Exit >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

// TODO iterate all test suites
// TODO maybe load ELF sections for debugging
// TODO if step PC matches test symbol address, then log that we entered the test case

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

func loadStepContractCode(t *testing.T) []byte {
	dat, err := os.ReadFile("../rvsol/out/Step.sol/Step.json")
	require.NoError(t, err)
	var outDat struct {
		DeployedBytecode struct {
			Object hexutil.Bytes `json:"object"`
		} `json:"deployedBytecode"`
	}
	err = json.Unmarshal(dat, &outDat)
	require.NoError(t, err)
	return outDat.DeployedBytecode.Object
}

var stepAddr = common.HexToAddress("0x1337")

func newEVMEnv(t *testing.T, stepContractCode []byte) *vm.EVM {
	chainCfg := params.MainnetChainConfig
	bc := &dummyChain{}
	header := bc.GetHeader(common.Hash{}, 100)
	blockContext := core.NewEVMBlockContext(header, bc, nil)
	vmCfg := vm.Config{}
	db := rawdb.NewMemoryDatabase()
	statedb := state.NewDatabase(db)
	state, err := state.New(types.EmptyRootHash, statedb, nil)
	require.NoError(t, err)

	state.SetCode(stepAddr, stepContractCode)

	return vm.NewEVM(blockContext, vm.TxContext{}, state, chainCfg, vmCfg)
}

func runEVMTestSuite(t *testing.T, path string) {
	code := loadStepContractCode(t)
	vmenv := newEVMEnv(t, code)
	vmenv.Config.Debug = false
	vmenv.Config.Tracer = logger.NewMarkdownLogger(&logger.Config{}, os.Stdout)

	sender := common.HexToAddress("0xaaaa")

	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	so := oracle.NewStateOracle()
	pre := vmState.Merkleize(so)

	maxAccessListLen := 0
	maxGasUsed := uint64(0)

	for i := 0; i < 10_000; i++ {
		so.BuildAccessList(true)
		//t.Logf("next step - pc: 0x%x\n", vmState.PC)

		post, err := slow.Step(pre, so, os.Stdout, os.Stderr)
		require.NoError(t, err)

		al := so.AccessList()

		// Now run the same in fast mode
		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)
		}

		fastRoot := vmState.Merkleize(so)
		if post != fastRoot {
			so.Diff(post, fastRoot, 1)
			t.Fatalf("slow state %x must match fast state %x", post, fastRoot)
		}

		// Now run the same in EVM, but using the access-list
		input := oracle.Input(al, pre)
		startingGas := uint64(30_000_000)
		ret, leftOverGas, err := vmenv.Call(vm.AccountRef(sender), stepAddr, input, startingGas, big.NewInt(0))
		require.NoError(t, err, "evm must no fail (ret: %x)", ret)
		gasUsed := startingGas - leftOverGas
		if gasUsed > maxGasUsed {
			maxGasUsed = gasUsed
		}
		require.Len(t, ret, 32)
		post2 := common.BytesToHash(ret)
		if post2 != fastRoot {
			so.Diff(post2, fastRoot, 1)
			t.Fatalf("evm state %x must match fast state %x", post2, fastRoot)
		}
		if len(al) > maxAccessListLen {
			maxAccessListLen = len(al)
		}

		pre = post

		if vmState.Exited {
			break
		}
	}

	t.Logf("max access-list length: %d", maxAccessListLen)
	t.Logf("max gas used: %d", maxGasUsed)

	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.Exit != 0 {
		testCaseNum := vmState.Exit >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

func TestFastStep(t *testing.T) {
	testsPath := filepath.FromSlash("../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runFastTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}

func TestSlowStep(t *testing.T) {
	testsPath := filepath.FromSlash("../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runSlowTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}

func TestEVMStep(t *testing.T) {
	testsPath := filepath.FromSlash("../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runEVMTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	//runTestCategory("rv64ua-p")  // TODO implement atomic instructions extension
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}
