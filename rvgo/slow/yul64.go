package slow

import "github.com/holiman/uint256"

// These are type-safe pure functions *styled to translate to yul*, to use uint256 variables for 64 bit math.

// U64 is like a Go uint64, always within range, but represented as uint256 in memory with 0 padding.
type U64 uint256.Int

func (v U64) val() uint64 {
	return (*uint256.Int)(&v).Uint64()
}

func toU256(v uint8) U256 {
	return *uint256.NewInt(uint64(v))
}

func toU64(v uint8) U64 {
	return U64(toU256(v))
}

func shortToU64(v uint16) U64 {
	return U64(*uint256.NewInt(uint64(v)))
}

func u256ToU64(v U256) U64 {
	return U64(and(v, U256(u64Mask())))
}

func u64ToU256(v U64) U256 {
	return U256(v)
}

func u64Mask() U64 { // max uint64
	return U64(shr(not(U256{}), toU256(192))) // 256-64 = 192
}

func u32Mask() U64 {
	return U64(shr(not(U256{}), toU256(224))) // 256-32 = 224
}

func u64Mod() U256 { // 1 << 64
	return shl(toU256(1), toU256(64))
}

func u64TopBit() U256 { // 1 << 63
	return shl(toU256(1), toU256(63))
}

func signExtend64(v U64, bit U64) U64 {
	switch and(U256(v), shl(toU256(1), U256(bit))) {
	case U256{}:
		// fill with zeroes, by masking
		return U64(and(U256(v), shr(U256(u64Mask()), sub(toU256(63), U256(bit)))))
	default:
		// fill with ones, by or-ing
		return U64(or(U256(v), shl(shr(U256(u64Mask()), U256(bit)), U256(bit))))
	}
}

func signExtend64To256(v U64) U256 {
	switch and(U256(v), u64TopBit()) {
	case U256{}:
		return U256(v)
	default:
		return or(shl(not(U256{}), toU256(64)), U256(v))
	}
}

func add64(x, y U64) (out U64) {
	out = U64(mod(add(U256(x), U256(y)), u64Mod()))
	return
}

func sub64(x, y U64) (out U64) {
	out = U64(mod(sub(U256(x), U256(y)), u64Mod()))
	return
}

func mul64(x, y U64) (out U64) {
	out = u256ToU64(mul(U256(x), U256(y)))
	return
}

func div64(x, y U64) (out U64) {
	out = u256ToU64(div(U256(x), U256(y)))
	return
}

func sdiv64(x, y U64) (out U64) { // note: signed overflow semantics are the same between Go and EVM assembly
	out = u256ToU64(sdiv(U256(x), U256(y)))
	return
}

func mod64(x, y U64) (out U64) {
	out = U64(mod(U256(x), U256(y)))
	return
}

func smod64(x, y U64) (out U64) {
	out = u256ToU64(smod(U256(x), U256(y)))
	return
}

func not64(x U64) (out U64) {
	out = u256ToU64(not(U256(x)))
	return
}

func lt64(x, y U64) (out U64) {
	out = U64(lt(U256(x), U256(y)))
	return
}

func gt64(x, y U64) (out U64) {
	out = U64(gt(U256(x), U256(y)))
	return
}

func slt64(x, y U64) (out U64) {
	out = U64(slt(signExtend64To256(x), signExtend64To256(y)))
	return
}

func sgt64(x, y U64) (out U64) {
	out = U64(sgt(signExtend64To256(x), signExtend64To256(y)))
	return
}

func eq64(x, y U64) (out U64) {
	out = U64(eq(U256(x), U256(y)))
	return
}

func iszero64(x U64) bool {
	return iszero(U256(x))
}

func and64(x, y U64) (out U64) {
	out = U64(and(U256(x), U256(y)))
	return
}

func or64(x, y U64) (out U64) {
	out = U64(or(U256(x), U256(y)))
	return
}

func xor64(x, y U64) (out U64) {
	out = U64(xor(U256(x), U256(y)))
	return
}

func shl64(x, y U64) (out U64) {
	out = u256ToU64(shl(U256(x), U256(y)))
	return
}

func shr64(x, y U64) (out U64) {
	out = U64(shr(U256(x), U256(y)))
	return
}

func sar64(x, y U64) (out U64) {
	out = u256ToU64(sar(signExtend64To256(x), U256(y)))
	return
}
