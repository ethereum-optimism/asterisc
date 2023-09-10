package fast

import (
	"debug/elf"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/protolambda/asterisc/rvgo/slow"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	inState := fast.NewInstrumentedState(vmState, nil, os.Stdout, os.Stderr)

	for i := 0; i < 10_000; i++ {
		//fmt.Printf("pc: 0x%x\n", vmState.PC)
		if _, err := inState.Step(false); err != nil {
			t.Fatalf("VM err at step %d, PC %x: %v", i, vmState.PC, err)
		}
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		testCaseNum := vmState.ExitCode >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

func runSlowTestSuite(t *testing.T, path string) {
	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	instState := fast.NewInstrumentedState(vmState, nil, nil, nil)
	maxAccessListLen := 0

	for i := 0; i < 10_000; i++ {
		//t.Logf("next step - pc: 0x%x\n", vmState.PC)

		wit, err := instState.Step(true)
		require.NoError(t, err)

		// Now run the same in slow mode
		calldata := wit.EncodeStepInput()
		post, err := slow.Step(calldata, nil)
		require.NoErrorf(t, err, "slow VM err at step %d, PC %d: %v", i, vmState.PC, err)

		fastPostState := vmState.EncodeWitness()

		fastRoot := crypto.Keccak256Hash(fastPostState)
		if post != fastRoot {
			t.Fatalf("slow state %x must match fast state %x", post, fastRoot)
		}

		if vmState.Exited {
			break
		}
	}

	t.Logf("max access-list length: %d", maxAccessListLen)

	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		testCaseNum := vmState.ExitCode >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

/*

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

func loadPreimageOracleContractCode(t *testing.T) []byte {
	dat, err := os.ReadFile("../rvsol/out/PreimageOracle.sol/PreimageOracle.json")
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

var preimageOracleAddr = common.HexToAddress("0xf00d")

func newEVMEnv(t *testing.T, stepContractCode []byte, preimageOracleCode []byte) *vm.EVM {
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
	env.StateDB.SetCode(stepAddr, stepContractCode)
	env.StateDB.SetCode(preimageOracleAddr, preimageOracleCode)
	env.StateDB.SetState(stepAddr, common.Hash{}, preimageOracleAddr.Hash())
	return env
}

func runEVMTestSuite(t *testing.T, path string) {
	code := loadStepContractCode(t)
	vmenv := newEVMEnv(t, code, []byte{0}) // these tests run without pre-image oracle
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

		post, err := slow.Step(pre, so, nil)
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
		require.NoError(t, err, "evm must not fail (ret: %x)", ret)
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
	if vmState.ExitCode != 0 {
		testCaseNum := vmState.ExitCode >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}
*/

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

/*
func TestEVMStep(t *testing.T) {
	testsPath := filepath.FromSlash("../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runEVMTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}
*/
