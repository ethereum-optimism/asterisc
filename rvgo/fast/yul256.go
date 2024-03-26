package fast

import "github.com/holiman/uint256"

// EVM yul functions
// Yul exposes all EVM opcodes as functions.

type U256 = uint256.Int

// bytes interpreted as big endian uint256
func b32asBEWord(v [32]byte) (out U256) {
	out.SetBytes32(v[:])
	return
}

func beWordAsB32(v U256) [32]byte {
	return v.Bytes32()
}

// nolint:unused
func add(x, y U256) (out U256) {
	out.Add(&x, &y)
	return
}

func sub(x, y U256) (out U256) {
	out.Sub(&x, &y)
	return
}

func mul(x, y U256) (out U256) {
	out.Mul(&x, &y)
	return
}

// nolint:unused
func div(x, y U256) (out U256) {
	out.Div(&x, &y)
	return
}

// nolint:unused
func sdiv(x, y U256) (out U256) { // note: signed overflow semantics are the same between Go and EVM assembly
	out.SDiv(&x, &y)
	return
}

// nolint:unused
func mod(x, y U256) (out U256) {
	out.Mod(&x, &y)
	return
}

// nolint:unused
func smod(x, y U256) (out U256) {
	out.SMod(&x, &y)
	return
}

func not(x U256) (out U256) {
	out.Not(&x)
	return
}

// nolint:unused
func lt(x, y U256) (out U256) {
	if x.Lt(&y) {
		out.SetUint64(1)
	}
	return
}

// nolint:unused
func gt(x, y U256) (out U256) {
	if x.Gt(&y) {
		out.SetUint64(1)
	}
	return
}

// nolint:unused
func slt(x, y U256) (out U256) {
	if x.Slt(&y) {
		out.SetUint64(1)
	}
	return
}

// nolint:unused
func sgt(x, y U256) (out U256) {
	if x.Sgt(&y) {
		out.SetUint64(1)
	}
	return
}

// nolint:unused
func eq(x, y U256) (out U256) {
	if x.Eq(&y) {
		out.SetUint64(1)
	}
	return
}

// nolint:unused
func iszero(x U256) bool {
	return x.IsZero()
}

func and(x, y U256) (out U256) {
	out.And(&x, &y)
	return
}

func or(x, y U256) (out U256) {
	out.Or(&x, &y)
	return
}

// nolint:unused
func xor(x, y U256) (out U256) {
	out.Xor(&x, &y)
	return
}

// returns y << x
func shl(x, y U256) (out U256) {
	if !x.IsUint64() && x.Uint64() >= 256 {
		return
	}
	out.Lsh(&y, uint(x.Uint64()))
	return
}

// returns y >> x
func shr(x, y U256) (out U256) {
	if !x.IsUint64() && x.Uint64() >= 256 {
		return
	}
	out.Rsh(&y, uint(x.Uint64()))
	return
}

// returns y >> x (signed)
// nolint:unused
func sar(x, y U256) (out U256) {
	if !x.IsUint64() && x.Uint64() >= 256 {
		return
	}
	out.SRsh(&y, uint(x.Uint64()))
	return
}
