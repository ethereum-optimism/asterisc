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
	// funct3 = 0 ++ opcode = 0 from C.ADDI4SPN.
	if instr == 0 {
		return 0, 0, fmt.Errorf("illegal instruction: %x", instr)
	}

	// opcode := and64(instr, toU64(3))
	// funct := parseFunct3C(instr)

	switch instr {
	// C.ADDI4SPN [OP: C0 | Funct3: 000 | Format: CIW]
	case 0x0:
		// TODO: Perform translation to 32 bit analogue.
	// C.NOP, C.ADDI [OP: C1 | Funct3: 000 | Format: CI]
	case 0x1:
		// TODO: Perform translation to 32 bit analogue.
	// C.SLLI64 [OP: C2 | Funct3: 000 | Format: CI]
	case 0x2:
		// TODO: Perform translation to 32 bit analogue.
	// C.FLD [OP: C0 | Funct3: 001 | Format: CL]
	case 0x4:
		// TODO: Perform translation to 32 bit analogue.
	// C.ADDIW [OP: C1 | Funct3: 001 | Format: CI]
	case 0x5:
		// TODO: Perform translation to 32 bit analogue.
	// C.FLDSP (Unsupported) [OP: C2 | Funct3: 001 | Format: CI]
	case 0x6:
		// TODO: Perform translation to 32 bit analogue.
	// C.LW [OP: C0 | Funct3: 010 | Format: CL]
	case 0x8:
		// TODO: Perform translation to 32 bit analogue.
	// C.LI [OP: C1 | Funct3: 010 | Format: CI]
	case 0x9:
		// TODO: Perform translation to 32 bit analogue.
	// C.LWSP [OP: C2 | Funct3: 010 | Format: CI]
	case 0xA:
		// TODO: Perform translation to 32 bit analogue.
	// C.LD [OP: C0 | Funct3: 011 | Format: CL]
	case 0xC:
		// TODO: Perform translation to 32 bit analogue.
	// C.ADDI16SP, C.LUI [OP: C1 | Funct3: 011 | Format: CI]
	case 0xD:
		// TODO: Perform translation to 32 bit analogue.
	// C.LDSP [OP: C2 | Funct3: 011 | Format: CI]
	case 0xE:
		// TODO: Perform translation to 32 bit analogue.
	// Reserved [OP: C0 | Funct3: 100 | Format: ~]
	case 0x10:
		// TODO: Perform translation to 32 bit analogue.
	// C.SRLI, S.SRLI64, C.SRAI64, C.ANDI, C.SUB, C.XOR, C.OR, C.AND, C.SUBW, C.ADDW [OP: C1 | Funct3: 100 | Format: ?]
	case 0x11:
		// TODO: Perform translation to 32 bit analogue.
	// C.JR, C.MV, C.EBREAK, C.JALR, C.ADD [OP: C2 | Funct3: 100 | Format: ?]
	case 0x12:
		// TODO: Perform translation to 32 bit analogue.
	// C.FSD (Unsupported) [OP: C0 | Funct3: 101 | Format: CS]
	case 0x14:
		// TODO: Perform translation to 32 bit analogue.
	// C.J [OP: C1 | Funct3: 101 | Format: CR]
	case 0x15:
		// TODO: Perform translation to 32 bit analogue.
	// C.FSDSP (Unsupported) [OP: C2 | Funct3: 101 | Format: CSS]
	case 0x16:
		// TODO: Perform translation to 32 bit analogue.
	// C.SW [OP: C0 | Funct3: 110 | Format: CS]
	case 0x18:
		// TODO: Perform translation to 32 bit analogue.
	// C.BEQZ [OP: C1 | Funct3: 110 | Format: CB]
	case 0x19:
		// TODO: Perform translation to 32 bit analogue.
	// C.SWSP [OP: C2 | Funct3: 110 | Format: CSS]
	case 0x1A:
		// TODO: Perform translation to 32 bit analogue.
	// C.SD [OP: C0 | Funct3: 111 | Format: CS]
	case 0x1C:
		// TODO: Perform translation to 32 bit analogue.
	// C.BNEZ [OP: C1 | Funct3: 111 | Format: CB]
	case 0x1D:
		// TODO: Perform translation to 32 bit analogue.
	// C.SDSP [OP: C2 | Funct3: 111 | Format: CSS]
	case 0x1E:
		// TODO: Perform translation to 32 bit analogue.
	default:
		return 0, 0, fmt.Errorf("unknown instruction: %x", instr)
	}

	return 0, 0, fmt.Errorf("unknown instruction: %x", instr)
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
