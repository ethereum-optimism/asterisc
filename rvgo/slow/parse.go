package slow

// Functions to parse the instruction field values from different types of RISC-V instructions
// These should 1:1 match with the same definitions in the fast package.

func parseImmTypeI(instr U64) U64 {
	return signExtend64(shr64(instr, toU64(20)), toU64(11))
}

func parseImmTypeS(instr U64) U64 {
	return signExtend64(or64(shl64(shr64(instr, toU64(25)), toU64(5)), and64(shr64(instr, toU64(7)), toU64(0x1F))), toU64(11))
}

func parseImmTypeB(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				shl64(and64(shr64(instr, toU64(8)), toU64(0xF)), toU64(1)),
				shl64(and64(shr64(instr, toU64(25)), toU64(0x3F)), toU64(5)),
			),
			or64(
				shl64(and64(shr64(instr, toU64(7)), toU64(1)), toU64(11)),
				shl64(shr64(instr, toU64(31)), toU64(12)),
			),
		),
		toU64(12),
	)
}

func parseImmTypeU(instr U64) U64 {
	return signExtend64(shr64(instr, toU64(12)), toU64(19))
}

func parseImmTypeJ(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				shl64(and64(shr64(instr, toU64(21)), shortToU64(0x1FF)), toU64(1)),
				shl64(and64(shr64(instr, toU64(20)), toU64(1)), toU64(10)),
			),
			or64(
				shl64(and64(shr64(instr, toU64(12)), toU64(0xFF)), toU64(11)),
				shl64(shr64(instr, toU64(31)), toU64(19)),
			),
		),
		toU64(19),
	)
}

func parseOpcode(instr U64) U64 {
	return and64(instr, toU64(0x7F))
}

func parseRd(instr U64) U64 {
	return and64(shr64(instr, toU64(7)), toU64(0x1F))
}

func parseFunct3(instr U64) U64 {
	return and64(shr64(instr, toU64(12)), toU64(0x7))
}

func parseRs1(instr U64) U64 {
	return and64(shr64(instr, toU64(15)), toU64(0x1F))
}

func parseRs2(instr U64) U64 {
	return and64(shr64(instr, toU64(20)), toU64(0x1F))
}

func parseFunct7(instr U64) U64 {
	return shr64(instr, toU64(25))
}
