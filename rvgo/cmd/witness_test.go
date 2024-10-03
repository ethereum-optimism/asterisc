package cmd

import (
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/optimism/cannon/mipsevm"
)

//go:embed test_data/state.json
var testState []byte

var asteriscWitnessLen = 362

func TestLoadState(t *testing.T) {
	t.Run("Uncompressed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json")
		require.NoError(t, os.WriteFile(path, testState, 0644))

		state, err := fast.LoadVMStateFromFile(path)
		require.NoError(t, err)

		var expected fast.VMState
		require.NoError(t, json.Unmarshal(testState, &expected))
		require.Equal(t, &expected, state)
	})

	t.Run("Gzipped", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "state.json.gz")
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		require.NoError(t, err)
		defer f.Close()
		writer := gzip.NewWriter(f)
		_, err = writer.Write(testState)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		state, err := fast.LoadVMStateFromFile(path)
		require.NoError(t, err)

		var expected fast.VMState
		require.NoError(t, json.Unmarshal(testState, &expected))
		require.Equal(t, &expected, state)
	})

	t.Run("InvalidStateWitness", func(t *testing.T) {
		invalidWitnessLen := asteriscWitnessLen - 1
		state := &fast.VMState{
			Step:    10,
			Exited:  true,
			Witness: make([]byte, invalidWitnessLen),
		}
		err := validateState(state)
		require.ErrorContains(t, err, "invalid witness")
	})

	t.Run("InvalidStateHash", func(t *testing.T) {
		state := &fast.VMState{
			Step:    10,
			Exited:  true,
			Witness: make([]byte, asteriscWitnessLen),
		}
		// Unknown exit code
		state.StateHash[0] = 37
		err := validateState(state)
		require.ErrorContains(t, err, "invalid stateHash: unknown exitCode")
		// Exited but ExitCode is VMStatusUnfinished
		state.StateHash[0] = 3
		err = validateState(state)
		require.ErrorContains(t, err, "invalid stateHash: invalid exitCode")
		// Not Exited but ExitCode is not VMStatusUnfinished
		state.Exited = false
		for exitCode := 0; exitCode < 3; exitCode++ {
			state.StateHash[0] = byte(exitCode)
			err = validateState(state)
			require.ErrorContains(t, err, "invalid stateHash: invalid exitCode")
		}
	})
}

// validateState performs verification of state; it is not perfect.
// It does not recalculate whether witness nor stateHash is correctly set from state.
func validateState(state *fast.VMState) error {
	if err := validateStateHash(state); err != nil {
		return err
	}
	if err := validateWitness(state); err != nil {
		return err
	}
	return nil
}

func validateStateHash(state *fast.VMState) error {
	exitCode := state.StateHash[0]
	if exitCode >= 4 {
		return fmt.Errorf("invalid stateHash: unknown exitCode %d", exitCode)
	}
	if (state.Exited && exitCode == mipsevm.VMStatusUnfinished) || (!state.Exited && exitCode != mipsevm.VMStatusUnfinished) {
		return fmt.Errorf("invalid stateHash: invalid exitCode %d", exitCode)
	}
	return nil
}

func validateWitness(state *fast.VMState) error {
	witnessLen := len(state.Witness)
	if witnessLen != asteriscWitnessLen {
		return fmt.Errorf("invalid witness: Length must be 362 but got %d", witnessLen)
	}
	return nil
}
