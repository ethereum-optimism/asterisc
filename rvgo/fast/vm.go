package fast

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ethereum-optimism/asterisc/rvgo/riscv"
)

type UnsupportedSyscallErr struct {
	SyscallNum U64
}

func (e *UnsupportedSyscallErr) Error() string {
	return fmt.Sprintf("unsupported system call: %d", e.SyscallNum)
}

type UnrecognizedSyscallErr struct {
	SyscallNum U64
}

func (e *UnrecognizedSyscallErr) Error() string {
	return fmt.Sprintf("unrecognized system call: %d", e.SyscallNum)
}

type UnrecognizedResourceErr struct {
	Resource U64
}

func (e *UnrecognizedResourceErr) Error() string {
	return fmt.Sprintf("unrecognized resource limit lookup: %d", e.Resource)
}

// riscvStep runs a single instruction
// Note: errors are only returned in debugging/tooling modes, not in production use.
func (inst *InstrumentedState) riscvStep() (outErr error) {
	var revertCode uint64
	defer func() {
		if errInterface := recover(); errInterface != nil {
			if err, ok := errInterface.(error); ok {
				outErr = fmt.Errorf("revert: %w", err)
			} else {
				outErr = fmt.Errorf("revert: %v", err) // nolint:errorlint
			}

		}
		if revertCode != 0 {
			outErr = fmt.Errorf("revert %x: %w", revertCode, outErr)
		}
	}()

	revertWithCode := func(code uint64, err error) {
		revertCode = code
		panic(err)
	}

	//
	// Yul64 - functions to do 64 bit math - see yul64.go
	//

	//
	// Bit hacking util
	//

	//
	// State layout
	// N/A

	//
	// State loading
	//
	s := inst.state

	//
	// State output
	// N/A

	//
	// State access
	//

	// getMemRoot
	// setMemRoot

	getPreimageKey := func() [32]byte {
		return s.PreimageKey
	}
	setPreimageKey := func(k [32]byte) {
		s.PreimageKey = k
	}

	getPreimageOffset := func() U64 {
		return s.PreimageOffset
	}
	setPreimageOffset := func(v U64) {
		s.PreimageOffset = v
	}

	getPC := func() U64 {
		return s.PC
	}
	setPC := func(pc U64) {
		s.PC = pc
	}

	getExited := func() (exited bool) {
		return s.Exited
	}
	setExited := func() {
		s.Exited = true
	}

	// no getExitCode necessary
	setExitCode := func(v uint8) {
		s.ExitCode = v
	}

	getStep := func() U64 {
		return s.Step
	}
	setStep := func(v U64) {
		s.Step = v
	}

	getHeap := func() U64 {
		return s.Heap
	}
	setHeap := func(v U64) {
		s.Heap = v
	}

	getLoadReservation := func() U64 {
		return s.LoadReservation
	}
	setLoadReservation := func(addr U64) {
		s.LoadReservation = addr
	}

	getRegister := func(reg U64) U64 {
		if reg > 31 {
			revertWithCode(riscv.ErrInvalidRegister, fmt.Errorf("cannot load invalid register: %d", reg))
		}
		//fmt.Printf("load reg %2d: %016x\n", reg, state.Registers[reg])
		return s.Registers[reg]
	}
	setRegister := func(reg U64, v U64) {
		//fmt.Printf("write reg %2d: %016x   value: %016x\n", reg, state.Registers[reg], v)
		if reg == 0 { // reg 0 must stay 0
			// v is a HINT, but no hints are specified by standard spec, or used by us.
			return
		}
		if reg >= 32 {
			panic(fmt.Errorf("unknown register %d, cannot write %x", reg, v))
		}
		s.Registers[reg] = v
	}

	//
	// Parse - functions to parse RISC-V instructions - see parse.go
	//

	//
	// Memory functions
	//

	getMemoryB32 := func(addr U64, proofIndex uint8) (out [32]byte) {
		if addr&31 != 0 { // quick addr alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("addr %d not aligned with 32 bytes", addr))
		}
		inst.trackMemAccess(addr, proofIndex)
		s.Memory.GetUnaligned(addr, out[:])
		return
	}

	setMemoryB32 := func(addr U64, v [32]byte, proofIndex uint8) {
		if addr&31 != 0 {
			panic(fmt.Errorf("addr %d not aligned with 32 bytes", addr))
		}
		inst.verifyMemChange(addr, proofIndex)
		s.Memory.SetUnaligned(addr, v[:])
	}

	// load unaligned, optionally signed, little-endian, integer of 1 ... 8 bytes from memory
	loadMem := func(addr U64, size U64, signed bool, proofIndexL uint8, proofIndexR uint8) (out U64) {
		if size > 8 {
			revertWithCode(riscv.ErrLoadExceeds8Bytes, fmt.Errorf("cannot load more than 8 bytes: %d", size))
		}
		inst.trackMemAccess(addr&^31, proofIndexL)
		if (addr+size-1)&^31 != addr&^31 {
			if proofIndexR == 0xff {
				revertWithCode(riscv.ErrUnexpectedRProofLoad, fmt.Errorf("unexpected need for right-side proof %d in loadMem", proofIndexR))
			}
			inst.trackMemAccess((addr+size-1)&^31, proofIndexR)
		}
		var v [8]byte
		s.Memory.GetUnaligned(addr, v[:size])
		out = binary.LittleEndian.Uint64(v[:])
		bitSize := size << 3
		if signed && out&(1<<(bitSize-1)) != 0 { // if the last bit is set, then extend it to the full 64 bits
			out |= 0xFFFF_FFFF_FFFF_FFFF << bitSize
		} // otherwise just leave it zeroed
		//fmt.Printf("load mem: %016x  size: %d  value: %016x  signed: %v\n", addr, size, v, signed)
		return out
	}

	storeMemUnaligned := func(addr U64, size U64, value U256, proofIndexL uint8, proofIndexR uint8, verifyL bool, verifyR bool) {
		if size > 32 {
			revertWithCode(riscv.ErrStoreExceeds32Bytes, fmt.Errorf("cannot store more than 32 bytes: %d", size))
		}
		var bytez [32]byte
		binary.LittleEndian.PutUint64(bytez[:8], value[0])
		binary.LittleEndian.PutUint64(bytez[8:16], value[1])
		binary.LittleEndian.PutUint64(bytez[16:24], value[2])
		binary.LittleEndian.PutUint64(bytez[24:], value[3])

		leftAddr := addr &^ 31
		if verifyL {
			inst.trackMemAccess(leftAddr, proofIndexL)
		}
		inst.verifyMemChange(leftAddr, proofIndexL)
		if (addr+size-1)&^31 == addr&^31 { // if aligned
			s.Memory.SetUnaligned(addr, bytez[:size])
			return
		}
		if proofIndexR == 0xff {
			revertWithCode(riscv.ErrUnexpectedRProofStoreUnaligned, fmt.Errorf("unexpected need for right-side proof %d in storeMemUnaligned", proofIndexR))
		}
		// if not aligned
		rightAddr := leftAddr + 32
		leftSize := rightAddr - addr
		s.Memory.SetUnaligned(addr, bytez[:leftSize])
		if verifyR {
			inst.trackMemAccess(rightAddr, proofIndexR)
		}
		inst.verifyMemChange(rightAddr, proofIndexR)
		s.Memory.SetUnaligned(rightAddr, bytez[leftSize:size])
	}

	storeMem := func(addr U64, size U64, value U64, proofIndexL uint8, proofIndexR uint8, verifyL bool, verifyR bool) {
		if size > 8 {
			revertWithCode(riscv.ErrStoreExceeds8Bytes, fmt.Errorf("cannot store more than 8 bytes: %d", size))
		}
		var bytez [8]byte
		binary.LittleEndian.PutUint64(bytez[:], value)
		leftAddr := addr &^ 31
		if verifyL {
			inst.trackMemAccess(leftAddr, proofIndexL)
		}
		inst.verifyMemChange(leftAddr, proofIndexL)
		if (addr+size-1)&^31 == addr&^31 { // if aligned
			s.Memory.SetUnaligned(addr, bytez[:size])
			return
		}
		// if not aligned
		if proofIndexR == 0xff {
			revertWithCode(riscv.ErrUnexpectedRProofStore, fmt.Errorf("unexpected need for right-side proof %d in storeMem", proofIndexR))
		}
		rightAddr := leftAddr + 32
		leftSize := rightAddr - addr
		s.Memory.SetUnaligned(addr, bytez[:leftSize])
		if verifyR {
			inst.trackMemAccess(rightAddr, proofIndexR)
		}
		inst.verifyMemChange(rightAddr, proofIndexR)
		s.Memory.SetUnaligned(rightAddr, bytez[leftSize:size])
	}

	//
	// Preimage oracle interactions
	//
	writePreimageKey := func(addr U64, count U64) U64 {
		// adjust count down, so we only have to read a single 32 byte leaf of memory
		alignment := and64(addr, toU64(31))
		maxData := sub64(toU64(32), alignment)
		if gt64(count, maxData) != 0 {
			count = maxData
		}

		dat := b32asBEWord(getMemoryB32(sub64(addr, alignment), 1))
		// shift out leading bits
		dat = shl(u64ToU256(shl64(toU64(3), alignment)), dat)
		// shift to right end, remove trailing bits
		dat = shr(u64ToU256(shl64(toU64(3), sub64(toU64(32), count))), dat)

		bits := shl(toU256(3), u64ToU256(count))

		preImageKey := getPreimageKey()

		// Append to key content by bit-shifting
		key := b32asBEWord(preImageKey)
		key = shl(bits, key)
		key = or(key, dat)

		// We reset the pre-image value offset back to 0 (the right part of the merkle pair)
		setPreimageKey(beWordAsB32(key))
		setPreimageOffset(toU64(0))
		return count
	}

	readPreimageValue := func(addr U64, count U64) U64 {
		preImageKey := getPreimageKey()
		offset := getPreimageOffset()

		pdatB32, pdatlen, err := inst.readPreimage(preImageKey, offset) // pdat is left-aligned
		if err != nil {
			revertWithCode(riscv.ErrFailToReadPreimage, err)
		}
		if iszero64(pdatlen) { // EOF
			return toU64(0)
		}
		alignment := and64(addr, toU64(31))    // how many bytes addr is offset from being left-aligned
		maxData := sub64(toU64(32), alignment) // higher alignment leaves less room for data this step
		if gt64(count, maxData) != 0 {
			count = maxData
		}
		if gt64(count, pdatlen) != 0 { // cannot read more than pdatlen
			count = pdatlen
		}

		bits := shl64(toU64(3), sub64(toU64(32), count))             // 32-count, in bits
		mask := not(sub(shl(u64ToU256(bits), toU256(1)), toU256(1))) // left-aligned mask for count bytes
		alignmentBits := u64ToU256(shl64(toU64(3), alignment))
		mask = shr(alignmentBits, mask)                  // mask of count bytes, shifted by alignment
		pdat := shr(alignmentBits, b32asBEWord(pdatB32)) // pdat, shifted by alignment

		// update pre-image reader with updated offset
		newOffset := add64(offset, count)
		setPreimageOffset(newOffset)

		node := getMemoryB32(sub64(addr, alignment), 1)
		dat := and(b32asBEWord(node), not(mask)) // keep old bytes outside of mask
		dat = or(dat, and(pdat, mask))           // fill with bytes from pdat
		setMemoryB32(sub64(addr, alignment), beWordAsB32(dat), 1)
		return count
	}

	//
	// Syscall handling
	//
	sysCall := func() {
		a7 := getRegister(toU64(17))
		switch a7 {
		case riscv.SysExit: // exit the calling thread. No multi-thread support yet, so just exit.
			a0 := getRegister(toU64(10))
			setExitCode(uint8(a0))
			setExited()
			// program stops here, no need to change registers.
		case riscv.SysExitGroup: // exit-group
			a0 := getRegister(toU64(10))
			setExitCode(uint8(a0))
			setExited()
		case riscv.SysBrk: // brk
			// Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

			// brk(0) changes nothing about the memory, and returns the current page break
			v := shl64(toU64(30), toU64(1)) // set program break at 1 GiB
			setRegister(toU64(10), v)
			setRegister(toU64(11), toU64(0)) // no error
		case riscv.SysMmap: // mmap
			// A0 = addr (hint)
			addr := getRegister(toU64(10))
			// A1 = n (length)
			length := getRegister(toU64(11))
			// A2 = prot (memory protection type, can ignore)
			// A3 = flags (shared with other process and or written back to file)
			flags := getRegister(toU64(13))
			// A4 = fd (file descriptor, can ignore because we support anon memory only)
			fd := getRegister(toU64(14))
			// A5 = offset (offset in file, we don't support any non-anon memory, so we can ignore this)

			errCode := toU64(0)

			// ensure MAP_ANONYMOUS is set and fd == -1
			if (flags&0x20) == 0 || fd != u64Mask() {
				addr = u64Mask()
				errCode = toU64(0x4d) // EBADF
			} else {
				// ignore: prot, flags, fd, offset
				switch addr {
				case 0:
					// No hint, allocate it ourselves, by as much as the requested length.
					// Increase the length to align it with desired page size if necessary.
					align := and64(length, shortToU64(4095))
					if align != 0 {
						length = add64(length, sub64(shortToU64(4096), align))
					}
					prevHeap := getHeap()
					addr = prevHeap
					setHeap(add64(prevHeap, length)) // increment heap with length
					//fmt.Printf("mmap: 0x%016x (+ 0x%x increase)\n", s.Heap, length)
				default:
					// allow hinted memory address (leave it in A0 as return argument)
					//fmt.Printf("mmap: 0x%016x (0x%x allowed)\n", addr, length)
				}
			}
			setRegister(toU64(10), addr)
			setRegister(toU64(11), errCode)
		case riscv.SysRead: // read
			fd := getRegister(toU64(10))    // A0 = fd
			addr := getRegister(toU64(11))  // A1 = *buf addr
			count := getRegister(toU64(12)) // A2 = count
			var n U64
			var errCode U64
			switch fd {
			case riscv.FdStdin: // stdin
				n = toU64(0) // never read anything from stdin
				errCode = toU64(0)
			case riscv.FdHintRead: // hint-read
				// say we read it all, to continue execution after reading the hint-write ack response
				n = count
				errCode = toU64(0)
			case riscv.FdPreimageRead: // preimage read
				n = readPreimageValue(addr, count)
				errCode = toU64(0)
			default:
				n = u64Mask()         //  -1 (reading error)
				errCode = toU64(0x4d) // EBADF
			}
			setRegister(toU64(10), n)
			setRegister(toU64(11), errCode)
		case riscv.SysWrite: // write
			fd := getRegister(toU64(10))    // A0 = fd
			addr := getRegister(toU64(11))  // A1 = *buf addr
			count := getRegister(toU64(12)) // A2 = count
			var n U64
			var errCode U64
			switch fd {
			case riscv.FdStdout: // stdout
				_, err := io.Copy(inst.stdOut, s.Memory.ReadMemoryRange(addr, count))
				if err != nil {
					panic(fmt.Errorf("stdout writing err: %w", err))
				}
				n = count // write completes fully in single instruction step
				errCode = toU64(0)
			case riscv.FdStderr: // stderr
				_, err := io.Copy(inst.stdErr, s.Memory.ReadMemoryRange(addr, count))
				if err != nil {
					panic(fmt.Errorf("stderr writing err: %w", err))
				}
				n = count // write completes fully in single instruction step
				errCode = toU64(0)
			case riscv.FdHintWrite: // hint-write
				hintData, _ := io.ReadAll(s.Memory.ReadMemoryRange(addr, count))
				s.LastHint = append(inst.state.LastHint, hintData...)
				for len(s.LastHint) >= 4 { // process while there is enough data to check if there are any hints
					hintLen := binary.BigEndian.Uint32(s.LastHint[:4])
					if hintLen <= uint32(len(s.LastHint[4:])) {
						hint := s.LastHint[4 : 4+hintLen] // without the length prefix
						s.LastHint = s.LastHint[4+hintLen:]
						inst.preimageOracle.Hint(hint)
					} else {
						break // stop processing hints if there is incomplete data buffered
					}
				}
				n = count
				errCode = toU64(0)
			case riscv.FdPreimageWrite: // pre-image key write
				n = writePreimageKey(addr, count)
				errCode = toU64(0) // no error
			default: // any other file, including (3) hint read (5) preimage read
				n = u64Mask()         //  -1 (writing error)
				errCode = toU64(0x4d) // EBADF
			}
			setRegister(toU64(10), n)
			setRegister(toU64(11), errCode)
		case riscv.SysFcntl: // fcntl - file descriptor manipulation / info lookup
			fd := getRegister(toU64(10))  // A0 = fd
			cmd := getRegister(toU64(11)) // A1 = cmd
			var out U64
			var errCode U64
			switch cmd {
			case 0x1: // F_GETFD: get file descriptor flags
				switch fd {
				case 0: // stdin
					out = toU64(0) // no flag set
				case 1: // stdout
					out = toU64(0) // no flag set
				case 2: // stderr
					out = toU64(0) // no flag set
				case 3: // hint-read
					out = toU64(0) // no flag set
				case 4: // hint-write
					out = toU64(0) // no flag set
				case 5: // pre-image read
					out = toU64(0) // no flag set
				case 6: // pre-image write
					out = toU64(0) // no flag set
				default:
					out = u64Mask()
					errCode = toU64(0x4d) //EBADF
				}
			case 0x3: // F_GETFL: get file descriptor flags
				switch fd {
				case 0: // stdin
					out = toU64(0) // O_RDONLY
				case 1: // stdout
					out = toU64(1) // O_WRONLY
				case 2: // stderr
					out = toU64(1) // O_WRONLY
				case 3: // hint-read
					out = toU64(0) // O_RDONLY
				case 4: // hint-write
					out = toU64(1) // O_WRONLY
				case 5: // pre-image read
					out = toU64(0) // O_RDONLY
				case 6: // pre-image write
					out = toU64(1) // O_WRONLY
				default:
					out = u64Mask()
					errCode = toU64(0x4d) // EBADF
				}
			default: // no other commands: don't allow changing flags, duplicating FDs, etc.
				out = u64Mask()
				errCode = toU64(0x16) // EINVAL (cmd not recognized by this kernel)
			}
			setRegister(toU64(10), out)
			setRegister(toU64(11), errCode) // EBADF
		case riscv.SysOpenat: // openat - the Go linux runtime will try to open optional /sys/kernel files for performance hints
			setRegister(toU64(10), u64Mask())
			setRegister(toU64(11), toU64(0xd)) // EACCES - no access allowed
		case riscv.SysClockGettime: // clock_gettime
			addr := getRegister(toU64(11)) // addr of timespec struct
			// write 1337s + 42ns as time
			value := or(shortToU256(1337), shl(shortToU256(64), toU256(42)))
			storeMemUnaligned(addr, toU64(16), value, 1, 2, true, true)
			setRegister(toU64(10), toU64(0))
			setRegister(toU64(11), toU64(0))
		case riscv.SysClone: // clone - not supported
			setRegister(toU64(10), toU64(1))
			setRegister(toU64(11), toU64(0))
		case riscv.SysGetrlimit: // getrlimit
			res := getRegister(toU64(10))
			addr := getRegister(toU64(11))
			switch res {
			case 0x7: // RLIMIT_NOFILE
				// first 8 bytes: soft limit. 1024 file handles max open
				// second 8 bytes: hard limit
				storeMemUnaligned(addr, toU64(16), or(shortToU256(1024), shl(toU256(64), shortToU256(1024))), 1, 2, true, true)
				setRegister(toU64(10), toU64(0))
				setRegister(toU64(11), toU64(0))
			default:
				revertWithCode(riscv.ErrUnrecognizedResource, &UnrecognizedResourceErr{Resource: res})
			}
		case riscv.SysPrlimit64: // prlimit64 -- unsupported, we have getrlimit, is prlimit64 even called?
			revertWithCode(riscv.ErrInvalidSyscall, &UnsupportedSyscallErr{SyscallNum: a7})
		case riscv.SysFutex: // futex - not supported, for now
			revertWithCode(riscv.ErrInvalidSyscall, &UnsupportedSyscallErr{SyscallNum: a7})
		case riscv.SysNanosleep: // nanosleep - not supported, for now
			revertWithCode(riscv.ErrInvalidSyscall, &UnsupportedSyscallErr{SyscallNum: a7})
		default:
			// Ignore(no-op) unsupported system calls
			setRegister(toU64(10), toU64(0))
			setRegister(toU64(11), toU64(0))
			// List of ignored(no-op) syscalls used by op-program:
			// sched_getaffinity - hardcode to indicate affinity with any cpu-set mask
			// sched_yield - nothing to yield, synchronous execution only, for now
			// rt_sigprocmask - ignore any sigset changes
			// sigaltstack - ignore any hints of an alternative signal receiving stack addr
			// gettid - hardcode to 0
			// rt_sigaction - no-op, we never send signals, and thus need no sig handler info
			// madvise, epoll_create1, epoll_ctl, pipe2, readlinkat, newfstatat, newuname, munmap,
			// getrandom, ioctl, getcwd, getuid, getgid
		}
	}

	//
	// Instruction execution
	//

	if getExited() { // early exit if we can
		return nil
	}
	setStep(add64(getStep(), toU64(1)))

	pc := getPC()
	instr := loadMem(pc, toU64(4), false, 0, 0xff) // raw instruction

	// these fields are ignored if not applicable to the instruction type / opcode
	opcode := parseOpcode(instr)
	rd := parseRd(instr) // destination register index
	funct3 := parseFunct3(instr)
	rs1 := parseRs1(instr) // source register 1 index
	rs2 := parseRs2(instr) // source register 2 index
	funct7 := parseFunct7(instr)

	switch opcode {
	case 0x03: // 000_0011: memory loading
		// LB, LH, LW, LD, LBU, LHU, LWU
		imm := parseImmTypeI(instr)
		signed := iszero64(and64(funct3, toU64(4)))      // 4 = 100 -> bitflag
		size := shl64(and64(funct3, toU64(3)), toU64(1)) // 3 = 11 -> 1, 2, 4, 8 bytes size
		rs1Value := getRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		rdValue := loadMem(memIndex, size, signed, 1, 2)
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x23: // 010_0011: memory storing
		// SB, SH, SW, SD
		imm := parseImmTypeS(instr)
		size := shl64(funct3, toU64(1))
		value := getRegister(rs2)
		rs1Value := getRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
		storeMem(memIndex, size, value, 1, 2, true, true)
		setPC(add64(pc, toU64(4)))
	case 0x63: // 110_0011: branching
		rs1Value := getRegister(rs1)
		rs2Value := getRegister(rs2)
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
			// imm is a signed offset, in multiples of 2 bytes.
			// So it's really 13 bits with a hardcoded 0 bit.
			pc = add64(pc, imm)
		}
		// not like the other opcodes: nothing to write to rd register, and PC has already changed
		setPC(pc)
	case 0x13: // 001_0011: immediate arithmetic and logic
		rs1Value := getRegister(rs1)
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
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x1B: // 001_1011: immediate arithmetic and logic signed 32 bit
		rs1Value := getRegister(rs1)
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3 {
		case 0: // 000 = ADDIW
			rdValue = mask32Signed64(add64(rs1Value, imm))
		case 1: // 001 = SLLIW
			// SLLIW where imm[5] != 0 is reserved
			if and64(imm, toU64(0x20)) != 0 {
				revertWithCode(riscv.ErrInvalidSyscall, fmt.Errorf("illegal instruction %d: reserved instruction encoding", instr))
			}
			rdValue = mask32Signed64(shl64(and64(imm, toU64(0x1F)), rs1Value))
		case 5: // 101 = SR~
			// SRLIW and SRAIW where imm[5] != 0 is reserved
			if and64(imm, toU64(0x20)) != 0 {
				revertWithCode(riscv.ErrInvalidSyscall, fmt.Errorf("illegal instruction %d: reserved instruction encoding", instr))
			}
			shamt := and64(imm, toU64(0x1F))
			switch shr64(toU64(5), imm) { // top 7 bits select the shift type
			case 0x00: // 0000000 = SRLIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
			case 0x20: // 0100000 = SRAIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
			}
		}
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x33: // 011_0011: register arithmetic and logic
		rs1Value := getRegister(rs1)
		rs2Value := getRegister(rs2)
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
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x3B: // 011_1011: register arithmetic and logic in 32 bits
		rs1Value := getRegister(rs1)
		rs2Value := getRegister(rs2)
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
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x37: // 011_0111: LUI = Load upper immediate
		imm := parseImmTypeU(instr)
		rdValue := shl64(toU64(12), imm)
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x17: // 001_0111: AUIPC = Add upper immediate to PC
		imm := parseImmTypeU(instr)
		rdValue := add64(pc, signExtend64(shl64(toU64(12), imm), toU64(31)))
		setRegister(rd, rdValue)
		setPC(add64(pc, toU64(4)))
	case 0x6F: // 110_1111: JAL = Jump and link
		imm := parseImmTypeJ(instr)
		rdValue := add64(pc, toU64(4))
		setRegister(rd, rdValue)
		setPC(add64(pc, signExtend64(shl64(toU64(1), imm), toU64(20)))) // signed offset in multiples of 2 bytes (last bit is there, but ignored)
	case 0x67: // 110_0111: JALR = Jump and link register
		rs1Value := getRegister(rs1)
		imm := parseImmTypeI(instr)
		rdValue := add64(pc, toU64(4))
		setRegister(rd, rdValue)
		setPC(and64(add64(rs1Value, signExtend64(imm, toU64(11))), xor64(u64Mask(), toU64(1)))) // least significant bit is set to 0
	case 0x73: // 111_0011: environment things
		switch funct3 {
		case 0: // 000 = ECALL/EBREAK
			switch shr64(toU64(20), instr) { // I-type, top 12 bits
			case 0: // imm12 = 000000000000 ECALL
				sysCall()
				setPC(add64(pc, toU64(4)))
			default: // imm12 = 000000000001 EBREAK
				setPC(add64(pc, toU64(4))) // ignore breakpoint
			}
		default: // CSR instructions
			setRegister(rd, 0) // ignore CSR instructions
			setPC(add64(pc, toU64(4)))
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
			revertWithCode(riscv.ErrBadAMOSize, fmt.Errorf("bad AMO size: %d", size))
		}
		addr := getRegister(rs1)
		if addr&3 != 0 { // quick addr alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("addr %d not aligned with 4 bytes", addr))
		}

		op := shr64(toU64(2), funct7)
		switch op {
		case 0x2: // 00010 = LR = Load Reserved
			v := loadMem(addr, size, true, 1, 2)
			setRegister(rd, v)
			setLoadReservation(addr)
		case 0x3: // 00011 = SC = Store Conditional
			rdValue := toU64(1)
			if eq64(addr, getLoadReservation()) != 0 {
				rs2Value := getRegister(rs2)
				storeMem(addr, size, rs2Value, 1, 2, true, true)
				rdValue = toU64(0)
			}
			setRegister(rd, rdValue)
			setLoadReservation(toU64(0))
		default: // AMO: Atomic Memory Operation
			rs2Value := getRegister(rs2)
			if eq64(size, toU64(4)) != 0 {
				rs2Value = mask32Signed64(rs2Value)
			}
			value := rs2Value
			v := loadMem(addr, size, true, 1, 2)
			rdValue := v
			switch op {
			case 0x0: // 00000 = AMOADD = add
				v = add64(v, value)
			case 0x1: // 00001 = AMOSWAP
				v = value
			case 0x4: // 00100 = AMOXOR = xor
				v = xor64(v, value)
			case 0x8: // 01000 = AMOOR = or
				v = or64(v, value)
			case 0xc: // 01100 = AMOAND = and
				v = and64(v, value)
			case 0x10: // 10000 = AMOMIN = min signed
				if slt64(value, v) != 0 {
					v = value
				}
			case 0x14: // 10100 = AMOMAX = max signed
				if sgt64(value, v) != 0 {
					v = value
				}
			case 0x18: // 11000 = AMOMINU = min unsigned
				if lt64(value, v) != 0 {
					v = value
				}
			case 0x1c: // 11100 = AMOMAXU = max unsigned
				if gt64(value, v) != 0 {
					v = value
				}
			default:
				revertWithCode(riscv.ErrUnknownAtomicOperation, fmt.Errorf("unknown atomic operation %d", op))
			}
			storeMem(addr, size, v, 1, 3, false, true) // after overwriting 1, proof 2 is no longer valid
			setRegister(rd, rdValue)
		}
		setPC(add64(pc, toU64(4)))
	case 0x0F: // 000_1111: fence
		// Used to impose additional ordering constraints; flushing the mem operation pipeline.
		// This VM doesn't have a pipeline, nor additional harts, so this is a no-op.
		// FENCE / FENCE.TSO / FENCE.I all no-op: there's nothing to synchronize.
		setPC(add64(pc, toU64(4)))
	case 0x07: // FLW/FLD: floating point load word/double
		setPC(add64(pc, toU64(4))) // no-op this.
	case 0x27: // FSW/FSD: floating point store word/double
		setPC(add64(pc, toU64(4))) // no-op this.
	case 0x53: // FADD etc. no-op is enough to pass Go runtime check
		setPC(add64(pc, toU64(4))) // no-op this.
	default:
		revertWithCode(riscv.ErrUnknownOpCode, fmt.Errorf("unknown instruction opcode: %d", opcode))
	}
	return nil
}
