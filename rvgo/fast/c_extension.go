package fast

import "fmt"

const (
	// DECOMPRESSED_SIZE is the size (in bytes) of a standard 32-bit instruction. Used to signal how many bytes the
	// program counter should be incremented by.
	DECOMPRESSED_SIZE = U64(4)
	// COMPRESSED_SIZE is the size (in bytes) of a `C` extension 16-bit instruction. Used to signal how many bytes the
	// program counter should be incremented by.
	COMPRESSED_SIZE = U64(2)
	// C_REGISTER_OFFSET is the offset of the register mapping from the `C` instructions to regular 32 bit instructions.
	// In the `C` extension, register fields are only allotted 3 bits, allowing for 8 possible register designations.
	C_REGISTER_OFFSET = U64(8)
	// SP is the index of the stack pointer register defined in the RV32I/RV64I base ISA.
	SP = U64(2)
)

// DecompressInstruction decompresses a 16-bit `C` extension RISC-V instruction into its 32-bit standard counterpart.
// For instructions passed that are not a part of the `C` extension, they are returned as-is. If an unknown compressed
// instruction is passed, an error is returned.
func DecompressInstruction(instr U64) (instrOut U64, pcBump U64, err error) {
	// If the instruction is not a `C` extension instruction, return it as-is.
	if !isCompressed(instr) {
		return instr, DECOMPRESSED_SIZE, nil
	}

	// If the instruction is compressed, mask off the low-order 16 bits prior to decompression. It is possible that the
	// VM reads 1, 1 1/2, or 2 instructions due to the static size of the `loadMem` operation that fetches the
	// instruction.
	instr = and64(instr, U64(0xFFFF))

	// Fully unset bits signals an illegal instruction, so we check here prior to the switch case in order to disambiguate
	// funct3 == 0 && opcode == 0 from C.ADDI4SPN.
	if instr == 0 {
		return 0, 0, fmt.Errorf("illegal instruction: %x", instr)
	}

	var decompressedInstr U64
	switch switchKeyC(instr) {
	// C.ADDI4SPN [OP: C0 | Funct3: 000 | Format: CIW]
	case 0x0:
		imm, reg := decodeCIW(instr)
		imm = or64(
			or64(
				shr64(toU64(2), and64(imm, toU64(0xC0))),
				shl64(toU64(4), and64(imm, toU64(0x3C))),
			),
			or64(
				shl64(toU64(1), and64(imm, toU64(0x02))),
				shl64(toU64(3), and64(imm, toU64(0x01))),
			),
		)
		decompressedInstr = recodeIType(
			0b0010011, // Arithmetic
			reg,       // rd - reg
			0,         // ADDI
			SP,        // rs1 - SP
			imm,       // immediate
		)
	// C.NOP, C.ADDI [OP: C1 | Funct3: 000 | Format: CI]
	case 0x1:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = recodeIType(
			0b0010011, // Arithmetic
			reg,       // rd
			0,         // ADDI
			reg,       // rs1
			imm,       // immediate
		)
	// C.SLLI [OP: C2 | Funct3: 000 | Format: CI]
	case 0x2:
		imm, reg := decodeCI(instr)
		imm = and64(imm, toU64(0x1F))
		decompressedInstr = recodeIType(
			0b0010011, // Arithmetic
			reg,       // rd
			1,         // SLLI - funct3
			reg,       // rs1
			imm,       // shamt
		)
	// C.FLD (Unsupported) [OP: C0 | Funct3: 001 | Format: CL]
	case 0x4:
		panic("unsupported - C.FLD")
	// C.ADDIW [OP: C1 | Funct3: 001 | Format: CI]
	case 0x5:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = recodeIType(
			0b0011011, // Arithmetic (RV64I)
			reg,       // rd
			0,         // ADDIW
			reg,       // rs1
			imm,       // immediate
		)
	// C.FLDSP (Unsupported) [OP: C2 | Funct3: 001 | Format: CI]
	case 0x6:
		panic("unsupported - C.FLDSP")
	// C.LW [OP: C0 | Funct3: 010 | Format: CL]
	case 0x8:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = shl64(toU64(1), and64(
			or64(shl64(toU64(5), imm), imm),
			toU64(0x3E),
		))
		decompressedInstr = recodeIType(
			0b0000011, // Load
			reg2,      // rd
			0b010,     // LW
			reg1,      // rs1
			imm,       // immediate
		)
	// C.LI [OP: C1 | Funct3: 010 | Format: CI]
	case 0x9:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = recodeIType(
			0b0010011, // Arithmetic
			reg,       // rd
			0,         // ADDI
			0,         // rs1
			imm,       // immediate
		)
	// C.LWSP [OP: C2 | Funct3: 010 | Format: CI]
	case 0xA:
		imm, reg := decodeCI(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			toU64(0xFC),
		)
		decompressedInstr = recodeIType(
			0b0000011, // Load
			reg,       // rd
			0b010,     // LW
			SP,        // rs1 - SP
			imm,       // immediate
		)
	// C.LD [OP: C0 | Funct3: 011 | Format: CL]
	case 0xC:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), shl64(toU64(1), imm)),
			toU64(0xF8),
		)
		decompressedInstr = recodeIType(
			0b0000011, // Load
			reg2,      // rd
			0b011,     // LD
			reg1,      // rs1
			imm,       // immediate
		)
	// C.ADDI16SP, C.LUI [OP: C1 | Funct3: 011 | Format: CI]
	case 0xD:
		imm, reg := decodeCI(instr)
		if reg == 2 {
			// C.ADDI16SP
			imm = or64(
				or64(
					or64(
						shl64(toU64(4), and64(imm, toU64(0x20))),
						and64(imm, toU64(0x10)),
					),
					or64(
						shl64(toU64(3), and64(imm, toU64(0x08))),
						shl64(toU64(6), and64(imm, toU64(0x06))),
					),
				),
				shl64(toU64(5), and64(imm, toU64(0x01))),
			)
			imm = signExtend64(imm, 9)
			decompressedInstr = recodeIType(
				0b0010011, // Arithmetic
				SP,        // rd - SP
				0,         // ADDI
				SP,        // rs1 - SP
				imm,       // immediate
			)
		} else {
			// C.LUI
			imm = signExtend64(shl64(toU64(12), imm), 17)
			decompressedInstr = recodeUType(
				0b0110111, // LUI
				reg,       // rd
				imm,       // immediate
			)
		}
	// C.LDSP [OP: C2 | Funct3: 011 | Format: CI]
	case 0xE:
		imm, reg := decodeCI(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			U64(0x1F8),
		)
		decompressedInstr = recodeIType(
			0b0000011, // Load
			reg,       // rd
			0b011,     // LD
			SP,        // rs1 - SP
			imm,       // immediate
		)
	// Reserved [OP: C0 | Funct3: 100 | Format: ~]
	case 0x10:
		return 0, 0, fmt.Errorf("hit reserved instruction: %x", instr)
	// C.SRLI, S.SRLI64, C.SRAI64, C.ANDI, C.SUB, C.XOR, C.OR, C.AND, C.SUBW, C.ADDW [OP: C1 | Funct3: 100 | Format: ?]
	case 0x11:
		switch instr >> 10 & 0x3 {
		// C.SRLI
		case 0x00:
			imm, reg := decodeShiftCB(instr)
			decompressedInstr = recodeRType(
				0b0010011, // Arithmetic
				reg,       // rd
				0b101,     // SRLI
				reg,       // rs1
				imm,       // shamt
				0,         // funct7 - SRLI
			)
		// C.SRAI
		case 0x01:
			imm, reg := decodeShiftCB(instr)
			decompressedInstr = recodeRType(
				0b0010011, // Arithmetic
				reg,       // rd
				0b101,     // SRAI
				reg,       // rs1
				imm,       // shamt
				0b0100000, // SRAI
			)
		// C.ANDI
		case 0x02:
			imm, reg := decodeShiftCB(instr)
			imm = signExtend64(imm, 5)
			decompressedInstr = recodeIType(
				0b0010011, // Arithmetic
				reg,       // rd
				0b111,     // ANDI
				reg,       // rs1
				imm,       // immediate
			)
		}

		_, reg1, reg2 := decodeCLCS(instr)
		switch (instr >> 8 & 0x1C) | (instr >> 5 & 0x03) {
		// C.SUB
		case 0x0C:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0b000,     // SUB
				reg1,      // rs1
				reg2,      // rs2
				0b0100000, // SUB
			)
		// C.XOR
		case 0x0D:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0b100,     // XOR
				reg1,      // rs1
				reg2,      // rs2
				0,         // XOR
			)
		// C.OR
		case 0x0E:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0b110,     // OR
				reg1,      // rs1
				reg2,      // rs2
				0,         // OR
			)
		// C.AND
		case 0x0F:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0b111,     // AND
				reg1,      // rs1
				reg2,      // rs2
				0,         // OR
			)
		// C.SUBW
		case 0x1C:
			decompressedInstr = recodeRType(
				0b0111011, // Arithmetic
				reg1,      // rd
				0b000,     // SUBW
				reg1,      // rs1
				reg2,      // rs2
				0b0100000, // SUBW
			)
		// C.ADDW
		case 0x1D:
			decompressedInstr = recodeRType(
				0b0111011, // Arithmetic
				reg1,      // rd
				0b000,     // ADDW
				reg1,      // rs1
				reg2,      // rs2
				0,         // ADDW
			)
		// Reserved
		case 0x1E, 0x1F:
		}
	// C.JR, C.MV, C.EBREAK, C.JALR, C.ADD [OP: C2 | Funct3: 100 | Format: ?]
	case 0x12:
		reg1, reg2 := decodeCR(instr)
		fnsel := instr & 0x1000
		switch {
		// C.JR
		case fnsel == 0 && reg2 == 0:
			decompressedInstr = recodeIType(
				0b1100111, // JALR
				0,         // rd
				0,         // JALR funct3
				reg1,      // rs1
				0,         // immediate
			)
		// C.MV
		case fnsel == 0:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0,         // funct3
				0,         // rs1
				reg2,      // rs2
				0,         // funct7 - ADD
			)
		// C.EBREAK
		case fnsel == 0x1000 && reg1 == 0 && reg2 == 0:
			decompressedInstr = 0b00000000000100000000000001110011
		// C.JALR
		case fnsel == 0x1000 && reg2 == 0:
			decompressedInstr = recodeIType(
				0b1100111, // JALR
				1,         // rd - RA register
				0,         // JALR funct3
				reg1,      // rs1
				0,         // immediate
			)
		// C.ADD
		default:
			decompressedInstr = recodeRType(
				0b0110011, // Arithmetic
				reg1,      // rd
				0,         // funct3
				reg1,      // rs1
				reg2,      // rs2
				0,         // funct7 - ADD
			)
		}
	// C.FSD (Unsupported) [OP: C0 | Funct3: 101 | Format: CS]
	case 0x14:
		panic("unsupported - C.FSD")
	// C.J [OP: C1 | Funct3: 101 | Format: CR]
	case 0x15:
		imm := decodeCJ(instr)
		imm = or64(
			or64(
				or64(
					shr64(toU64(5), and64(imm, U64(0x200))),
					shl64(toU64(4), and64(imm, toU64(0x40))),
				),
				or64(
					shl64(toU64(1), and64(imm, U64(0x5A0))),
					shl64(toU64(3), and64(imm, toU64(0x10))),
				),
			),
			or64(
				and64(toU64(0x0E), imm),
				shl64(toU64(5), and64(imm, toU64(0x01))),
			),
		)
		imm = signExtend64(imm, 11)
		decompressedInstr = recodeJType(
			0b1101111, // JAL
			0,         // rd
			imm,       // immediate
		)
	// C.FSDSP (Unsupported) [OP: C2 | Funct3: 101 | Format: CSS]
	case 0x16:
		panic("unsupported - C.FSDSP")
	// C.SW [OP: C0 | Funct3: 110 | Format: CS]
	case 0x18:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = and64(
			shl64(toU64(1), or64(shl64(toU64(5), imm), imm)),
			toU64(0x7C),
		)
		decompressedInstr = recodeSType(
			0b0100011, // Store
			0b010,     // SW
			reg1,      // rs1
			reg2,      // rs2
			imm,       // immediate
		)
	// C.BEQZ [OP: C1 | Funct3: 110 | Format: CB]
	case 0x19:
		imm, reg := decodeCB(instr)
		imm = or64(
			or64(
				or64(
					shl64(toU64(1), and64(imm, toU64(0x80))),
					shr64(toU64(2), and64(imm, toU64(0x60))),
				),
				or64(
					shl64(toU64(3), and64(imm, toU64(0x18))),
					and64(imm, toU64(0x06)),
				),
			),
			shl64(toU64(5), and64(imm, toU64(0x01))),
		)
		imm = signExtend64(imm, 8)
		decompressedInstr = recodeBType(
			0b1100011, // branch
			0,         // BEQ
			reg,       // rs1
			0,         // rs2 - ZERO
			imm,       // immediate
		)
	// C.SWSP [OP: C2 | Funct3: 110 | Format: CSS]
	case 0x1A:
		imm, reg := decodeCSS(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			toU64(0xFC),
		)
		decompressedInstr = recodeSType(
			0b0100011, // Store
			0b010,     // SW
			SP,        // rs1
			reg,       // rs2
			imm,       // immediate
		)
	// C.SD [OP: C0 | Funct3: 111 | Format: CS]
	case 0x1C:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = and64(
			shl64(
				toU64(1),
				or64(shl64(toU64(5), imm), imm),
			),
			toU64(0xF8),
		)
		decompressedInstr = recodeSType(
			0b0100011, // Store
			0b011,     // SD
			reg1,      // rs1
			reg2,      // rs2
			imm,       // immediate
		)
	// C.BNEZ [OP: C1 | Funct3: 111 | Format: CB]
	case 0x1D:
		imm, reg := decodeCB(instr)
		imm = or64(
			or64(
				or64(
					shl64(toU64(1), and64(imm, toU64(0x80))),
					shr64(toU64(2), and64(imm, toU64(0x60))),
				),
				or64(
					shl64(toU64(3), and64(imm, toU64(0x18))),
					and64(imm, toU64(0x06)),
				),
			),
			shl64(toU64(5), and64(imm, toU64(0x01))),
		)
		imm = signExtend64(imm, toU64(8))
		decompressedInstr = recodeBType(
			0b1100011, // Branch
			1,         // BNE
			reg,       // rs1
			0,         // rs2 - ZERO
			imm,       // immediate
		)
	// C.SDSP [OP: C2 | Funct3: 111 | Format: CSS]
	case 0x1E:
		imm, reg := decodeCSS(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			U64(0x1F8),
		)
		decompressedInstr = recodeSType(
			0b0100011, // Store
			0b011,     // SD
			SP,        // rs1
			reg,       // rs2
			imm,       // immediate
		)
	default:
		return 0, 0, fmt.Errorf("unknown instruction: %x", instr)
	}

	return decompressedInstr, COMPRESSED_SIZE, nil
}

////////////////////////////////////////////////////////////////
//                          HELPERS                           //
////////////////////////////////////////////////////////////////

// isCompressed returns whether or not the instruction is compressed or not.
// In the 32-bit instructions, the lowest-order 2 bits are always set, whereas in the compressed 16-bit instruction set,
// this is not the case.
func isCompressed(instr U64) bool {
	return and64(instr, toU64(3)) != toU64(3)
}

// switchKeyC returns the switch key for the passed compressed instruction. The switch key is a 5-bit value that allows
// for matching `C` extension instructions in a single switch-case.
func switchKeyC(instr U64) (key U64) {
	return or64(
		and64(shr64(toU64(11), instr), toU64(0x1C)),
		and64(instr, toU64(3)),
	)
}

// mapCompressedRegister maps a compressed register to its 32-bit counterpart. RVC uses the following register aliases
// to fit the compressed register number within 3 bits:
// ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐
// │ 000 │ 001 │ 010 │ 011 │ 100 │ 101 │ 110 │ 111 │
// ├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤
// │ x8  │ x9  │ x10 │ x11 │ x12 │ x13 │ x14 │ x15 │
// └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘
func mapCompressedRegister(register U64) (decompressedReg U64) {
	return register + C_REGISTER_OFFSET
}

////////////////////////////////////////////////////////////////
//                   DECODING - C EXTENSION                   //
////////////////////////////////////////////////////////////////

// parseFunct3C pulls a 3-bit function out of the high-order bits of the passed compressed instruction.
func parseFunct3C(instr U64) (funct3 U64) {
	return shr64(toU64(13), instr)
}

// decodeCIW pulls the `immediate` and `reg` out of a CIW formatted instruction from the RVC extension.
func decodeCIW(instr U64) (immediate, reg U64) {
	immediate = and64(shr64(toU64(5), instr), toU64(0xFF))
	reg = mapCompressedRegister(and64(shr64(toU64(2), instr), toU64(0x07)))
	return immediate, reg
}

// decodeCJ pulls the `immediate` out of a CJ formatted instruction from the RVC extension.
func decodeCJ(instr U64) (immediate U64) {
	return and64(shr64(toU64(2), instr), U64(0x7FF))
}

// decodeCLCS pulls the `immediate`, `reg1`, and `reg2` out of a CL or CS formatted instruction from the RVC extension.
// NOTE: In the case of the CL format, `reg2` is `rd`, whereas for the CS format, `reg2` is `rs2`.
func decodeCLCS(instr U64) (immediate, reg1, reg2 U64) {
	immediate = or64(
		and64(shr64(toU64(8), instr), toU64(0x1C)),
		and64(shr64(toU64(5), instr), toU64(3)),
	)
	reg1 = mapCompressedRegister(and64(shr64(toU64(7), instr), toU64(7)))
	reg2 = mapCompressedRegister(and64(shr64(toU64(2), instr), toU64(7)))
	return immediate, reg1, reg2
}

// decodeCB pulls the `immediate` and `reg` out of a CB formatted instruction from the RVC extension.
func decodeCB(instr U64) (immediate, reg U64) {
	immediate = or64(
		and64(shr64(toU64(5), instr), toU64(0xE0)),
		and64(shr64(toU64(2), instr), toU64(0x1F)),
	)
	reg = mapCompressedRegister(and64(shr64(toU64(7), instr), toU64(7)))
	return immediate, reg
}

// decodeShiftCB
func decodeShiftCB(instr U64) (shamt, reg U64) {
	shamt = or64(
		shr64(toU64(7), and64(instr, U64(0x1000))),
		and64(toU64(0x1F), shr64(toU64(2), instr)),
	)
	reg = mapCompressedRegister(and64(shr64(toU64(7), instr), toU64(7)))
	return shamt, reg
}

// decodeCI pulls the `immediate` and `reg` out of a CI formatted instruction from the RVC extension.
func decodeCI(instr U64) (immediate, reg U64) {
	immediate = or64(
		and64(shr64(toU64(7), instr), toU64(0x20)),
		and64(shr64(toU64(2), instr), toU64(0x1F)),
	)
	reg = and64(shr64(toU64(7), instr), toU64(0x1F))
	return immediate, reg
}

// decodeCR pulls the `reg1` and `reg2` out of a CR formatted instruction from the RVC extension.
func decodeCR(instr U64) (reg1, reg2 U64) {
	reg1 = and64(shr64(toU64(7), instr), toU64(0x1F))
	reg2 = and64(shr64(toU64(2), instr), toU64(0x1F))
	return reg1, reg2
}

// decodeCSS pulls the `immediate` and `reg` out of a CSS formatted instruction from the RVC extension.
func decodeCSS(instr U64) (immediate, reg U64) {
	immediate = and64(shr64(toU64(7), instr), toU64(0x3F))
	reg = and64(shr64(toU64(2), instr), toU64(0x1F))
	return immediate, reg
}

////////////////////////////////////////////////////////////////
//                    ENCODING - BASE ISA                     //
////////////////////////////////////////////////////////////////

// bitSlice returns a slice of bits in the `in` U64 from [start, end] (inclusive)
func bitSlice(in U64, start U64, end U64) (part U64) {
	in = and64(in, shl64(start, 0xFFFFFFFF_FFFFFFFF))
	in = and64(in, shr64(sub64(toU64(63), end), 0xFFFFFFFF_FFFFFFFF))
	return in >> start
}

// recodeRType re-encodes parameters into an R-type instruction
func recodeRType(opcode, rd, funct3, rs1, rs2, funct7 U64) (instr U64) {
	return or64(
		or64(
			and64(opcode, toU64(0x7F)),
			shl64(toU64(7), and64(rd, toU64(0x1F))),
		),
		or64(
			or64(
				shl64(toU64(12), and64(funct3, toU64(0x07))),
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
			),
			or64(
				shl64(toU64(20), and64(rs2, toU64(0x1F))),
				shl64(toU64(25), and64(funct7, toU64(0x7F))),
			),
		),
	)
}

// recodeIType re-encodes parameters into an I-type instruction.
func recodeIType(opcode, rd, funct3, rs1, immediate U64) (instr U64) {
	return or64(
		and64(opcode, toU64(0x7F)),
		or64(
			or64(
				shl64(toU64(7), and64(rd, toU64(0x1F))),
				shl64(toU64(12), and64(funct3, toU64(0x07))),
			),
			or64(
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
				shl64(toU64(20), and64(immediate, U64(0xFFF))),
			),
		),
	)
}

// recodeSType re-encodes parameters into an S-type instruction
func recodeSType(opcode, funct3, rs1, rs2, immediate U64) (instr U64) {
	return or64(
		or64(
			and64(opcode, toU64(0x7F)),
			shl64(toU64(7), and64(immediate, toU64(0x1F))),
		),
		or64(
			or64(
				shl64(toU64(12), and64(funct3, 0x07)),
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
			),
			or64(
				shl64(toU64(20), and64(rs2, toU64(0x1F))),
				shl64(toU64(25), bitSlice(immediate, 5, 11)),
			),
		),
	)
}

// recodeBType re-encodes parameters into a B-type instruction.
func recodeBType(opcode, funct3, rs1, rs2, immediate U64) (instr U64) {
	immLeft := or64(
		bitSlice(immediate, 11, 11),
		shl64(toU64(1), bitSlice(immediate, 1, 4)),
	)
	immRight := or64(
		bitSlice(immediate, 5, 10),
		shl64(toU64(6), bitSlice(immediate, 12, 12)),
	)
	return or64(
		or64(
			and64(opcode, toU64(0x7F)),
			shl64(7, immLeft),
		),
		or64(
			or64(
				shl64(toU64(12), and64(funct3, 0x07)),
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
			),
			or64(
				shl64(toU64(20), and64(rs2, toU64(0x1F))),
				shl64(25, immRight),
			),
		),
	)
}

// recodeUType re-encodes parameters into a U-type instruction.
func recodeUType(opcode, rd, immediate U64) (instr U64) {
	return or64(
		and64(opcode, toU64(0x7F)),
		or64(
			shl64(toU64(7), and64(rd, toU64(0x1F))),
			and64(immediate, shl64(12, U64(0xFFFFF))),
		),
	)
}

// recodeJType re-encodes parameters into a J-type instruction.
func recodeJType(opcode, rd, immediate U64) (instr U64) {
	twiddledImmediate := or64(
		or64(
			bitSlice(immediate, 12, 19),
			shl64(toU64(8), bitSlice(immediate, 11, 11)),
		),
		or64(
			shl64(toU64(9), bitSlice(immediate, 1, 10)),
			shl64(toU64(19), bitSlice(immediate, 20, 20)),
		),
	)
	return or64(
		and64(opcode, toU64(0x7F)),
		or64(
			shl64(toU64(7), and64(rd, toU64(0x1F))),
			shl64(toU64(12), twiddledImmediate),
		),
	)
}
