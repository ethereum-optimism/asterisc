package fast

import (
	"fmt"
)

func Step(s *VMState) {
	if s.Exited {
		return
	}

	sysCall := func() {
		a7 := s.loadRegister(toU64(17))

		switch a7 {
		case 93: // exit
			a0 := s.loadRegister(toU64(0))
			s.Exit = a0
			s.Exited = true
		case 214: // brk
			// Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

			// brk(0) changes nothing about the memory, and returns the current page break
			v := shl64(toU64(30), toU64(1)) // set program break at 1 GiB
			s.writeRegister(toU64(10), v)
		case 222: // mmap
			// A0 = addr (hint)
			addr := s.loadRegister(toU64(10))
			// A1 = n (length)
			length := s.loadRegister(toU64(11))
			// A2 = prot (memory protection type, can ignore)
			// A3 = flags (shared with other process and or written back to file, can ignore)  // TODO maybe assert the MAP_ANONYMOUS flag is set
			// A4 = fd (file descriptor, can ignore because we support anon memory only)
			// A5 = offset (offset in file, we don't support any non-anon memory, so we can ignore this)

			// ignore: prot, flags, fd, offset
			switch addr {
			case 0:
				// no hint, allocate it ourselves, by as much as the requested length
				s.Heap += length // increment heap with length
				s.writeRegister(toU64(10), s.Heap)
			default:
				// allow hinted memory address (leave it in A0 as return argument)
			}
			s.writeRegister(toU64(11), toU64(0)) // no error
		default:
			// TODO maybe revert if the syscall is unrecognized?
		}
	}

	pc := s.PC
	instr := s.loadMem(pc, 4, false) // raw instruction

	// these fields are ignored if not applicable to the instruction type / opcode
	opcode := parseOpcode(instr)
	rd := parseRd(instr) // destination register index
	funct3 := parseFunct3(instr)
	rs1 := parseRs1(instr) // source register 1 index
	rs2 := parseRs2(instr) // source register 2 index
	funct7 := parseFunct7(instr)
	rs1Value := s.loadRegister(rs1) // loaded source registers. Only load if rs1/rs2 are not zero.
	rs2Value := s.loadRegister(rs2)

	//fmt.Printf("fast PC: %x\n", pc)
	//fmt.Printf("fast INSTR: %x\n", instr)
	//fmt.Printf("fast OPCODE: %x\n", opcode)
	//fmt.Printf("fast rs1 value: %x\n", rs1Value)
	//fmt.Printf("fast rs2 value: %x\n", rs2Value)

	switch opcode {
	case 0x03: // 000_0011: memory loading
		// LB, LH, LW, LD, LBU, LHU, LWU
		imm := parseImmTypeI(instr)
		signed := iszero64(and64(funct3, toU64(4)))      // 4 = 100 -> bitflag
		size := shl64(and64(funct3, toU64(3)), toU64(1)) // 3 = 11 -> 1, 2, 4, 8 bytes size
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		rdValue := s.loadMem(memIndex, size, signed)
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x23: // 010_0011: memory storing
		// SB, SH, SW, SD
		imm := parseImmTypeS(instr)
		size := shl64(funct3, toU64(1))
		value := rs2Value
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		s.storeMem(memIndex, size, value)
		s.setPC(add64(pc, toU64(4)))
	case 0x63: // 110_0011: branching
		branchHit := toU64(0)
		switch funct3 {
		case 0: // 000 = BEQ
			branchHit = eq64(rs1Value, rs2Value)
		case 1: // 001 = BNE
			branchHit = and64(not64(eq64(rs1Value, rs2Value)), toU64(1))
		case 4: // 100 = BLT
			branchHit = slt64(rs1Value, rs2Value)
		case 5: // 101 = BGE
			branchHit = and64(not64(slt64(rs1Value, rs2Value)), toU64(1))
		case 6: // 110 = BLTU
			branchHit = lt64(rs1Value, rs2Value)
		case 7: // 111 = BGEU
			branchHit = and64(not64(lt64(rs1Value, rs2Value)), toU64(1))
		}
		switch branchHit {
		case 0:
			pc = add64(pc, toU64(4))
		default:
			imm := parseImmTypeB(instr)
			// imm12 is a signed offset, in multiples of 2 bytes
			pc = add64(pc, signExtend64(imm, toU64(11)))
		}
		// not like the other opcodes: nothing to write to rd register, and PC has already changed
		s.setPC(pc)
	case 0x13: // 001_0011: immediate arithmetic and logic
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3 {
		case 0: // 000 = ADDI
			rdValue = add64(rs1Value, imm)
		case 1: // 001 = SLLI
			rdValue = shl64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
		case 2: // 010 = SLTI
			rdValue = slt64(rs1Value, imm)
		case 3: // 011 = SLTIU
			rdValue = lt64(rs1Value, imm)
		case 4: // 100 = XORI
			rdValue = xor64(rs1Value, imm)
		case 5: // 101 = SR~
			switch funct7 {
			case 0x00: // 0000000 = SRLI
				rdValue = shr64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
			case 0x20: // 0100000 = SRAI
				rdValue = sar64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
			}
		case 6: // 110 = ORI
			rdValue = or64(rs1Value, imm)
		case 7: // 111 = ANDI
			rdValue = and64(rs1Value, imm)
		}
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x1B: // 001_1011: immediate arithmetic and logic signed 32 bit
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3 {
		case 0: // 000 = ADDIW
			rdValue = mask32Signed64(add64(rs1Value, imm))
		case 1: // 001 = SLLIW
			rdValue = mask32Signed64(shl64(and64(imm, toU64(0x1F)), rs1Value))
		case 5: // 101 = SR~
			shamt := and64(imm, toU64(0x1F))
			switch funct7 {
			case 0x00: // 0000000 = SRLIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
			case 0x20: // 0100000 = SRAIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
			}
		}
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x33: // 011_0011: register arithmetic and logic
		var rdValue U64
		switch funct7 {
		case 1: // RV M extension
			switch funct3 {
			case 0: // 000 = MUL: signed x signed
				rdValue = mul64(rs1Value, rs2Value)
			case 1: // 001 = MULH: upper bits of signed x signed
				rdValue = u256ToU64(shr(toU256(64), mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value))))
			case 2: // 010 = MULHSU: upper bits of signed x unsigned
				rdValue = u256ToU64(shr(toU256(64), mul(signExtend64To256(rs1Value), u64ToU256(rs2Value))))
			case 3: // 011 = MULHU: upper bits of unsigned x unsigned
				rdValue = u256ToU64(shr(toU256(64), mul(u64ToU256(rs1Value), u64ToU256(rs2Value))))
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
					rdValue = sub64(rs1Value, rs2Value)
				}
			case 1: // 001 = SLL
				rdValue = shl64(and64(rs2Value, toU64(0x3F)), rs1Value) // only the low 6 bits are consider in RV6VI
			case 2: // 010 = SLT
				rdValue = slt64(rs1Value, rs2Value)
			case 3: // 011 = SLTU
				rdValue = lt64(rs1Value, rs2Value)
			case 4: // 100 = XOR
				rdValue = xor64(rs1Value, rs2Value)
			case 5: // 101 = SR~
				switch funct7 {
				case 0x00: // 0000000 = SRL
					rdValue = shr64(and64(rs2Value, toU64(0x3F)), rs1Value) // logical: fill with zeroes
				case 0x20: // 0100000 = SRA
					rdValue = sar64(and64(rs2Value, toU64(0x3F)), rs1Value) // arithmetic: sign bit is extended
				}
			case 6: // 110 = OR
				rdValue = or64(rs1Value, rs2Value)
			case 7: // 111 = AND
				rdValue = and64(rs1Value, rs2Value)
			}
		}
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x3B: // 011_1011: register arithmetic and logic in 32 bits
		var rdValue U64
		switch funct7 {
		case 1: // RV M extension
			switch funct3 {
			case 0: // 000 = MULW
				rdValue = mask32Signed64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
			case 4: // 100 = DIVW
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = mask32Signed64(sdiv64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
				}
			case 5: // 101 = DIVUW
				switch rs2Value {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = mask32Signed64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				}
			case 6: // 110 = REMW
				switch rs2Value {
				case 0:
					rdValue = mask32Signed64(rs1Value)
				default:
					rdValue = mask32Signed64(smod64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
				}
			case 7: // 111 = REMUW
				switch rs2Value {
				case 0:
					rdValue = mask32Signed64(rs1Value)
				default:
					rdValue = mask32Signed64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				}
			}
		default:
			switch funct3 {
			case 0: // 000 = ADDW/SUBW
				switch funct7 {
				case 0x00: // 0000000 = ADDW
					rdValue = mask32Signed64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				case 0x20: // 0100000 = SUBW
					rdValue = mask32Signed64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				}
			case 1: // 001 = SLLW
				rdValue = mask32Signed64(shl64(and64(rs2Value, toU64(0x1F)), rs1Value))
			case 5: // 101 = SR~
				shamt := and64(rs2Value, toU64(0x1F))
				switch funct7 {
				case 0x00: // 0000000 = SRLW
					rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
				case 0x20: // 0100000 = SRAW
					rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
				}
			}
		}
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x37: // 011_0111: LUI = Load upper immediate
		imm := parseImmTypeU(instr)
		rdValue := shl64(toU64(12), imm)
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x17: // 001_0111: AUIPC = Add upper immediate to PC
		imm := parseImmTypeU(instr)
		rdValue := add64(pc, signExtend64(shl64(toU64(12), imm), toU64(31)))
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x6F: // 110_1111: JAL = Jump and link
		imm := parseImmTypeJ(instr)
		rdValue := add64(pc, toU64(4))
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, signExtend64(imm, toU64(21)))) // signed offset in multiples of 2 bytes (last bit is there, but ignored)
	case 0x67: // 110_0111: JALR = Jump and link register
		imm := parseImmTypeI(instr)
		rdValue := add64(pc, toU64(4))
		s.writeRegister(rd, rdValue)
		s.setPC(and64(add64(rs1Value, signExtend64(imm, toU64(12))), xor64(u64Mask(), toU64(1)))) // least significant bit is set to 0
	case 0x73: // 111_0011: environment things
		switch funct3 {
		case 0: // 000 = ECALL/EBREAK
			switch shr64(toU64(20), instr) { // I-type, top 12 bits
			case 0: // imm12 = 000000000000 ECALL
				sysCall()
				s.setPC(add64(pc, toU64(4)))
			default: // imm12 = 000000000001 EBREAK
				// ignore breakpoint
				s.setPC(add64(pc, toU64(4)))
			}
		default: // CSR instructions
			imm := parseCSSR(instr)
			rdValue := s.readCSR(imm)
			value := rs1
			if iszero64(and64(funct3, toU64(4))) {
				value = rs1Value
			}
			switch and64(funct3, toU64(3)) {
			case 1: // ?01 = CSRRW(I) = "atomic Read/Write bits in CSR"
				s.writeCSR(imm, value)
			case 2: // ?10 = CSRRS = "atomic Read and Set bits in CSR"
				s.writeCSR(imm, or64(rdValue, value)) // v=0 will be no-op
			case 3: // ?11 = CSRRC = "atomic Read and Clear Bits in CSR"
				s.writeCSR(imm, and64(rdValue, not64(value))) // v=0 will be no-op
			}
			// TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
			s.writeRegister(rd, rdValue)
			s.setPC(add64(pc, toU64(4)))
		}
	case 0x2F: // 010_1111: RV32A and RV32A atomic operations extension
		// TODO atomic operations
		// 0b010 == RV32A W variants
		// 0b011 == RV64A D variants
		//size := 1 << funct3
		switch shr64(toU64(2), and64(funct7, toU64(0x1F))) {
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
		//s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x0F: // 000_1111: fence
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
		// We can no-op FENCE, there's nothing to synchronize
		//s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	default:
		panic(fmt.Errorf("unknown opcode: %b full instruction: %b", opcode, instr))
	}
	return
}
