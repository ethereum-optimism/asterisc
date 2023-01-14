package slow

import (
	"encoding/binary"
	"fmt"

	"github.com/holiman/uint256"

	"github.com/protolambda/asterisc/rvgo/oracle"
)

type VMScratchpad struct {
	SubIndex uint64 // step in the instruction execution
	PC       U64    // PC counter

	Instr U64 // raw instruction

	Opcode U64 // parsed instruction
	Funct3 U64 //
	Funct7 U64 //

	Rd  U64 // destination register index
	Rs1 U64 // source register 1 index
	Rs2 U64 // source register 2 index

	RdValue  U64 // destination register value to write
	Rs1Value U64 // loaded source registers. Only load if rs1/rs2 are not zero.
	Rs2Value U64 //

	StateRoot [32]byte // commits to state as structured binary tree

	// State machine
	StateGindex      U256     // target
	StateStackHash   [32]byte // hash of previous traversed stack and last stack element
	StateStackGindex U256     // to navigate the state loading/writing
	StateStackDepth  uint8
	StateValue       [32]byte

	Scratch [8]byte

	Gindex1 U256  // first leaf to read from or write to
	Gindex2 U256  // second leaf to read from or write to
	Offset  uint8 // offset: value might start anywhere in Gindex1 leaf, and overflow to Gindex2 leaf
	Value   U64   // value to write
	Signed  bool  // if value to read should be sign-extended
	Size    uint8 // size of value to read, may be 1, 2, 4 or 8 bytes
	Dest    U64   // destination to load a value back into
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

// tree:
// ```
//
//	         1
//	    2          3
//	 4    5     6     7
//	8 9 10 11 12 13 14 15
//
// ```
var (
	pcGindex        = toU256(8)
	memoryGindex    = toU256(9)
	registersGindex = toU256(10)
	csrGindex       = toU256(11)
	exitGindex      = toU256(12)
)

func makeMemGindex(byteIndex U64) U256 {
	// memory is packed in 32 byte leaf values
	return or(shl(memoryGindex, toU256(64-5)), shr(U256(byteIndex), toU256(5)))
}

func makeRegisterGindex(register U64) U256 {
	if x := U256(register); x.Uint64() >= 32 {
		panic("there are only 32 valid registers")
	}
	return or(shl(registersGindex, toU256(5)), U256(register))
}

func memToStateOp(memIndex U64, size U64) (offset uint8, gindex1, gindex2 U256) {
	gindex1 = makeMemGindex(memIndex)
	offset = uint8(and64(memIndex, toU64(31)).val())
	gindex2 = U256{}
	if iszero(lt(add(toU256(offset), U256(size)), toU256(32))) { // if offset+size >= 32, then it spans into the next memory chunk
		// note: intentional overflow, circular 64 bit memory is part of riscv5 spec (chapter 1.4)
		gindex2 = makeMemGindex(add64(memIndex, sub64(size, toU64(1))))
	}
	return
}

var (
	destNone     = toU64(0)
	destPc       = toU64(1)
	destInstr    = toU64(2)
	destRs1Value = toU64(3)
	destRs2Value = toU64(4)
	destRdvalue  = toU64(5)
)

func encodePacked(v U64) (out [8]byte) {
	binary.BigEndian.PutUint64(out[:], v.val())
	return
}

func decodeU64(v []byte) (out U64) {
	if len(v) != 8 {
		panic("bad u64 decode")
	}
	(*U256)(&out).SetBytes8(v[:])
	return
}

func SubStep(s VMScratchpad, so oracle.VMStateOracle) VMScratchpad {
	// this first part with state stack machine can be written in solidity - it's heavy on memory/calldata interactions.
	if s.StateStackGindex.Lt(&s.StateGindex) {
		// READING MODE: if the stack gindex is lower than target, then traverse to target
		if s.StateStackGindex.Eq(uint256.NewInt(1)) {
			s.StateValue = s.StateRoot
		}
		s.StateStackGindex = shl(s.StateStackGindex, toU256(1))
		a, b := so.Get(s.StateValue)
		if eq(and(shr(s.StateGindex, toU256(s.StateStackDepth)), toU256(1)), toU256(1)) != (U256{}) {
			s.StateStackGindex = or(s.StateStackGindex, toU256(1))
			s.StateValue = b
			// keep track of where we have been, to use the trail to go back up the stack when writing
			s.StateStackHash = so.Remember(s.StateStackHash, a)
		} else {
			s.StateValue = a
			// keep track of where we have been, to use the trail to go back up the stack when writing
			s.StateStackHash = so.Remember(s.StateStackHash, b)
		}
		s.StateStackDepth += 1
		return s
	} else if s.StateStackGindex.Gt(&s.StateGindex) {
		// WRITING MODE: if the stack gindex is higher than the target, then traverse back to root and update along the way
		prevStackHash, prevSibling := so.Get(s.StateStackHash)
		s.StateStackHash = prevStackHash
		if eq(and(s.StateStackGindex, toU256(1)), toU256(1)) != (U256{}) {
			s.StateValue = so.Remember(prevSibling, s.StateValue)
		} else {
			s.StateValue = so.Remember(s.StateValue, prevSibling)
		}
		s.StateStackGindex = shr(s.StateStackGindex, toU256(1))
		s.StateStackDepth -= 1
		if s.StateStackGindex == toU256(1) {
			s.StateRoot = s.StateValue
		}
		return s
	}

	// if we want to read/write a value at the given gindex, then try.
	if s.Gindex1 != (U256{}) {
		if s.Gindex1 != s.StateGindex {
			// if we have not reached the gindex yet, then we need to start traversal to it
			s.StateStackGindex = toU256(1)
			s.StateGindex = s.Gindex1
		} else {
			s.Gindex1 = U256{}

			// we reached the value, now load/write it
			if s.Dest != destNone { // writing
				// note: StateValue holds the old 32 bytes, some of which may stay the same
				v := encodePacked(s.Value)
				copy(s.StateValue[s.Offset:], v[:s.Size])
				s.StateGindex = toU256(1)
			} else { // reading
				copy(s.Scratch[:], s.StateValue[s.Offset:])
			}
		}
		return s
	}

	if s.Gindex2 != (U256{}) {
		if s.Gindex2 != s.StateGindex {
			// if we have not reached the gindex yet, then we need to start traversal to it
			s.StateStackGindex = toU256(1)
			s.StateGindex = s.Gindex2
		} else {
			s.Gindex2 = U256{}

			firstChunkBytes := 32 - s.Offset
			if firstChunkBytes > s.Size {
				firstChunkBytes = s.Size
			}
			secondChunkBytes := s.Size - firstChunkBytes

			// we reached the value, now load/write it
			if s.Dest != destNone { // writing
				// note: StateValue holds the old 32 bytes, some of which may stay the same
				v := encodePacked(s.Value)
				copy(s.StateValue[:secondChunkBytes], v[firstChunkBytes:s.Size])
				s.StateGindex = toU256(1)
			} else { // reading
				copy(s.Scratch[firstChunkBytes:], s.StateValue[:secondChunkBytes])
			}
		}
		return s
	}

	if s.Dest != destNone { // complete reading work if any
		val := decodeU64(s.Scratch[:s.Size])

		if s.Signed {
			topBitIndex := (s.Size << 3) - 1
			val = signExtend64(decodeU64(s.Scratch[:8]), toU64(topBitIndex))
		}

		// if unaligned, these values written to dest may not be the final values
		switch s.Dest {
		case destPc:
			s.Instr = val
		case destInstr:
			s.Instr = val
		case destRs1Value:
			s.Rs1Value = val
		case destRs2Value:
			s.Rs1Value = val
		case destRdvalue:
			s.RdValue = val
		}
		s.Dest = destNone
		return s
	}

	// unpacked sub-step state
	rdValue := s.RdValue
	pc := s.PC
	instr := s.Instr
	opcode := s.Opcode
	rd := s.Rd
	funct3 := s.Funct3
	rs1 := s.Rs1
	rs2 := s.Rs2
	funct7 := s.Funct7
	rs1Value := s.Rs1Value
	rs2Value := s.Rs2Value

	// write-only sub-state. All reading/processing happens in state machinery above
	var gindex1 U256
	var gindex2 U256
	var size U64
	var signed bool
	var value U64
	var dest U64
	var offset uint8

	// run the next step
	switch s.SubIndex {
	case StepLoadPC:
		dest = destPc
		size = toU64(4)
		signed = false
		gindex1 = pcGindex
		s.SubIndex += 1
	case StepLoadInstr:
		dest = destInstr
		gindex1 = makeMemGindex(s.PC)
		size = toU64(4)
		s.SubIndex += 1
	case StepParseInstr:
		// these fields are ignored if not applicable to the instruction type / opcode
		opcode = parseOpcode(instr)
		rd = parseRd(instr) // destination register index
		funct3 = parseFunct3(instr)
		rs1 = parseRs1(instr) // source register 1 index
		rs2 = parseRs2(instr) // source register 2 index
		funct7 = parseFunct7(instr)
		s.SubIndex += 1
	case StepLoadRs1:
		dest = destRs1Value
		gindex1 = makeRegisterGindex(s.Rs1)
		s.SubIndex += 1
	case StepLoadRs2:
		dest = destRs2Value
		gindex1 = makeRegisterGindex(s.Rs2)
		s.SubIndex += 1
	case StepOpcode:
		switch s.Opcode.val() {
		case 0b0000011: // memory loading
			// LB, LH, LW, LD, LBU, LHU, LWU
			imm := parseImmTypeI(instr)
			signed = iszero64(and64(funct3, toU64(8)))      // 8 = 100 -> bitflag
			size = shl64(toU64(1), and64(funct3, toU64(3))) // 3 = 11 -> 1, 2, 4, 8 bytes size
			memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
			dest = destRdvalue
			offset, gindex1, gindex2 = memToStateOp(memIndex, size)
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b0100011: // memory storing
			// SB, SH, SW, SD
			imm := parseImmTypeS(instr)
			size = shl64(toU64(1), funct3)
			value = rs2Value
			memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
			offset, gindex1, gindex2 = memToStateOp(memIndex, size)
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWritePC
		case 0b1100011: // branching
			branchHit := toU64(0)
			switch funct3.val() {
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
			switch branchHit.val() {
			case 0:
				pc = add64(pc, toU64(4))
			default:
				imm := parseImmTypeB(instr)
				// imm12 is a signed offset, in multiples of 2 bytes
				pc = add64(pc, shl64(signExtend64(imm, toU64(11)), toU64(1)))
			}
			// not like the other opcodes: nothing to write to rd register, and PC has already changed
			s.SubIndex = StepWritePC
		case 0b0010011: // immediate arithmetic and logic
			imm := parseImmTypeI(instr)
			switch funct3.val() {
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
				switch funct7.val() {
				case 0x00: // 0000000 = SRLI
					rdValue = shr64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
				case 0x20: // 0100000 == SRAI
					rdValue = sar64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
				}
			case 6: // 110 = ORI
				rdValue = or64(rs1Value, imm)
			case 7: // 111 = ANDI
				rdValue = and64(rdValue, imm)
			}
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b0011011: // immediate arithmetic and logic signed 32 bit
			imm := parseImmTypeI(instr)
			switch funct3.val() {
			case 0: // 000 = ADDIW
				rdValue = add64(rs1Value, imm)
			case 1: // 001 = SLLIW
				rdValue = shl64(rs1Value, and64(imm, toU64(0x1F)))
			case 5: // 101 = SR~
				shamt := and64(imm, toU64(0x1F))
				switch funct7.val() {
				case 0x00: // 0000000 = SRLIW
					rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
				case 0x20: // 0100000 = SRAIW
					rdValue = signExtend64(sar64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
				}
			}
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b0110011: // register arithmetic and logic
			switch funct7.val() {
			case 1: // RV32M extension
				switch funct3.val() {
				case 0: // 000 = MUL: signed x signed
					rdValue = mul64(rs1Value, rs2Value)
				case 1: // 001 = MULH: upper bits of signed x signed
					rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value)), toU256(64)))
				case 2: // 010 = MULHSU: upper bits of signed x unsigned
					rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
				case 3: // 011 = MULHU: upper bits of unsigned x unsigned
					rdValue = u256ToU64(shr(mul(u64ToU256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
				case 4: // 100 = DIV
					switch rs2Value.val() {
					case 0:
						rdValue = u64Mask()
					default:
						rdValue = sdiv64(rs1Value, rs2Value)
					}
				case 5: // 101 = DIVU
					switch rs2Value.val() {
					case 0:
						rdValue = u64Mask()
					default:
						rdValue = div64(rs1Value, rs2Value)
					}
				case 6: // 110 = REM
					switch rs2Value.val() {
					case 0:
						rdValue = rs1Value
					default:
						rdValue = smod64(rs1Value, rs2Value)
					}
				case 7: // 111 = REMU
					switch rs2Value.val() {
					case 0:
						rdValue = rs1Value
					default:
						rdValue = mod64(rs1Value, rs2Value)
					}
				}
			default:
				switch funct3.val() {
				case 0: // 000 = ADD/SUB
					switch funct7.val() {
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
					switch funct7.val() {
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
			s.SubIndex = StepWriteRd
		case 0b0111011: // register arithmetic and logic in 32 bits
			switch funct7.val() {
			case 1: // RV64M extension
				switch funct3.val() {
				case 0: // 000 = MULW
					rdValue = signExtend64(and64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
				case 4: // 100 = DIVW
					switch rs2Value.val() {
					case 0:
						rdValue = u64Mask()
					default:
						rdValue = signExtend64(and64(sdiv64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
					}
				case 5: // 101 = DIVUW
					switch rs2Value.val() {
					case 0:
						rdValue = u64Mask()
					default:
						rdValue = and64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
					}
				case 6: // 110 = REMW
					switch rs2Value.val() {
					case 0:
						rdValue = and64(rs1Value, u32Mask())
					default:
						rdValue = and64(smod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
					}
				case 7: // 111 = REMUW
					switch rs2Value.val() {
					case 0:
						rdValue = and64(rs1Value, u32Mask())
					default:
						rdValue = and64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask())
					}
				}
			default: // RV32M extension
				switch funct3.val() {
				case 0: // 000 = ADDW/SUBW
					switch funct7.val() {
					case 0x00: // 0000000 = ADDW
						rdValue = signExtend64(and64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
					case 0x20: // 0100000 = SUBW
						rdValue = signExtend64(and64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())), u32Mask()), toU64(31))
					}
				case 1: // 001 = SLLW
					rdValue = signExtend64(and64(shl64(rs1Value, and64(rs2Value, toU64(0x1F))), u32Mask()), toU64(31))
				case 5: // 101 = SR~
					shamt := and64(rs2Value, toU64(0x1F))
					switch funct7.val() {
					case 0x00: // 0000000 = SRLW
						rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
					case 0x20: // 0100000 = SRAW
						rdValue = signExtend64(sar64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
					}
				}
			}
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b0110111: // LUI = Load upper immediate
			imm := parseImmTypeU(instr)
			rdValue = shl64(imm, toU64(12))
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b0010111: // AUIPC = Add upper immediate to PC
			imm := parseImmTypeU(instr)
			rdValue = add64(pc, shl64(imm, toU64(12)))
			pc = add64(pc, toU64(4))
			s.SubIndex = StepWriteRd
		case 0b1101111: // JAL = Jump and link
			imm := parseImmTypeJ(instr)
			rdValue = add64(pc, toU64(4))
			pc = add64(pc, signExtend64(shl64(imm, toU64(1)), toU64(21))) // signed offset in multiples of 2 bytes
			s.SubIndex = StepWriteRd
		case 0b1100111: // JALR = Jump and link register
			imm := parseImmTypeI(instr)
			rdValue = add64(pc, toU64(4))
			pc = and64(add64(rs1Value, signExtend64(imm, toU64(12))), xor64(u64Mask(), toU64(1))) // least significant bit is set to 0
			s.SubIndex = StepWriteRd
		case 0b1110011:
			switch funct3.val() {
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
			s.SubIndex = StepWriteRd
		case 0b0101111: // RV32A and RV32A atomic operations extension
			// TODO atomic operations
			// 0b010 == RV32A W variants
			// 0b011 == RV64A D variants
			//size := 1 << funct3
			switch shr64(and64(funct7, toU64(0x1F)), toU64(2)).val() {
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
			s.SubIndex = StepWriteRd
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
			s.SubIndex = StepWriteRd
		default:
			panic(fmt.Errorf("unknown opcode: %b full instruction: %b", opcode, instr))
		}
	case StepWriteRd:
		switch rd.val() {
		case 0:
			// never write to register 0, it must stay zero
			s.SubIndex = StepWritePC
		default:
			gindex1 = makeRegisterGindex(rd)
			value = rdValue
		}
	case StepWritePC:
		gindex1 = pcGindex
		value = pc
	case StepFinal:
		stateRoot := s.StateRoot
		// zero out everything in preparation of next instruction
		return VMScratchpad{StateRoot: stateRoot}
	}

	// encode sub-step state
	s.PC = pc
	s.Instr = instr
	s.Opcode = opcode
	s.Rd = rd
	s.Funct3 = funct3
	s.Rs1 = rs1
	s.Rs2 = rs2
	s.Funct7 = funct7
	s.Rs1Value = rs1Value
	s.Rs2Value = rs2Value
	s.Gindex1 = gindex1
	s.Gindex2 = gindex2
	s.Offset = offset
	s.Value = value
	s.Signed = signed
	s.Size = uint8(size.val())
	s.Dest = dest

	return s
}
