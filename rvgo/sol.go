package rvgo

// solidity assembly opcodes

func add(x, y uint64) uint64 {
	return x + y
}

func sub(x, y uint64) uint64 {
	return x - y
}

func mul(x, y uint64) uint64 {
	return x * y
}

func div(x, y uint64) uint64 {
	if y == 0 {
		return 0
	}
	return x / y
}

func sdiv(x, y uint64) uint64 { // note: signed overflow semantics are the same between Go and EVM assembly
	if y == 0 {
		return 0
	}
	if x == uint64(1<<63) && y == ^uint64(0) {
		return 1 << 63
	}
	return uint64(int64(x) / int64(y))
}

func mod(x, y uint64) uint64 {
	if y == 0 {
		return 0
	} else {
		return x % y
	}
}

func smod(x, y uint64) uint64 {
	if y == 0 {
		return 0
	} else {
		return uint64(int64(x) % int64(y))
	}
}

func not(x uint64) uint64 {
	return ^x
}

func lt(x, y uint64) uint64 {
	if x < y {
		return 1
	} else {
		return 0
	}
}

func gt(x, y uint64) uint64 {
	if x > y {
		return 1
	} else {
		return 0
	}
}

func slt(x, y uint64) uint64 {
	if int64(x) < int64(y) {
		return 1
	} else {
		return 0
	}
}

func sgt(x, y uint64) uint64 {
	if int64(x) > int64(y) {
		return 1
	} else {
		return 0
	}
}

func eq(x, y uint64) uint64 {
	if int64(x) == int64(y) {
		return 1
	} else {
		return 0
	}
}

func iszero(x uint64) bool {
	return x == 0
}

func and(x, y uint64) uint64 {
	return x & y
}

func or(x, y uint64) uint64 {
	return x | y
}

func xor(x, y uint64) uint64 {
	return x ^ y
}

func shl(x, y uint64) uint64 {
	return x << y
}

func shr(x, y uint64) uint64 {
	return x >> y
}

func sar(x, y uint64) uint64 {
	return uint64(int64(x) >> y)
}
