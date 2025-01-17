package fast

// Functions to parse the instruction field values from different types of RISC-V instructions
// These should 1:1 match with the same definitions in the slow package.

func parseImmTypeI(instr U64) U64 {
	return signExtend64(shr64(byteToU64(20), instr), byteToU64(11))
}

func parseImmTypeS(instr U64) U64 {
	return signExtend64(
		or64(
			shl64(byteToU64(5), shr64(byteToU64(25), instr)),
			and64(shr64(byteToU64(7), instr), byteToU64(0x1F)),
		),
		byteToU64(11))
}

func parseImmTypeB(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				shl64(byteToU64(1), and64(shr64(byteToU64(8), instr), byteToU64(0xF))),
				shl64(byteToU64(5), and64(shr64(byteToU64(25), instr), byteToU64(0x3F))),
			),
			or64(
				shl64(byteToU64(11), and64(shr64(byteToU64(7), instr), byteToU64(1))),
				shl64(byteToU64(12), shr64(byteToU64(31), instr)),
			),
		),
		byteToU64(12),
	)
}

func parseImmTypeU(instr U64) U64 {
	return signExtend64(shr64(byteToU64(12), instr), byteToU64(19))
}

func parseImmTypeJ(instr U64) U64 {
	return signExtend64(
		or64(
			or64(
				and64(shr64(byteToU64(21), instr), shortToU64(0x3FF)),                  // 10 bits for index 0:9
				shl64(byteToU64(10), and64(shr64(byteToU64(20), instr), byteToU64(1))), // 1 bit for index 10
			),
			or64(
				shl64(byteToU64(11), and64(shr64(byteToU64(12), instr), byteToU64(0xFF))), // 8 bits for index 11:18
				shl64(byteToU64(19), shr64(byteToU64(31), instr)),                         // 1 bit for index 19
			),
		),
		byteToU64(19),
	)
}

func parseOpcode(instr U64) U64 {
	return and64(instr, byteToU64(0x7F))
}

func parseRd(instr U64) U64 {
	return and64(shr64(byteToU64(7), instr), byteToU64(0x1F))
}

func parseFunct3(instr U64) U64 {
	return and64(shr64(byteToU64(12), instr), byteToU64(0x7))
}

func parseRs1(instr U64) U64 {
	return and64(shr64(byteToU64(15), instr), byteToU64(0x1F))
}

func parseRs2(instr U64) U64 {
	return and64(shr64(byteToU64(20), instr), byteToU64(0x1F))
}

func parseFunct7(instr U64) U64 {
	return shr64(byteToU64(25), instr)
}
