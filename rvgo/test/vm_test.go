package test

import (
	"debug/elf"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/asterisc/rvgo/slow"
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

	for i := 0; i < 10_000; i++ {
		//t.Logf("next step - pc: 0x%x\n", vmState.PC)

		wit, err := instState.Step(true)
		require.NoError(t, err)

		// Now run the same in slow mode
		input, err := wit.EncodeStepInput(fast.LocalContext{})
		require.NoError(t, err)
		post, err := slow.Step(input, nil)
		require.NoErrorf(t, err, "slow VM err at step %d, PC %08x: %v", i, vmState.PC, err)

		fastPostState := vmState.EncodeWitness()
		fastRoot, err := fastPostState.StateHash()
		require.NoError(t, err)
		if post != fastRoot {
			t.Fatalf("slow state %x must match fast state %x", post, fastRoot)
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

func runEVMTestSuite(t *testing.T, path string) {
	contracts := testContracts(t)
	addrs := testAddrs
	env := newEVMEnv(t, contracts, addrs)
	//addTracer(t, env, addrs, contracts)

	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	instState := fast.NewInstrumentedState(vmState, nil, nil, nil)

	maxGasUsed := uint64(0)

	for i := uint64(0); i < 10_000; i++ {
		//t.Logf("next step - pc: 0x%x\n", vmState.PC)

		wit, err := instState.Step(true)
		require.NoError(t, err)

		evmPost, evmPostHash, gasUsed := stepEVM(t, env, wit, addrs, i, nil)
		if gasUsed > maxGasUsed {
			maxGasUsed = gasUsed
		}

		fastPostState := vmState.EncodeWitness()
		fastStateHash, err := fastPostState.StateHash()
		require.NoError(t, err)
		if evmPostHash != fastStateHash {
			t.Fatalf("evm state %x must match fast state %x\nfast:%x\nevm: %x\n", evmPostHash, fastStateHash, fastPostState, evmPost)
		}

		if vmState.Exited {
			break
		}
	}

	t.Logf("max gas used: %d", maxGasUsed)

	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		testCaseNum := vmState.ExitCode >> 1
		t.Fatalf("failed at test case %d", testCaseNum)
	}
}

func TestFastStep(t *testing.T) {
	testsPath := filepath.FromSlash("../../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runFastTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?) https://github.com/ethereum-optimism/asterisc/issues/89
}

func TestSlowStep(t *testing.T) {
	testsPath := filepath.FromSlash("../../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runSlowTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?) https://github.com/ethereum-optimism/asterisc/issues/89
}

func TestEVMStep(t *testing.T) {
	testsPath := filepath.FromSlash("../../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runEVMTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	runTestCategory("rv64ua-p")
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?) https://github.com/ethereum-optimism/asterisc/issues/89
}
