package rvgo

import (
	"encoding/binary"
	"math/big"
)

type VMScratchpad struct {
	SubIndex uint64 // step in the instruction execution
	PC       uint64 // PC counter

	Instr uint64 // raw instruction

	Opcode uint64 // parsed instruction
	Funct3 uint64 //
	Funct7 uint64 //

	Rd  uint64 // destination register index
	Rs1 uint64 // source register 1 index
	Rs2 uint64 // source register 2 index

	RdValue  uint64 // destination register value to write
	Rs1Value uint64 // loaded source registers. Only load if rs1/rs2 are not zero.
	Rs2Value uint64 //

	StateRoot [32]byte // commits to state as structured binary tree

	// State machine
	StateGindex      uint64   // target // TODO: might have to be uint256 if memory space is 64 bits large
	StateStackHash   [32]byte // hash of previous traversed stack and last stack element
	StateStackGindex uint64   // to navigate the state loading/writing // TODO uint256
	StateStackDepth  uint8
	StateValue       [32]byte

	Scratch         [8]byte
	InstrPhaseIndex uint8
}

// sub index steps
const (
	StepLoadPC     = iota // N steps to load PC
	StepLoadInstr         // N steps to load instruction at memory from PC
	StepParseInstr        // 1 step to parse instruction
	StepLoadRs1           // N steps to load rs1
	StepLoadRs2           // N steps to load rs2
	StepOpcode            // N steps to execute opcode
	StepWriteRd           // N steps to write rd
	StepWritePC           // N steps to write PC
	StepFinal             // cleanup step
)

const (
	StateStartFirst = iota
	StateEndFirst
	StateStartSecond
	StateEndSecond
)

func makeMemGindex(byteIndex uint64) uint64 {
	if byteIndex >= (1 << 64) {
		panic("only support 64 bit memory, need larger gindex values")
	}
	return (memoryGindex << (64 - 5)) | (byteIndex >> 5)
}

func makeRegisterGindex(register uint64) uint64 {
	if register >= 32 {
		panic("there are only 32 valid registers")
	}
	return (registersGindex << 5) | uint64(register)
}

func Step(s VMScratchpad, so VMStateOracle) VMScratchpad {
	signExtend := func(v uint64, bit uint64) uint64 {
		switch and(v, shl(1, bit)) {
		case 0:
			// fill with zeroes, by masking
			return and(v, shr(0xFFFF_FFFF_FFFF_FFFF, sub(63, bit)))
		default:
			// fill with ones, by or-ing
			return or(v, shl(shr(0xFFFF_FFFF_FFFF_FFFF, bit), bit))
		}
	}
	parseImmTypeI := func(instr uint64) uint64 {
		return signExtend(shr(instr, 20), 11)
	}
	parseImmTypeS := func(instr uint64) uint64 {
		return signExtend(or(shl(shr(instr, 25), 5), and(shr(instr, 7), 0x1F)), 11)
	}
	parseImmTypeB := func(instr uint64) uint64 {
		return signExtend(
			or(
				or(
					shl(and(shr(instr, 8), 0xF), 1),
					shl(and(shr(instr, 25), 0x3F), 5),
				),
				or(
					shl(and(shr(instr, 7), 1), 11),
					shl(shr(instr, 31), 12),
				),
			),
			12,
		)
	}
	parseImmTypeU := func(instr uint64) uint64 {
		return signExtend(shl(shr(instr, 20), 12), 19)
	}
	parseImmTypeJ := func(instr uint64) uint64 {
		return signExtend(
			or(
				or(
					shl(and(shr(instr, 21), 0x1FF), 1),
					shl(and(shr(instr, 20), 1), 10),
				),
				or(
					shl(and(shr(instr, 12), 0xFF), 11),
					shl(shr(instr, 31), 19),
				),
			),
			19,
		)
	}

	if s.StateStackGindex < s.StateGindex {
		if s.StateStackGindex == 1 {
			s.StateValue = s.StateRoot
		}
		s.StateStackGindex <<= 1
		a, b := so.Get(s.StateValue)
		if (s.StateGindex>>s.StateStackDepth)&1 == 1 {
			s.StateStackGindex |= 1
			s.StateValue = b
			// keep track of where we have been, to use the trail to go back up the stack when writing
			s.StateStackHash = so.Remember(s.StateStackHash, a)
		} else {
			s.StateValue = a
			// keep track of where we have been, to use the trail to go back up the stack when writing
			s.StateStackHash = so.Remember(s.StateStackHash, b)
		}
		s.StateStackDepth += 1
	} else if s.StateStackGindex > s.StateGindex {
		// go back up the tree, to update the state root
		prevStackHash, prevSibling := so.Get(s.StateStackHash)
		s.StateStackHash = prevStackHash
		if s.StateStackGindex&1 == 1 {
			s.StateValue = so.Remember(prevSibling, s.StateValue)
		} else {
			s.StateValue = so.Remember(s.StateValue, prevSibling)
		}
		s.StateStackGindex >>= 1
		s.StateStackDepth -= 1
		if s.StateStackGindex == 1 {
			s.StateRoot = s.StateValue
		}
	} else {
		var memIndex uint64

		var dest64 *uint64
		var size uint64
		var signed bool
		var gindex1 uint64
		var gindex2 uint64
		var offset uint64
		var value [8]byte

		// run the next step
		switch s.SubIndex {
		case StepLoadPC:
			dest64 = &s.PC
			gindex1 = pcGindex
			goto loadOrWrite
		case StepLoadInstr:
			dest64 = &s.Instr
			gindex1 = makeMemGindex(s.PC)
			size = 4
			goto loadOrWrite
		case StepParseInstr:
			// these fields are ignored if not applicable to the instruction type / opcode
			s.Rd = and(shr(s.Instr, 7), 0b1_1111)
			s.Funct3 = and(shr(s.Instr, 12), 0b111)
			s.Rs1 = and(shr(s.Instr, 15), 0b1_1111)
			s.Rs2 = and(shr(s.Instr, 20), 0b1_1111)
			s.SubIndex += 1
			goto end
		case StepLoadRs1:
			dest64 = &s.Rs1Value
			gindex1 = makeRegisterGindex(s.Rs1)
			goto loadOrWrite
		case StepLoadRs2:
			dest64 = &s.Rs1Value
			gindex1 = makeRegisterGindex(s.Rs2)
			goto loadOrWrite
		case StepOpcode:
			switch s.Opcode {
			case 0b0000011: // memory loading
				// LB, LH, LW, LD, LBU, LHU, LWU
				imm := parseImmTypeI(s.Instr)
				signed = iszero(and(uint64(s.Funct3), 0b100))
				size = shl(uint64(1), and(uint64(s.Funct3), 0b11))
				dest64 = &s.RdValue
				memIndex = add(s.Rs1Value, signExtend(imm, 11))
				goto memLoadOrWrite
			case 0b0100011: // memory storing
				// SB, SH, SW, SD
				imm := parseImmTypeS(s.Instr)
				size = shl(uint64(1), uint64(s.Funct3))
				binary.BigEndian.PutUint64(value[:], s.Rs2Value)
				memIndex = add(s.Rs1Value, signExtend(imm, 11))
				goto memLoadOrWrite
			case 0b1100011: // branching
				branchHit := uint64(0)
				switch s.Funct3 {
				case 000: // BEQ
					branchHit = eq(s.Rs1Value, s.Rs2Value)
				case 001: // BNE
					branchHit = not(eq(s.Rs1Value, s.Rs2Value))
				case 100: // BLT
					branchHit = slt(s.Rs1Value, s.Rs2Value)
				case 101: // BGE
					branchHit = not(slt(s.Rs1Value, s.Rs2Value))
				case 110: // BLTU
					branchHit = lt(s.Rs1Value, s.Rs2Value)
				case 111: // BGEU
					branchHit = not(lt(s.Rs1Value, s.Rs2Value))
				}
				if iszero(branchHit) {
					s.PC += 4
				} else {
					imm := parseImmTypeB(s.Instr)
					// imm12 is a signed offset, in multiples of 2 bytes
					s.PC = add(s.PC, shl(signExtend(imm, 11), 1))
				}
				// not like the other opcodes: nothing to write to rd register, and PC has already changed
				s.SubIndex = StepWritePC
				goto end
			case 0b0010011: // immediate arithmetic and logic
				imm := parseImmTypeI(s.Instr)
				switch s.Funct3 {
				case 0b000: // ADDI
					s.RdValue = add(s.Rs1Value, signExtend(imm, 11))
				case 0b001: // SLLI
					s.RdValue = shl(s.Rs1Value, and(imm, 0b11_111)) // lower 6 bits in 64 bit mode
				case 0b010: // SLTI
					s.RdValue = slt(s.Rs1Value, imm)
				case 0b011: // SLTIU
					s.RdValue = lt(s.Rs1Value, imm)
				case 0b100: // XORI
					s.RdValue = xor(s.Rs1Value, imm)
				case 0b101:
					switch s.Funct7 {
					case 0b0000000: // SRLI
						s.RdValue = shr(s.Rs1Value, and(imm, 0b11_111)) // lower 6 bits in 64 bit mode
					case 0b0100000: // SRAI
						s.RdValue = sar(s.Rs1Value, and(imm, 0b11_111)) // lower 6 bits in 64 bit mode
					}
				case 0b110: // ORI
					s.RdValue = or(s.Rs1Value, imm)
				case 0b111: // ANDI
					s.RdValue = and(s.RdValue, imm)
				}
			case 0b0011011: // immediate arithmetic and logic signed 32 bit
				imm := parseImmTypeI(s.Instr)
				switch s.Funct3 {
				case 0b000: // ADDIW
					s.RdValue = add(s.Rs1Value, imm)
				case 0b001: // SLLIW
					s.RdValue = shl(s.Rs1Value, and(imm, 0b1_1111))
				case 0b101:
					shamt := and(imm, 0b1_1111)
					switch s.Funct7 {
					case 0b0000000: // SRLIW
						s.RdValue = signExtend(shr(and(s.Rs1Value, 0xFFFF_FFFF), shamt), sub(31, shamt))
					case 0b0100000: // SRAIW
						s.RdValue = signExtend(sar(and(s.Rs1Value, 0xFFFF_FFFF), shamt), sub(31, shamt))
					}
				}
			case 0b0110011: // register arithmetic and logic
				if s.Funct7 == 0b0000001 {
					switch s.Funct3 {
					case 0b000: // MUL: signed x signed
						s.RdValue = uint64(int64(s.Rs1Value) * int64(s.Rs2Value))
					case 0b001: // MULH: upper bits of signed x signed
						s.RdValue = uint64(new(big.Int).Rsh(new(big.Int).Mul(big.NewInt(int64(s.Rs1Value)), big.NewInt(int64(s.Rs2Value))), 64).Int64())
					case 0b010: // MULHSU: upper bits of signed x unsigned
						s.RdValue = uint64(new(big.Int).Rsh(new(big.Int).Mul(big.NewInt(int64(s.Rs1Value)), new(big.Int).SetUint64(s.Rs2Value)), 64).Int64())
					case 0b011: // MULHU: upper bits of unsigned x unsigned
						s.RdValue = new(big.Int).Rsh(new(big.Int).Mul(new(big.Int).SetUint64(s.Rs2Value), new(big.Int).SetUint64(s.Rs2Value)), 64).Uint64()
					case 0b100: // DIV
						if iszero(s.Rs2Value) {
							s.RdValue = ^uint64(0)
						} else {
							s.RdValue = sdiv(s.Rs1Value, s.Rs2Value)
						}
					case 0b101: // DIVU
						if iszero(s.Rs2Value) {
							s.RdValue = ^uint64(0)
						} else {
							s.RdValue = div(s.Rs1Value, s.Rs2Value)
						}
					case 0b110: // REM
						if iszero(s.Rs2Value) {
							s.RdValue = s.Rs1Value
						} else {
							s.RdValue = smod(s.Rs1Value, s.Rs2Value)
						}
					case 0b111: // REMU
						if iszero(s.Rs2Value) {
							s.RdValue = s.Rs1Value
						} else {
							s.RdValue = mod(s.Rs1Value, s.Rs2Value)
						}
					}
				} else {
					switch s.Funct3 {
					case 0b000:
						switch s.Funct7 {
						case 0b0000000: // ADD
							s.RdValue = add(s.Rs1Value, s.Rs2Value)
						case 0b0100000: // SUB
							s.RdValue = sub(s.Rs1Value, s.Rs1Value)
						}
					case 001: // SLL
						s.RdValue = shl(s.Rs1Value, and(s.Rs2Value, 0b11_1111))
					case 010: // SLT
						s.RdValue = slt(s.Rs1Value, s.Rs2Value)
					case 011: // SLTU
						s.RdValue = lt(s.Rs1Value, s.Rs2Value)
					case 100: // XOR
						s.RdValue = xor(s.Rs1Value, s.RdValue)
					case 101:
						switch s.Funct7 {
						case 0b0000000: // SRL
							s.RdValue = shr(s.Rs1Value, and(s.Rs2Value, 0b11_1111)) // logical: fill with zeroes
						case 0b0100000: // SRA
							s.RdValue = sar(s.Rs1Value, and(s.Rs2Value, 0b11_1111)) // arithmetic: sign bit is extended
						}
					case 110: // OR
						s.RdValue = or(s.Rs1Value, s.Rs2Value)
					case 111: // AND
						s.RdValue = and(s.Rs1Value, s.Rs2Value)
					}
				}
			case 0b0111011: // register arithmetic and logic in 32 bits
				if s.Funct7 == 0b0000001 { // RV64M extension
					switch s.Funct3 {
					case 000: // MULW
						s.RdValue = signExtend(and(mul(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF), 31)
					case 100: // DIVW
						if uint32(s.Rs2Value) == 0 {
							s.RdValue = ^uint64(0)
						} else {
							s.RdValue = signExtend(and(sdiv(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF), 31)
						}
					case 101: // DIVUW
						if uint32(s.Rs2Value) == 0 {
							s.RdValue = ^uint64(0)
						} else {
							s.RdValue = and(div(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF)
						}
					case 110: // REMW
						if uint32(s.Rs2Value) == 0 {
							s.RdValue = and(s.Rs1Value, 0xFFFF_FFFF)
						} else {
							s.RdValue = and(smod(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF)
						}
					case 111: // REMUW
						if uint32(s.Rs2Value) == 0 {
							s.RdValue = and(s.Rs1Value, 0xFFFF_FFFF)
						} else {
							s.RdValue = and(mod(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF)
						}
					}
				} else { // RV32M extension
					switch s.Funct3 {
					case 0b000:
						switch s.Funct7 {
						case 0b0000000: // ADDW
							s.RdValue = signExtend(and(add(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF), 31)
						case 0b0100000: // SUBW
							s.RdValue = signExtend(and(sub(and(s.Rs1Value, 0xFFFF_FFFF), and(s.Rs2Value, 0xFFFF_FFFF)), 0xFFFF_FFFF), 31)
						}
					case 0b001: // SLLW
						s.RdValue = signExtend(and(shl(s.Rs1Value, and(s.Rs2Value, 0b1_1111)), 0xFFFF_FFFF), 31)
					case 0b101:
						shamt := and(s.Rs2Value, 0b1_1111)
						switch s.Funct7 {
						case 0b0000000: // SRLW
							s.RdValue = signExtend(shr(and(s.Rs1Value, 0xFFFF_FFFF), shamt), sub(31, shamt))
						case 0b0100000: // SRAW
							s.RdValue = signExtend(sar(and(s.Rs1Value, 0xFFFF_FFFF), shamt), sub(31, shamt))
						}
					}
				}
			case 0b0110111: // LUI = Load upper immediate
				imm := parseImmTypeU(s.Instr)
				s.RdValue = shl(imm, 12)
			case 0b0010111: // AUIPC = Add upper immediate to PC
				imm := parseImmTypeU(s.Instr)
				s.RdValue = add(s.PC, shl(imm, 12))
			case 0b1101111: // JAL = Jump and link
				imm := parseImmTypeJ(s.Instr)
				s.RdValue = add(s.PC, 4)
				s.PC = add(s.PC, signExtend(shl(imm, 1), 21)) // signed offset in multiples of 2 bytes
				s.SubIndex = StepWriteRd
				goto end
			case 0b1100111: // JALR = Jump and link register
				imm := parseImmTypeI(s.Instr)
				s.RdValue = add(s.PC, 4)
				s.PC = and(add(s.Rs1Value, signExtend(imm, 12)), 0xFFFF_FFFF_FFFF_FFFF) // least significant bit is set to 0
				s.SubIndex = StepWriteRd
				goto end
			case 0b1110011:
				switch s.Funct3 {
				case 0b000:
					// TODO: I type instruction
					//000000000000 1110011 ECALL
					//000000000001 1110011 EBREAK
				case 0b001: // CSRRW
				case 0b010: // CSRRS  a.k.a. SYSTEM instruction
					// TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
				case 0b011: // CSRRC
				case 0b101: // CSRRWI
				case 0b110: // CSRRSI
				case 0b111: // CSRRCI
				}
			case 0b0101111: // RV32A and RV32A atomic operations extension
				// 0b010 == RV32A W variants
				// 0b011 == RV64A D variants
				//size := 1 << s.Funct3
				switch (s.Funct7 & 0b11111) >> 2 {
				case 0b00010: // LR
				case 0b00011: // SC
				case 0b00001: // AMOSWAP
				case 0b00000: // AMOADD
				case 0b00100: // AMOXOR
				case 0b01100: // AMOAND
				case 0b01000: // AMOOR
				case 0b10000: // AMOMIN
				case 0b10100: // AMOMAX
				case 0b11000: // AMOMINU
				case 0b11100: // AMOMAXU
				}
			case 0b0001111:
				// TODO: different layout of func data
				// "fm pred succ"
				switch s.Funct3 {
				case 0b000:
					switch s.Funct7 {
					case 0b1000001: // FENCE.TSO
					default: // FENCE
					}
				case 0b001: // FENCE.I
				}
			}
			s.PC = add(s.PC, 4)
			s.SubIndex = StepWriteRd
			goto end
		case StepWriteRd:
			gindex1 = makeRegisterGindex(s.Rd)
			binary.BigEndian.PutUint64(value[:], s.RdValue)
			goto loadOrWrite
		case StepWritePC:
			gindex1 = pcGindex
			binary.BigEndian.PutUint64(value[:], s.PC)
			goto loadOrWrite
		case StepFinal:
			stateRoot := s.StateRoot
			// zero out everything in preparation of next instruction
			return VMScratchpad{StateRoot: stateRoot}
		}

	memLoadOrWrite:
		gindex1 = makeMemGindex(memIndex)
		offset = and(memIndex, 31)
		gindex2 = uint64(0)
		if iszero(lt(add(offset, size), 32)) { // if offset+size >= 32, then it spans into the next memory chunk
			// note: intentional overflow, circular 64 bit memory is part of riscv5 spec (chapter 1.4)
			gindex2 = makeMemGindex(add(memIndex, sub(size, 1)))
		}
		goto loadOrWrite
	loadOrWrite:
		switch s.InstrPhaseIndex {
		case StateStartFirst:
			s.StateStackGindex = 1
			s.StateGindex = gindex1
			s.InstrPhaseIndex = StateEndFirst
		case StateEndFirst:
			if dest64 == nil { // writing
				if gindex1 != (makeRegisterGindex(0)) { // can't write to register ZERO
					// note: StateValue holds the old 32 bytes, some of which may stay the same
					copy(s.StateValue[offset:], value[:])
					s.StateGindex = 1
				}
			} else { // reading
				copy(s.Scratch[:], s.StateValue[offset:])

				if signed {
					binary.BigEndian.PutUint64(s.Scratch[:8], signExtend(binary.BigEndian.Uint64(s.Scratch[:8]), sub(shl(size, 3), 1)))
				}

				// if unaligned, these values written to dest may not be the final values
				*dest64 = binary.BigEndian.Uint64(s.Scratch[:8])
			}
			if gindex2 != 0 {
				s.InstrPhaseIndex = StateStartSecond
			} else {
				s.SubIndex += 1
				s.InstrPhaseIndex = 0
			}
		case StateStartSecond:
			s.StateStackGindex = 1
			s.StateGindex = gindex2
			s.InstrPhaseIndex = StateEndSecond
		case StateEndSecond:
			firstChunkBytes := 32 - offset
			if firstChunkBytes > size {
				firstChunkBytes = size
			}
			secondChunkBytes := size - firstChunkBytes
			if dest64 == nil { // writing
				// note: StateValue holds the old 32 bytes, some of which may stay the same
				copy(s.StateValue[:secondChunkBytes], value[firstChunkBytes:size])
				s.StateGindex = 1
			} else { // reading
				copy(s.Scratch[firstChunkBytes:], s.StateValue[:secondChunkBytes])

				if signed {
					binary.BigEndian.PutUint64(s.Scratch[:8], signExtend(binary.BigEndian.Uint64(s.Scratch[:8]), sub(shl(size, 3), 1)))
				}

				*dest64 = binary.BigEndian.Uint64(s.Scratch[:8])
			}
			s.SubIndex += 1
			s.InstrPhaseIndex = 0
		}
		goto end
	}
end:
	return s
}
