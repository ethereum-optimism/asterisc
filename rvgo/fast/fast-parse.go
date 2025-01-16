package fast

func ParseImmTypeI(instr U64) U64 {
	return parseImmTypeI(instr)
}

func ParseImmTypeS(instr U64) U64 {
	return parseImmTypeS(instr)
}

func ParseImmTypeB(instr U64) U64 {
	return parseImmTypeB(instr)
}

func ParseImmTypeU(instr U64) U64 {
	return parseImmTypeU(instr)
}

func ParseImmTypeJ(instr U64) U64 {
	return parseImmTypeJ(instr)
}

func ParseOpcode(instr U64) U64 {
	return parseOpcode(instr)
}

func ParseRd(instr U64) U64 {
	return parseRd(instr)
}

func ParseFunct3(instr U64) U64 {
	return parseFunct3(instr)
}

func ParseRs1(instr U64) U64 {
	return parseRs1(instr)
}

func ParseRs2(instr U64) U64 {
	return parseRs2(instr)
}
func ParseFunct7(instr U64) U64 {
	return parseFunct7(instr)
}
