package fast

import (
	"debug/elf"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStep(t *testing.T) {
	// TODO iterate all test suites
	// TODO maybe load ELF sections for debugging
	// TODO if step PC matches test symbol address, then log that we entered the test case

	testSuiteELF, err := elf.Open("../../../riscv-tests/isa/rv64ui-p-add")
	require.NoError(t, err)
	defer testSuiteELF.Close()

	vmState, err := LoadELF(testSuiteELF)
	require.NoError(t, err)

	for i := 0; i < 1000; i++ {
		fmt.Printf("pc: 0x%x\n", vmState.PC)
		Step(vmState)
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
