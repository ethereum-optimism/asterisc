package fast

import (
	"debug/elf"
	"fmt"
	"github.com/protolambda/asterisc/rvgo/oracle"
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

func runSlowTestSuite(t *testing.T, path string) {
	testSuiteELF, err := elf.Open(path)
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := fast.LoadELF(testSuiteELF)
	require.NoError(t, err, "must load test suite ELF binary")

	so := oracle.NewStateOracle()
	pre := slow.VMSubState{StateRoot: vmState.Merkleize(so)}

	for i := 0; i < 10_000; i++ {
		require.Equal(t, pre, slow.VMSubState{StateRoot: pre.StateRoot}, "vm state must be clean at start of instruction")
		//fmt.Printf("pc: 0x%x\n", vmState.PC)

		for i := 0; i < 1000; i++ {
			// build access list while we run the sub-step on the full oracle data
			so.BuildAccessList(true)
			post := slow.SubStep(pre, so)
			fmt.Println()
			fmt.Println("------------")
			fmt.Printf("state sub step: %d\n", post.SubIndex)
			fmt.Printf("state value: %x\n", post.StateValue[:])
			fmt.Printf("state stack depth:  %d\n", post.StateStackDepth)
			fmt.Printf("state stack gindex: %b\n", post.StateStackGindex.ToBig())
			fmt.Printf("state gindex:       %b\n", post.StateGindex.ToBig())

			// now run the step again, but on the access-list version of the oracle
			//al := so.AccessList()
			//so2 := &oracle.AccessListOracle{AccessList: al}
			//post2 := slow.SubStep(pre, so2)
			//require.Equal(t, post, post2, "need to reproduce same post-state with access list based oracle")

			// if the vm state is clean, the sub-step is done
			if post == (slow.VMSubState{StateRoot: post.StateRoot}) {
				break
			}

			pre = post
		}

		// Now run the same in fast mode
		fast.Step(vmState)

		require.Equal(t, pre.StateRoot, vmState.Merkleize(so), "slow state must match fast state")

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

// TODO iterate all test suites
// TODO maybe load ELF sections for debugging
// TODO if step PC matches test symbol address, then log that we entered the test case

func TestFastStep(t *testing.T) {
	testsPath := filepath.FromSlash("../tests/riscv-tests")
	runTestCategory := func(name string) {
		t.Run(name, func(t *testing.T) {
			forEachTestSuite(t, filepath.Join(testsPath, name), runFastTestSuite)
		})
	}
	runTestCategory("rv64ui-p")
	runTestCategory("rv64um-p")
	//runTestCategory("rv64ua-p")  // TODO implement atomic instructions extension
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
	//runTestCategory("rv64ua-p")  // TODO implement atomic instructions extension
	//runTestCategory("benchmarks")  TODO benchmarks (fix ELF bench data loading and wrap in Go benchmark?)
}
