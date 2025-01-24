package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/asterisc/rvgo/fast"
	"github.com/ethereum-optimism/asterisc/rvgo/slow"
)

func FuzzParseTypeI(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseImmTypeI(slow.U64{instr})
		var fastOutput = fast.ParseImmTypeI(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseTypeS(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseImmTypeS(slow.U64{instr})
		var fastOutput = fast.ParseImmTypeS(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseTypeB(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseImmTypeB(slow.U64{instr})
		var fastOutput = fast.ParseImmTypeB(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseTypeU(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseImmTypeU(slow.U64{instr})
		var fastOutput = fast.ParseImmTypeU(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseTypeJ(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseImmTypeJ(slow.U64{instr})
		var fastOutput = fast.ParseImmTypeJ(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseOpcode(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseOpcode(slow.U64{instr})
		var fastOutput = fast.ParseOpcode(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
func FuzzParseRd(f *testing.F) {
	f.Fuzz(func(t *testing.T, instr uint64) {
		var slowOutput = slow.ParseRd(slow.U64{instr})
		var fastOutput = fast.ParseRd(fast.U64(instr))

		require.Equal(t, slow.Val(slowOutput), fastOutput)
	})
}
