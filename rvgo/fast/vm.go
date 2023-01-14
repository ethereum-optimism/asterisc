package fast

import "fmt"

func Step(s *VMState) {
	// memory operation arguments
	var memIndex U64
	var size U64
	var signed bool
	var value U64

	var rdValue U64 // value that is being written this instruction

	pc := s.PC
	instr := s.loadMem(pc, 4, false) // raw instruction
	// these fields are ignored if not applicable to the instruction type / opcode
	opcode := parseOpcode(instr)
	rd := parseRd(instr) // destination register index
	funct3 := parseFunct3(instr)
	rs1 := parseRs1(instr) // source register 1 index
	rs2 := parseRs2(instr) // source register 2 index
	funct7 := parseFunct7(instr)
	rs1Value := s.Registers[rs1] // loaded source registers. Only load if rs1/rs2 are not zero.
	rs2Value := s.Registers[rs2]

	switch opcode {
	case 0b0000011: // memory loading
		// LB, LH, LW, LD, LBU, LHU, LWU
		imm := parseImmTypeI(instr)
		signed = iszero64(and64(funct3, toU64(8)))      // 8 = 100 -> bitflag
		size = shl64(toU64(1), and64(funct3, toU64(3))) // 3 = 11 -> 1, 2, 4, 8 bytes size
		memIndex = add64(rs1Value, signExtend64(imm, toU64(11)))
		rdValue = s.loadMem(memIndex, size, signed)
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0100011: // memory storing
		// SB, SH, SW, SD
		imm := parseImmTypeS(instr)
		size = shl64(toU64(1), funct3)
		value = rs2Value
		memIndex = add64(rs1Value, signExtend64(imm, toU64(11)))
		s.storeMem(memIndex, size, value)
		pc = add64(pc, 4)
		s.PC = pc
	case 0b1100011: // branching
		branchHit := toU64(0)
		switch funct3 {
		case 0: // 000 = BEQ
			branchHit = eq64(rs1Value, rs2Value)
		case 1: // 001 = BNE
			branchHit = not64(eq64(rs1Value, rs2Value))
		case 4: // 100 = BLT
			branchHit = slt64(rs1Value, rs2Value)
		case 5: // 101 = BGE
			branchHit = not64(slt64(rs1Value, rs2Value))
		case 6: // 110 = BLTU
			branchHit = lt64(rs1Value, rs2Value)
		case 7: // 111 = BGEU
			branchHit = not64(lt64(rs1Value, rs2Value))
		}
		switch branchHit {
		case 0:
			pc = add64(pc, toU64(4))
		default:
			imm := parseImmTypeB(instr)
			// imm12 is a signed offset, in multiples of 2 bytes
			pc = add64(pc, shl64(signExtend64(imm, toU64(11)), toU64(1)))
		}
		// not like the other opcodes: nothing to write to rd register, and PC has already changed
		s.PC = pc
	case 0b0010011: // immediate arithmetic and logic
		imm := parseImmTypeI(instr)
		switch funct3 {
		case 0: // 000 = ADDI
			rdValue = add64(rs1Value, signExtend64(imm, toU64(11)))
		case 1: // 001 = SLLI
			rdValue = shl64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
		case 2: // 010 = SLTI
			rdValue = slt64(rs1Value, imm)
		case 3: // 011 = SLTIU
			rdValue = lt64(rs1Value, imm)
		case 4: // 100 = XORI
			rdValue = xor64(rs1Value, imm)
		case 5: // 101 = SR~
			switch funct7 {
			case 0x00: // 0000000 = SRLI
				rdValue = shr64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
			case 0x20: // 0100000 = SRAI
				rdValue = sar64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
			}
		case 6: // 110 = ORI
			rdValue = or64(rs1Value, imm)
		case 7: // 111 = ANDI
			rdValue = and64(rdValue, imm)
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0011011: // immediate arithmetic and logic signed 32 bit
		imm := parseImmTypeI(instr)
		switch funct3 {
		case 0: // 000 = ADDIW
			rdValue = add64(rs1Value, imm)
		case 1: // 001 = SLLIW
			rdValue = shl64(rs1Value, and64(imm, toU64(0x1F)))
		case 5: // 101 = SR~
			shamt := and64(imm, toU64(0x1F))
			switch funct7 {
			case 0x00: // 0000000 = SRLIW
				rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
			case 0x20: // 0100000 = SRAIW
				rdValue = signExtend64(sar64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
			}
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0110011: // register arithmetic and logic
		switch funct7 {
		case 1: // RV32M extension
			switch funct3 {
			case 0: // 000 = MUL: signed x signed
				rdValue = mul64(rs1Value, rs2Value)
			case 1: // 001 = MULH: upper bits of signed x signed
				rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value)), toU256(64)))
			case 2: // 010 = MULHSU: upper bits of signed x unsigned
				rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
			case 3: // 011 = MULHU: upper bits of unsigned x unsigned
				rdValue = u256ToU64(shr(mul(u64ToU256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
			case 4: // 100 = DIV
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = sdiv64(rs1Value, rs2Value)
				}
			case 5: // 101 = DIVU
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = div64(rs1Value, rs2Value)
				}
			case 6: // 110 = REM
				switch rs2Value {
				case 0:
					rdValue = rs1Value
				default:
					rdValue = smod64(rs1Value, rs2Value)
				}
			case 7: // 111 = REMU
				switch rs2Value {
				case 0:
					rdValue = rs1Value
				default:
					rdValue = mod64(rs1Value, rs2Value)
				}
			}
		default:
			switch funct3 {
			case 0: // 000 = ADD/SUB
				switch funct7 {
				case 0x00: // 0000000 = ADD
					rdValue = add64(rs1Value, rs2Value)
				case 0x20: // 0100000 = SUB
					rdValue = sub64(rs1Value, rs1Value)
				}
			case 1: // 001 = SLL
				rdValue = shl64(rs1Value, and64(rs2Value, toU64(0x3F))) // only the low 6 bits are consider in RV6VI
			case 2: // 010 = SLT
				rdValue = slt64(rs1Value, rs2Value)
			case 3: // 011 = SLTU
				rdValue = lt64(rs1Value, rs2Value)
			case 4: // 100 = XOR
				rdValue = xor64(rs1Value, rdValue)
			case 5: // 101 = SR~
				switch funct7 {
				case 0x00: // 0000000 = SRL
					rdValue = shr64(rs1Value, and64(rs2Value, toU64(0x3F))) // logical: fill with zeroes
				case 0x20: // 0100000 = SRA
					rdValue = sar64(rs1Value, and64(rs2Value, toU64(0x3F))) // arithmetic: sign bit is extended
				}
			case 6: // 110 = OR
				rdValue = or64(rs1Value, rs2Value)
			case 7: // 111 = AND
				rdValue = and64(rs1Value, rs2Value)
			}
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0111011: // register arithmetic and logic in 32 bits
		switch funct7 {
		case 1: // RV64M extension
			switch funct3 {
			case 0: // 000 = MULW
				rdValue = signExtend64(and64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
			case 4: // 100 = DIVW
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = signExtend64(and64(sdiv64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
				}
			case 5: // 101 = DIVUW
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = and64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
				}
			case 6: // 110 = REMW
				switch rs2Value {
				case 0:
					rdValue = and64(rs1Value, u32Mask())
				default:
					rdValue = and64(smod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
				}
			case 7: // 111 = REMUW
				switch rs2Value {
				case 0:
					rdValue = and64(rs1Value, u32Mask())
				default:
					rdValue = and64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
				}
			}
		default: // RV32M extension
			switch funct3 {
			case 0: // 000 = ADDW/SUBW
				switch funct7 {
				case 0x00: // 0000000 = ADDW
					rdValue = signExtend64(and64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
				case 0x20: // 0100000 = SUBW
					rdValue = signExtend64(and64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
				}
			case 1: // 001 = SLLW
				rdValue = signExtend64(and64(shl64(rs1Value, and64(rs2Value, toU64(0x1F))), u32Mask()), toU64(31))
			case 5: // 101 = SR~
				shamt := and64(rs2Value, toU64(0x1F))
				switch funct7 {
				case 0x00: // 0000000 = SRLW
					rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
				case 0x20: // 0100000 = SRAW
					rdValue = signExtend64(sar64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
				}
			}
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0110111: // LUI = Load upper immediate
		imm := parseImmTypeU(instr)
		rdValue = shl64(imm, toU64(12))
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0010111: // AUIPC = Add upper immediate to PC
		imm := parseImmTypeU(instr)
		rdValue = add64(pc, shl64(imm, toU64(12)))
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b1101111: // JAL = Jump and link
		imm := parseImmTypeJ(instr)
		rdValue = add64(pc, toU64(4))
		pc = add64(pc, signExtend64(shl64(imm, toU64(1)), toU64(21))) // signed offset in multiples of 2 bytes
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b1100111: // JALR = Jump and link register
		imm := parseImmTypeI(instr)
		rdValue = add64(pc, toU64(4))
		pc = and64(add64(rs1Value, signExtend64(imm, toU64(12))), xor64(u64Mask(), toU64(1))) // least significant bit is set to 0
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b1110011:
		switch funct3 {
		case 0: // 000 = ECALL/EBREAK
			// TODO: I type instruction
			//000000000000 1110011 ECALL
			//000000000001 1110011 EBREAK
		case 1: // 001 = CSRRW
		case 2: // 010 = CSRRS  a.k.a. SYSTEM instruction
			// TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
		case 3: // 011 = CSRRC
		case 5: // 101 = CSRRWI
		case 6: // 110 = CSRRSI
		case 7: // 111 = CSRRCI
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0101111: // RV32A and RV32A atomic operations extension
		// TODO atomic operations
		// 0b010 == RV32A W variants
		// 0b011 == RV64A D variants
		//size := 1 << funct3
		switch shr64(and64(funct7, toU64(0x1F)), toU64(2)) {
		case 0x0: // 00000 = AMOADD
		case 0x1: // 00001 = AMOSWAP
		case 0x2: // 00010 = LR
		case 0x3: // 00011 = SC
		case 0x4: // 00100 = AMOXOR
		case 0x8: // 01000 = AMOOR
		case 0xc: // 01100 = AMOAND
		case 0x10: // 10000 = AMOMIN
		case 0x14: // 10100 = AMOMAX
		case 0x18: // 11000 = AMOMINU
		case 0x1c: // 11100 = AMOMAXU
		}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	case 0b0001111:
		//// TODO: different layout of func data
		//// "fm pred succ"
		//switch funct3 {
		//case 0b000:
		//	switch funct7 {
		//	case 0b1000001: // FENCE.TSO
		//	default: // FENCE
		//	}
		//case 0b001: // FENCE.I
		//}
		pc = add64(pc, toU64(4))
		s.Registers[rd] = rdValue
		s.PC = pc
	default:
		panic(fmt.Errorf("unknown opcode: %b full instruction: %b", opcode, instr))
	}
}
