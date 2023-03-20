package slow

// Functions to parse the instruction field values from different types of RISC-V instructions
// These should 1:1 match with the same definitions in the fast package.

func parseImmTypeI(instr U64) U64 {
	return signExtend64(shr64(toU64(20), instr), toU64(11))
}

func parseImmTypeS(instr U64) U64 {
	return signExtend64(
		or64(
			shl64(toU64(5), shr64(toU64(25), instr)),
			and64(shr64(toU64(7), instr), toU64(0x1F)),
		),
		toU64(11))
}

func parseImmTypeB(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				shl64(toU64(1), and64(shr64(toU64(8), instr), toU64(0xF))),
				shl64(toU64(5), and64(shr64(toU64(25), instr), toU64(0x3F))),
			),
			or64(
				shl64(toU64(11), and64(shr64(toU64(7), instr), toU64(1))),
				shl64(toU64(12), shr64(toU64(31), instr)),
			),
		),
		toU64(12),
	)
}

func parseImmTypeU(instr U64) U64 {
	return signExtend64(shr64(toU64(12), instr), toU64(19))
}

func parseImmTypeJ(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				and64(shr64(toU64(21), instr), shortToU64(0x3FF)),          // 10 bits for index 0:9
				shl64(toU64(10), and64(shr64(toU64(20), instr), toU64(1))), // 1 bit for index 10
			),
			or64(
				shl64(toU64(11), and64(shr64(toU64(12), instr), toU64(0xFF))), // 8 bits for index 11:18
				shl64(toU64(19), shr64(toU64(31), instr)),                     // 1 bit for index 19
			),
		),
		toU64(19),
	)
}

func parseOpcode(instr U64) U64 {
	return and64(instr, toU64(0x7F))
}

func parseRd(instr U64) U64 {
	return and64(shr64(toU64(7), instr), toU64(0x1F))
}

func parseFunct3(instr U64) U64 {
	return and64(shr64(toU64(12), instr), toU64(0x7))
}

func parseRs1(instr U64) U64 {
	return and64(shr64(toU64(15), instr), toU64(0x1F))
}

func parseRs2(instr U64) U64 {
	return and64(shr64(toU64(20), instr), toU64(0x1F))
}

func parseFunct7(instr U64) U64 {
	return shr64(toU64(25), instr)
}

func parseCSSR(instr U64) U64 {
	return shr64(toU64(20), instr)
}
