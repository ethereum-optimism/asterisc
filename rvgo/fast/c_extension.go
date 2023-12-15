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
//
// Supported instructions:
// *todo*
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
		decompressedInstr = encodeIType(
			0b0010011, // Arithmetic
			0,         // ADDI
			reg,       // rs1
			SP,        // rd - SP
			imm,       // immediate
		)
	// C.NOP, C.ADDI [OP: C1 | Funct3: 000 | Format: CI]
	case 0x1:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = encodeIType(
			0b0010011, // Arithmetic
			0,         // ADDI
			reg,       // rd
			reg,       // rs1
			imm,       // immediate
		)
	// C.SLLI64 [OP: C2 | Funct3: 000 | Format: CI]
	case 0x2:
		imm, reg := decodeCI(instr)
		imm = and64(imm, toU64(0x1F))
		decompressedInstr = encodeIType(
			0b0010011, // Arithmetic
			1,         // SLLI - funct3
			reg,       // rd
			reg,       // rs1
			imm,       // shamt

		)
	// C.FLD (Unsupported) [OP: C0 | Funct3: 001 | Format: CL]
	case 0x4:
		panic("unsupported")
	// C.ADDIW [OP: C1 | Funct3: 001 | Format: CI]
	case 0x5:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = encodeIType(
			0b0011011, // Arithmetic (RV64I)
			0,         // ADDIW
			reg,       // rd
			reg,       // rs1
			imm,       // immediate
		)
	// C.FLDSP (Unsupported) [OP: C2 | Funct3: 001 | Format: CI]
	case 0x6:
		panic("unsupported")
	// C.LW [OP: C0 | Funct3: 010 | Format: CL]
	case 0x8:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = shl64(toU64(1), and64(
			or64(shl64(toU64(5), imm), imm),
			toU64(0x3E),
		))
		decompressedInstr = encodeIType(
			0b0000011, // Load
			0b010,     // LW
			reg2,      // rd
			reg1,      // rs1
			imm,       // immediate
		)
	// C.LI [OP: C1 | Funct3: 010 | Format: CI]
	case 0x9:
		imm, reg := decodeCI(instr)
		imm = signExtend64(imm, toU64(5))
		decompressedInstr = encodeIType(
			0b0010011, // Arithmetic
			0,         // ADDI
			reg,       // rd
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
		decompressedInstr = encodeIType(
			0b0000011, // Load
			0b010,     // LW
			reg,       // rd
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
		decompressedInstr = encodeIType(
			0b0000011, // Load
			0b011,     // LD
			reg2,      // rd
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
			decompressedInstr = encodeIType(
				0b0010011, // Arithmetic
				0,         // ADDI
				SP,        // rd - SP
				SP,        // rs1 - SP
				imm,       // immediate
			)
		} else {
			// C.LUI
			imm = signExtend64(shl64(toU64(12), imm), 17)
			// TODO: Immediate re-formatting?
			decompressedInstr = encodeUJType(
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
		decompressedInstr = encodeIType(
			0b0000011, // Load
			0b011,     // LD
			reg,       // rd
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
		// C.SRAI
		case 0x01:
		// C.ANDI
		case 0x02:
		}

		// _, reg1, reg2 := decodeCLCS(instr)
		switch (instr >> 8 & 0x1C) | (instr >> 5 & 0x03) {
		// C.SUB
		case 0x0C:
		// C.XOR
		case 0x0D:
		// C.OR
		case 0x0E:
		// C.AND
		case 0x0F:
		// C.SUBW
		case 0x1C:
		// C.ADDW
		case 0x1D:
		// Reserved
		case 0x1E:
		case 0x1F:
		}
		panic("unsupported")
	// C.JR, C.MV, C.EBREAK, C.JALR, C.ADD [OP: C2 | Funct3: 100 | Format: ?]
	case 0x12:
		reg1, reg2 := decodeCR(instr)
		fnsel := instr & 0x1000
		switch {
		// C.JR
		case fnsel == 0 && reg2 == 0:
			// TODO
			panic("unimplemented")
		// C.MV
		case fnsel == 0:
			decompressedInstr = encodeRSBType(
				0b0110011, // Arithmetic
				0,         // funct3
				0,         // funct7 - ADD
				reg1,      // rd
				0,         // rs1
				reg2,      // rs2
			)
		// C.EBREAK
		case fnsel == 0x1000 && reg1 == 0 && reg2 == 0:
			decompressedInstr = 0b00000000000100000000000001110011
		// C.JALR
		case fnsel == 0x1000 && reg2 == 0:
			// TODO
			panic("unimplemented")
		// C.ADD
		default:
			decompressedInstr = encodeRSBType(
				0b0110011, // Arithmetic
				0,         // funct3
				0,         // funct7 - ADD
				reg1,      // rd
				reg1,      // rs1
				reg2,      // rs2
			)
		}
	// C.FSD (Unsupported) [OP: C0 | Funct3: 101 | Format: CS]
	case 0x14:
		panic("unsupported")
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
		// TODO: Custom handler for C.J
		panic("unimplemented")
	// C.FSDSP (Unsupported) [OP: C2 | Funct3: 101 | Format: CSS]
	case 0x16:
		panic("unsupported")
	// C.SW [OP: C0 | Funct3: 110 | Format: CS]
	case 0x18:
		imm, reg1, reg2 := decodeCLCS(instr)
		imm = and64(
			shl64(toU64(1), or64(shl64(toU64(5), imm), imm)),
			toU64(0x7C),
		)
		decompressedInstr = encodeRSBType(
			0b0100011, // Store
			0b010,     // SW
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			reg1,      // rs1
			reg2,      // rs2
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
		decompressedInstr = encodeRSBType(
			0b1100011, // Branch
			0,         // BEQ
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			reg,       // rs1
			0,         // rs2 - ZERO
		)
	// C.SWSP [OP: C2 | Funct3: 110 | Format: CSS]
	case 0x1A:
		imm, reg := decodeCSS(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			toU64(0xFC),
		)
		decompressedInstr = encodeRSBType(
			0b0100011, // Store
			0b010,     // SW
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			SP,        // rs1
			reg,       // rs2
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
		decompressedInstr = encodeRSBType(
			0b0100011, // Store
			0b011,     // SD
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			reg1,      // rs1
			reg2,      // rs2
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
		decompressedInstr = encodeRSBType(
			0b1100011, // Branch
			1,         // BNE
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			reg,       // rs1
			0,         // rs2
		)
	// C.SDSP [OP: C2 | Funct3: 111 | Format: CSS]
	case 0x1E:
		imm, reg := decodeCSS(instr)
		imm = and64(
			or64(shl64(toU64(6), imm), imm),
			U64(0x1F8),
		)
		decompressedInstr = encodeRSBType(
			0b0100011, // Store
			0b011,     // SD
			0,         // placeholder - immediate part
			0,         // placeholder - immediate part
			SP,        // rs1
			reg,       // rs2
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

// encodeRSBType encodes parameters into an R-type, S-type, or U-type instruction.
func encodeRSBType(opcode, funct3, funct7, rd, rs1, rs2 U64) (instr U64) {
	return or64(
		or64(
			or64(
				and64(opcode, toU64(0x7F)),
				shl64(toU64(7), and64(rd, toU64(0x1F))),
			),
			or64(
				shl64(toU64(12), and64(funct3, toU64(7))),
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
			),
		),
		or64(
			shl64(toU64(20), and64(rs2, toU64(0x1F))),
			shl64(toU64(25), and64(funct7, toU64(0x7F))),
		),
	)
}

// encodeIType encodes parameters into an I-type instruction.
func encodeIType(opcode, funct3, rd, rs1, immediate U64) (instr U64) {
	return or64(
		and64(opcode, toU64(0x7F)),
		or64(
			or64(
				shl64(toU64(7), and64(rd, toU64(0x1F))),
				shl64(toU64(12), and64(funct3, toU64(7))),
			),
			or64(
				shl64(toU64(15), and64(rs1, toU64(0x1F))),
				shl64(toU64(20), and64(immediate, U64(0x7FF))),
			),
		),
	)
}

// encodeUJType encodes parameters into a U-type or J-type instruction.
func encodeUJType(opcode, rd, immediate U64) (instr U64) {
	return or64(
		and64(opcode, toU64(0x7F)),
		or64(
			shl64(toU64(7), and64(rd, toU64(0x1F))),
			shl64(toU64(12), and64(immediate, U64(0xFFFFF))),
		),
	)
}
