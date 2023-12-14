package fast

import "fmt"

// DecompressInstruction decompresses a 16-bit `C` extension RISC-V instruction into its 32-bit standard counterpart.
// For instructions passed that are not a part of the `C` extension, they are returned as-is. If an unknown compressed
// instruction is passed, an error is returned.
//
// Supported instructions:
// *todo*
func DecompressInstruction(instr U64) (instrOut U64, pcBump U64, err error) {
	// If the instruction is not a `C` extension instruction, return it as-is.
	if !isCompressed(instr) {
		return instr, toU64(4), nil
	}

	opcode := and64(instr, toU64(3))
	funct := parseFunct3C(instr)

	// Switch over the instruction opcode
	switch opcode {
	// C0
	case 0x00:
		// rd := and64(shr64(toU64(2), instr), toU64(7))
		// rs1 := and64(shr64(toU64(7), instr), toU64(7))
		switch funct {
		// CIW - C.ADDI4SPN
		case 0x00:
		// CL - C.LW
		case 0x02:
		// CL - C.LD
		case 0x03:
		// CS - C.SW
		case 0x05:
		// CS - C.SD
		case 0x06:
		}
	// C1
	case 0x01:
		switch funct {
		// CI - C.NOP | C.ADDI
		case 0x00:
		// CI - C.ADDIW
		case 0x01:
		// CI - C.LI
		case 0x02:
		// CI - C.ADDI16SP | C.LUI
		case 0x03:
		// C.SRLI, S.SRLI64, C.SRAI, C.ANDI, C.SUB, C.XOR, C.OR, C.AND, C.SUBW, C.ADDW
		case 0x04:
		// CR - C.J
		case 0x05:
		// CB - C.BEQZ
		case 0x06:
		// CB - C.BNEZ
		case 0x07:
		}
	// C2
	case 0x02:
		switch funct {
		// CI - C.SLLI64
		case 0x00:
		// CI - C.LWSP
		case 0x02:
		// CI - C.LDSP
		case 0x03:
		// C.JR, C.MV, C.EBREAK, C.JALR, C.ADD
		case 0x04:
		// CSS - C.SWSP
		case 0x06:
		// CSS - C.SDSP
		case 0x07:
		}
	}

	// placeholder - remove.
	return 0, 0, nil
}

// isCompressed returns whether or not the instruction is compressed or not.
// In the 32-bit instructions, the lowest-order 2 bits are always set, whereas in the compressed 16-bit instruction set,
// this is not the case.
func isCompressed(instr U64) bool {
	return and64(instr, toU64(3)) != toU64(3)
}

// parseFunct3C pulls a 3-bit function out of the high-order bits of the passed compressed instruction.
func parseFunct3C(instr U64) (funct3 U64) {
	return shr64(toU64(13), instr)
}

// mapCompressedRegister maps a compressed register to its 32-bit counterpart. RVC uses the following register aliases
// to fit the compressed register number within 3 bits:
// ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐
// │ 000 │ 001 │ 010 │ 011 │ 100 │ 101 │ 110 │ 111 │
// ├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤
// │ x8  │ x9  │ x10 │ x11 │ x12 │ x13 │ x14 │ x15 │
// └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘
func mapCompressedRegister(register U64) (U64, error) {
	if register > 7 {
		return 0, fmt.Errorf("invalid compressed register: %x", register)
	}
	return register + 8, nil
}

func convertCSS(instr U64) (U64, error) {
	return 0, nil
}
