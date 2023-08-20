package fast

import "github.com/holiman/uint256"

// Fast equivalent of the 64-bit yul functions of slow-mode

type U64 = uint64

func toU256(v uint8) U256 {
	return *uint256.NewInt(uint64(v))
}

func toU64(v uint8) U64 { return uint64(v) }

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
	return signExtend64(and64(v, u32Mask()), toU64(31))
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
		return or(shl(toU256(64), not(U256{})), *new(uint256.Int).SetUint64(v))
	}
}

func add64(x, y uint64) uint64 {
	return x + y
}

func sub64(x, y uint64) uint64 {
	return x - y
}

func mul64(x, y uint64) uint64 {
	return x * y
}

func div64(x, y uint64) uint64 {
	if y == 0 {
		return 0
	}
	return x / y
}

func sdiv64(x, y uint64) uint64 { // note: signed overflow semantics are the same between Go and EVM assembly
	if y == 0 {
		return 0
	}
	if x == uint64(1<<63) && y == ^uint64(0) {
		return 1 << 63
	}
	return uint64(int64(x) / int64(y))
}

func mod64(x, y uint64) uint64 {
	if y == 0 {
		return 0
	} else {
		return x % y
	}
}

func smod64(x, y uint64) uint64 {
	if y == 0 {
		return 0
	} else {
		return uint64(int64(x) % int64(y))
	}
}

func not64(x uint64) uint64 {
	return ^x
}

func lt64(x, y uint64) uint64 {
	if x < y {
		return 1
	} else {
		return 0
	}
}

func gt64(x, y uint64) uint64 {
	if x > y {
		return 1
	} else {
		return 0
	}
}

func slt64(x, y uint64) uint64 {
	if int64(x) < int64(y) {
		return 1
	} else {
		return 0
	}
}

func sgt64(x, y uint64) uint64 {
	if int64(x) > int64(y) {
		return 1
	} else {
		return 0
	}
}

func eq64(x, y uint64) uint64 {
	if x == y {
		return 1
	} else {
		return 0
	}
}

func iszero64(x uint64) bool {
	return x == 0
}

func and64(x, y uint64) uint64 {
	return x & y
}

func or64(x, y uint64) uint64 {
	return x | y
}

func xor64(x, y uint64) uint64 {
	return x ^ y
}

func shl64(x, y uint64) uint64 {
	return y << x
}

func shr64(x, y uint64) uint64 {
	return y >> x
}

func sar64(x, y uint64) uint64 {
	return uint64(int64(y) >> x)
}
