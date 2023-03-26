package fast

import (
	"fmt"
	"io"
)

func revertWithCode(code uint64, err error) error {
	return fmt.Errorf("code %x: %w", code, err)
}

// Step runs a single instruction
// Note: errors are only returned in debugging/tooling modes, not in production use.
func Step(s *VMState, stdOut, stdErr io.Writer) error {
	if s.Exited {
		return nil
	}

	sysCall := func() error {
		a7 := s.loadRegister(toU64(17))

		switch a7 {
		case 93: // exit
			a0 := s.loadRegister(toU64(10))
			s.Exit = a0
			s.Exited = true
			// program stops here, no need to change registers.
		case 214: // brk
			// Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

			// brk(0) changes nothing about the memory, and returns the current page break
			v := shl64(toU64(30), toU64(1)) // set program break at 1 GiB
			s.writeRegister(toU64(10), v)
			s.writeRegister(toU64(11), toU64(0)) // no error
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
				// No hint, allocate it ourselves, by as much as the requested length.
				// Increase the length to align it with desired page size if necessary.
				align := and64(length, 4095)
				if !iszero64(align) {
					length = add64(length, sub64(4096, align))
					length += align
				}
				s.writeRegister(toU64(10), s.Heap)
				s.Heap += length // increment heap with length
				fmt.Printf("mmap: 0x%016x (+ 0x%x increase)\n", s.Heap, length)
			default:
				// allow hinted memory address (leave it in A0 as return argument)
				fmt.Printf("mmap: 0x%016x (0x%x allowed)\n", addr, length)
			}
			s.writeRegister(toU64(11), toU64(0)) // no error
		case 63: // read
			fd := s.loadRegister(toU64(10))    // A0 = fd
			addr := s.loadRegister(toU64(11))  // A1 = *buf addr
			count := s.loadRegister(toU64(12)) // A2 = count
			var n, errCode U64
			switch fd {
			case 0: // stdin
				n = toU64(0) // never read anything from stdin
				errCode = toU64(0)
			case 3: // pre-image oracle
				x, err := s.readPreimageValue(addr, count)
				if err != nil {
					return fmt.Errorf("preimage reading error (PC: %d): %w", s.PC, err)
				}
				n = x
				errCode = toU64(0)
			default:
				n = u64Mask()         //  -1 (reading error)
				errCode = toU64(0x4d) // EBADF
			}
			s.writeRegister(toU64(10), n)
			s.writeRegister(toU64(11), errCode)
		case 64: // write
			fd := s.loadRegister(toU64(10))    // A0 = fd
			addr := s.loadRegister(toU64(11))  // A1 = *buf addr
			count := s.loadRegister(toU64(12)) // A2 = count
			var n, errCode U64
			switch fd {
			case 1: // stdout
				_, err := io.Copy(stdOut, s.GetMemRange(addr, count))
				if err != nil {
					return fmt.Errorf("stdout writing err: %w", err)
				}
				n = count // write completes fully in single instruction step
				errCode = toU64(0)
			case 2: // stderr
				_, err := io.Copy(stdErr, s.GetMemRange(addr, count))
				if err != nil {
					return fmt.Errorf("stderr writing err: %w", err)
				}
				n = count // write completes fully in single instruction step
				errCode = toU64(0)
			case 3: // pre-image oracle
				n = s.writePreimageKey(addr, count)
				errCode = toU64(0) // no error
			default: // any other file, including (4) pre-image hinter
				n = u64Mask()         //  -1 (writing error)
				errCode = toU64(0x4d) // EBADF
			}
			s.writeRegister(toU64(10), n)
			s.writeRegister(toU64(11), errCode)
		case 25: // fcntl - file descriptor manipulation / info lookup
			fd := s.loadRegister(toU64(10))  // A0 = fd
			cmd := s.loadRegister(toU64(11)) // A1 = cmd
			var out, errCode U64
			switch cmd {
			case 0x3: // F_GETFL: get file descriptor flags
				switch fd {
				case 0: // stdin
					out = 0 // O_RDONLY
				case 1: // stdout
					out = 1 // O_WRONLY
				case 2: // stderr
					out = 1 // O_WRONLY
				case 3: // pre-image oracle
					out = 2 // O_RDWR
				default:
					out = u64Mask()
					errCode = toU64(0x4d) // EBADF
				}
			default: // no other commands: don't allow changing flags, duplicating FDs, etc.
				out = u64Mask()
				errCode = toU64(0x16) // EINVAL (cmd not recognized by this kernel)
			}
			s.writeRegister(toU64(10), out)
			s.writeRegister(toU64(11), errCode) // EBADF
		case 56: // openat - the Go linux runtime will try to open optional /sys/kernel files for performance hints
			s.writeRegister(toU64(10), u64Mask())
			s.writeRegister(toU64(11), toU64(0xd)) // EACCES - no access allowed
		case 123: // sched_getaffinity - hardcode to indicate affinity with any cpu-set mask
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		case 113: // clock_gettime
			addr := s.loadRegister(toU64(11))        // addr of timespec struct
			s.storeMem(addr, 8, 0x1337)              // seconds
			s.storeMem(add64(addr, toU64(8)), 8, 42) // nanoseconds: must be nonzero to pass Go runtimeInitTime check
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		case 135: // rt_sigprocmask - ignore any sigset changes
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		case 132: // sigaltstack - ignore any hints of an alternative signal receiving stack addr
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		case 178: // gettid - hardcode to 0
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		case 134: // rt_sigaction - no-op, we never send signals, and thus need no sig handler info
			s.writeRegister(toU64(10), toU64(0))
			s.writeRegister(toU64(11), toU64(0))
		//case 220: // clone - not supported
		case 163: // getrlimit
			res := s.loadRegister(toU64(10))
			addr := s.loadRegister(toU64(11))
			switch res {
			case 0x7: // RLIMIT_NOFILE
				s.storeMem(addr, toU64(8), shortToU64(1024))           // soft limit. 1024 file handles max open
				s.storeMem(add64(addr, 8), toU64(8), shortToU64(1024)) // hard limit
			default:
				return revertWithCode(0xf0012, fmt.Errorf("unrecognized resource limit lookup: %d", res))
			}
		default: // every other syscall results in exit with error code
			return revertWithCode(0xf001ca11, fmt.Errorf("unrecognized system call: %d", a7))
		}
		return nil
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
		rs1Value := s.loadRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		rdValue := s.loadMem(memIndex, size, signed)
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x23: // 010_0011: memory storing
		// SB, SH, SW, SD
		imm := parseImmTypeS(instr)
		size := shl64(funct3, toU64(1))
		value := s.loadRegister(rs2)
		rs1Value := s.loadRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		s.storeMem(memIndex, size, value)
		s.setPC(add64(pc, toU64(4)))
	case 0x63: // 110_0011: branching
		rs1Value := s.loadRegister(rs1)
		rs2Value := s.loadRegister(rs2)
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
		rs1Value := s.loadRegister(rs1)
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
			switch shr64(toU64(6), imm) { // in rv64i the top 6 bits select the shift type
			case 0x00: // 000000 = SRLI
				rdValue = shr64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
			case 0x10: // 010000 = SRAI
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
		rs1Value := s.loadRegister(rs1)
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3 {
		case 0: // 000 = ADDIW
			rdValue = mask32Signed64(add64(rs1Value, imm))
		case 1: // 001 = SLLIW
			rdValue = mask32Signed64(shl64(and64(imm, toU64(0x1F)), rs1Value))
		case 5: // 101 = SR~
			shamt := and64(imm, toU64(0x1F))
			switch shr64(toU64(6), imm) { // in rv64i the top 6 bits select the shift type
			case 0x00: // 000000 = SRLIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
			case 0x10: // 010000 = SRAIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
			}
		}
		s.writeRegister(rd, rdValue)
		s.setPC(add64(pc, toU64(4)))
	case 0x33: // 011_0011: register arithmetic and logic
		rs1Value := s.loadRegister(rs1)
		rs2Value := s.loadRegister(rs2)
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
		rs1Value := s.loadRegister(rs1)
		rs2Value := s.loadRegister(rs2)
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
		s.setPC(add64(pc, signExtend64(shl64(toU64(1), imm), toU64(20)))) // signed offset in multiples of 2 bytes (last bit is there, but ignored)
	case 0x67: // 110_0111: JALR = Jump and link register
		rs1Value := s.loadRegister(rs1)
		imm := parseImmTypeI(instr)
		rdValue := add64(pc, toU64(4))
		s.writeRegister(rd, rdValue)
		s.setPC(and64(add64(rs1Value, signExtend64(imm, toU64(11))), xor64(u64Mask(), toU64(1)))) // least significant bit is set to 0
	case 0x73: // 111_0011: environment things
		switch funct3 {
		case 0: // 000 = ECALL/EBREAK
			switch shr64(toU64(20), instr) { // I-type, top 12 bits
			case 0: // imm12 = 000000000000 ECALL
				if err := sysCall(); err != nil {
					return err
				}
				if !s.Exited { // don't change any state if we have exited with the syscall already
					s.setPC(add64(pc, toU64(4)))
				}
			default: // imm12 = 000000000001 EBREAK
				// ignore breakpoint
				s.setPC(add64(pc, toU64(4)))
			}
		default: // CSR instructions
			imm := parseCSSR(instr)
			rdValue := s.readCSR(imm)
			value := rs1
			rs1Value := s.loadRegister(rs1)
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
		// acquire and release bits:
		//   aq := and64(shr64(toU64(1), funct7), toU64(1))
		//   rl := and64(funct7, toU64(1))
		// if none set: unordered
		// if aq is set: no following mem ops observed before acquire mem op
		// if rl is set: release mem op not observed before earlier mem ops
		// if both set: sequentially consistent
		// These are no-op here because there is no pipeline of mem ops to acquire/release.

		// 0b010 == RV32A W variants
		// 0b011 == RV64A D variants
		size := shl64(funct3, toU64(1))
		if lt64(size, toU64(4)) != 0 {
			return revertWithCode(0xbada70, fmt.Errorf("bad AMO size: %d", size))
		}
		addr := s.loadRegister(rs1)
		// TODO check if addr is aligned

		op := shr64(toU64(2), funct7)
		switch op {
		case 0x2: // 00010 = LR = Load Reserved
			v := s.loadMem(addr, size, true)
			s.writeRegister(rd, v)
			s.setLoadReservation(addr)
		case 0x3: // 00011 = SC = Store Conditional
			rdValue := toU64(1)
			if eq64(addr, s.getLoadReservation()) != 0 {
				rs2Value := s.loadRegister(rs2)
				s.storeMem(addr, size, rs2Value)
				rdValue = toU64(0)
			}
			s.writeRegister(rd, rdValue)
			s.setLoadReservation(0)
		default: // AMO: Atomic Memory Operation
			rs2Value := s.loadRegister(rs2)
			if eq64(size, toU64(4)) != 0 {
				rs2Value = mask32Signed64(rs2Value)
			}
			// Specifying the operation allows us to implement it closer to the memory for smaller witness data.
			// And that too can be optimized: only one 32 bytes leaf is affected,
			// since AMOs are always 4 or 8 byte aligned (Zam extension not supported here).
			var dest U64
			switch op {
			case 0x0: // 00000 = AMOADD = add
				dest = destADD
			case 0x1: // 00001 = AMOSWAP
				dest = destSWAP
			case 0x4: // 00100 = AMOXOR = xor
				dest = destXOR
			case 0x8: // 01000 = AMOOR = or
				dest = destOR
			case 0xc: // 01100 = AMOAND = and
				dest = destAND
			case 0x10: // 10000 = AMOMIN = min signed
				dest = destMIN
			case 0x14: // 10100 = AMOMAX = max signed
				dest = destMAX
			case 0x18: // 11000 = AMOMINU = min unsigned
				dest = destMINU
			case 0x1c: // 11100 = AMOMAXU = max unsigned
				dest = destMAXU
			default:
				return revertWithCode(0xf001a70, fmt.Errorf("unknown atomic operation %d", op))
			}
			rdValue := s.opMem(dest, addr, size, rs2Value)
			s.writeRegister(rd, rdValue)
		}
		s.setPC(add64(pc, toU64(4)))
	case 0x0F: // 000_1111: fence
		// Used to impose additional ordering constraints; flushing the mem operation pipeline.
		// This VM doesn't have a pipeline, nor additional harts, so this is a no-op.
		// FENCE / FENCE.TSO / FENCE.I all no-op: there's nothing to synchronize.
		s.setPC(add64(pc, toU64(4)))
	case 0x07: // FLW/FLD: floating point load word/double
		s.setPC(add64(pc, toU64(4))) // no-op this.
	case 0x27: // FSW/FSD: floating point store word/double
		s.setPC(add64(pc, toU64(4))) // no-op this.
	case 0x53: // FADD etc. no-op is enough to pass Go runtime check
		s.setPC(add64(pc, toU64(4))) // no-op this.
	default: // any other opcode results in an exit with error code
		return revertWithCode(0xf001c0de, fmt.Errorf("unknown instruction opcode: %d", opcode))
	}
	return nil
}
