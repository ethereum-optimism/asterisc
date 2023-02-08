package fast

import (
	"debug/elf"
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

func runTestSuite(t *testing.T, path string) {
	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	for i := 0; i < 10_000; i++ {
		//fmt.Printf("pc: 0x%x\n", vmState.PC)
		fast.Step(vmState)
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

func TestStep(t *testing.T) {
	// TODO iterate all test suites
	// TODO maybe load ELF sections for debugging
	// TODO if step PC matches test symbol address, then log that we entered the test case

	testsPath := filepath.FromSlash("../tests/riscv-tests")

	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runTestSuite)
		})
	}

	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	//runTestCategory("rv64ua-p")  // TODO implement atomic instructions extension
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}
