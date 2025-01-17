package fast

import "github.com/holiman/uint256"

// Fast equivalent of the 64-bit yul functions of slow-mode

type U64 = uint64

func byteToU256(v uint8) U256 {
	return *uint256.NewInt(uint64(v))
}

func byteToU64(v uint8) U64 { return uint64(v) }

func shortToU64(v uint16) U64 {
	return uint64(v)
}

func shortToU256(v uint16) U256 {
	return *uint256.NewInt(uint64(v))
}

func longToU256(v uint64) U256 {
	return *uint256.NewInt(v)
}

func u256ToU64(v U256) U64 {
	return v.Uint64()
}

func u64ToU256(v U64) U256 {
	return *uint256.NewInt(v)
}

func u64Mask() uint64 { // max uint64
	return 0xFFFF_FFFF_FFFF_FFFF
}

func u32Mask() uint64 {
	return 0xFFFF_FFFF
}

func mask32Signed64(v U64) U64 {
	return signExtend64(and64(v, u32Mask()), byteToU64(31))
}

func signExtend64(v uint64, bit uint64) uint64 {
	switch and64(v, shl64(bit, 1)) {
	case 0:
		// fill with zeroes, by masking
		return and64(v, shr64(sub64(63, bit), u64Mask()))
	default:
		// fill with ones, by or-ing
		return or64(v, shl64(bit, shr64(bit, u64Mask())))
	}
}

func signExtend64To256(v U64) U256 {
	switch v & (1 << 63) {
	case 0:
		return *new(uint256.Int).SetUint64(v)
	default:
		return or(shl(byteToU256(64), not(U256{})), *new(uint256.Int).SetUint64(v))
	}
}

func add64(x, y uint64) uint64 {
	return u256ToU64(add(longToU256(x), longToU256(y)))
}

func sub64(x, y uint64) uint64 {
	return u256ToU64(sub(longToU256(x), longToU256(y)))
}

func mul64(x, y uint64) uint64 {
	return u256ToU64(mul(longToU256(x), longToU256(y)))
}

func div64(x, y uint64) uint64 {
	return u256ToU64(div(longToU256(x), longToU256(y)))
}

func sdiv64(x, y uint64) uint64 { // note: signed overflow semantics are the same between Go and EVM assembly
	return u256ToU64(sdiv(signExtend64To256(x), signExtend64To256(y)))
}

func mod64(x, y uint64) uint64 {
	return u256ToU64(mod(longToU256(x), longToU256(y)))
}

func smod64(x, y uint64) uint64 {
	return u256ToU64(smod(signExtend64To256(x), signExtend64To256(y)))
}

func not64(x uint64) uint64 {
	return u256ToU64(not(longToU256(x)))
}

func lt64(x, y uint64) uint64 {
	return u256ToU64(lt(longToU256(x), longToU256(y)))
}

func gt64(x, y uint64) uint64 {
	return u256ToU64(gt(longToU256(x), longToU256(y)))
}

func slt64(x, y uint64) uint64 {
	return u256ToU64(slt(signExtend64To256(x), signExtend64To256(y)))
}

func sgt64(x, y uint64) uint64 {
	return u256ToU64(sgt(signExtend64To256(x), signExtend64To256(y)))
}

func eq64(x, y uint64) uint64 {
	return u256ToU64(eq(longToU256(x), longToU256(y)))
}

func iszero64(x uint64) bool {
	return iszero(longToU256(x))
}

func and64(x, y uint64) U64 {
	return u256ToU64(and(longToU256(x), longToU256(y)))
}

func or64(x, y uint64) uint64 {
	return u256ToU64(or(longToU256(x), longToU256(y)))
}

func xor64(x, y uint64) uint64 {
	return u256ToU64(xor(longToU256(x), longToU256(y)))
}

func shl64(x, y uint64) uint64 {
	return u256ToU64(shl(longToU256(x), longToU256(y)))
}

func shr64(x, y uint64) uint64 {
	return u256ToU64(shr(longToU256(x), longToU256(y)))
}

func sar64(x, y uint64) uint64 {
	return u256ToU64(sar(longToU256(x), signExtend64To256(y)))
}
