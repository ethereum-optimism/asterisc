package slow

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/asterisc/rvgo/riscv"
)

func decodeU64BE(v []byte) (out U64) {
	if len(v) != 8 {
		panic("bad u64 decode")
	}
	(*U256)(&out).SetUint64(binary.BigEndian.Uint64(v)) // mask out the lower bytes to get the size of uint we want
	return
}

func encodeU64BE(v U64) []byte {
	var dest [8]byte
	binary.BigEndian.PutUint64(dest[:], v.val())
	return dest[:]
}

const (
	stateSizeMemRoot         = 32
	stateSizePreimageKey     = 32
	stateSizePreimageOffset  = 8
	stateSizePC              = 8
	stateSizeExitCode        = 1
	stateSizeExited          = 1
	stateSizeStep            = 8
	stateSizeHeap            = 8
	stateSizeLoadReservation = 8
	stateSizeRegisters       = 8 * 32
)

const (
	stateOffsetMemRoot         = 0
	stateOffsetPreimageKey     = stateOffsetMemRoot + stateSizeMemRoot
	stateOffsetPreimageOffset  = stateOffsetPreimageKey + stateSizePreimageKey
	stateOffsetPC              = stateOffsetPreimageOffset + stateSizePreimageOffset
	stateOffsetExitCode        = stateOffsetPC + stateSizePC
	stateOffsetExited          = stateOffsetExitCode + stateSizeExitCode
	stateOffsetStep            = stateOffsetExited + stateSizeExited
	stateOffsetHeap            = stateOffsetStep + stateSizeStep
	stateOffsetLoadReservation = stateOffsetHeap + stateSizeHeap
	stateOffsetRegisters       = stateOffsetLoadReservation + stateSizeLoadReservation
	stateSize                  = stateOffsetRegisters + stateSizeRegisters
	paddedStateSize            = stateSize + ((32 - (stateSize % 32)) % 32)
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

type PreimageOracle interface {
	ReadPreimagePart(key [32]byte, offset uint64) (dat [32]byte, datlen uint8, err error)
}

func Step(calldata []byte, po PreimageOracle) (stateHash common.Hash, outErr error) {
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
	//

	//
	// Initial EVM memory / calldata checks
	//
	calldataload := func(offset U64) (out [32]byte) {
		copy(out[:], calldata[offset.val():])
		return
	}

	// First 4 bytes of keccak256("step(bytes,bytes,bytes32)")
	expectedSelector := []byte{0xe1, 0x4c, 0xed, 0x32}
	if len(calldata) < 4 || !bytes.Equal(calldata[:4], expectedSelector) {
		panic("invalid function selector")
	}

	stateContentOffset := uint8(4 + 32 + 32 + 32 + 32)
	if iszero(eq(add(b32asBEWord(calldataload(byteToU64(4))), shortToU256(32+4)), shortToU256(uint16(stateContentOffset)))) {
		// _stateData.offset = _stateData.pointer + 32 + 4
		// 32*4+4 = 132 expected state data offset
		panic("invalid state offset input")
	}

	if iszero(eq(b32asBEWord(calldataload(byteToU64(4+32*3))), shortToU256(stateSize))) { // user-provided state size must match expected state size
		panic("invalid state size input")
	}

	proofContentOffset := shortToU64(uint16(stateContentOffset) + paddedStateSize + 32)

	if mod(b32asBEWord(calldataload(shortToU64(uint16(stateContentOffset)+paddedStateSize))), shortToU256(60*32)) != byteToU256(0) {
		// proof offset must be stateContentOffset+paddedStateSize+32
		// proof size: 64-5+1=60 * 32 byte leaf,
		// but multiple memProof can be used, so the proofSize must be a multiple of 60
		panic("invalid proof size input")
	}

	if iszero(eq(add(b32asBEWord(calldataload(toU64(36))), shortToU256(32+4)), u64ToU256(proofContentOffset))) {
		// _proof.offset = proofContentOffset = _proof.pointer + 32 + 4
		panic("invalid proof offset input")
	}

	//
	// State loading
	//
	stateData := make([]byte, stateSize)
	copy(stateData, calldata[stateContentOffset:])

	//
	// State access
	//
	readState := func(offset uint64, length uint64) []byte {
		return stateData[offset : offset+length]
	}
	writeState := func(offset uint64, length uint64, data []byte) {
		if uint64(len(data)) != length {
			panic("unexpected input length")
		}
		copy(stateData[offset:offset+length], data)
	}
	getMemRoot := func() [32]byte {
		return *(*[32]byte)(readState(stateOffsetMemRoot, stateSizeMemRoot))
	}
	setMemRoot := func(v [32]byte) {
		writeState(stateOffsetMemRoot, stateSizeMemRoot, v[:])
	}

	getPreimageKey := func() [32]byte {
		return *(*[32]byte)(readState(stateOffsetPreimageKey, stateSizePreimageKey))
	}
	setPreimageKey := func(k [32]byte) {
		writeState(stateOffsetPreimageKey, stateSizePreimageKey, k[:])
	}

	getPreimageOffset := func() U64 {
		return decodeU64BE(readState(stateOffsetPreimageOffset, stateSizePreimageOffset))
	}
	setPreimageOffset := func(v U64) {
		writeState(stateOffsetPreimageOffset, stateSizePreimageOffset, encodeU64BE(v))
	}

	getPC := func() U64 {
		return decodeU64BE(readState(stateOffsetPC, stateSizePC))
	}
	setPC := func(pc U64) {
		writeState(stateOffsetPC, stateSizePC, encodeU64BE(pc))
	}

	getExited := func() (exited bool) {
		return stateData[stateOffsetExited] != 0
	}
	setExited := func() {
		stateData[stateOffsetExited] = 1
	}

	getExitCode := func() uint8 {
		return stateData[stateOffsetExitCode]
	}
	setExitCode := func(v uint8) {
		stateData[stateOffsetExitCode] = v
	}

	getStep := func() U64 {
		return decodeU64BE(readState(stateOffsetStep, stateSizeStep))
	}
	setStep := func(v U64) {
		writeState(stateOffsetStep, stateSizeStep, encodeU64BE(v))
	}

	getHeap := func() U64 {
		return decodeU64BE(readState(stateOffsetHeap, stateSizeHeap))
	}
	setHeap := func(v U64) {
		writeState(stateOffsetHeap, stateSizeHeap, encodeU64BE(v))
	}

	getLoadReservation := func() U64 {
		return decodeU64BE(readState(stateOffsetLoadReservation, stateSizeLoadReservation))
	}
	setLoadReservation := func(addr U64) {
		writeState(stateOffsetLoadReservation, stateSizeLoadReservation, encodeU64BE(addr))
	}

	getRegister := func(reg U64) U64 {
		if gt64(reg, byteToU64(31)) != (U64{}) {
			revertWithCode(riscv.ErrInvalidRegister, fmt.Errorf("cannot load invalid register: %d", reg.val()))
		}
		//fmt.Printf("load reg %2d: %016x\n", reg, state.Registers[reg])
		offset := add64(byteToU64(stateOffsetRegisters), mul64(reg, byteToU64(8)))
		return decodeU64BE(readState(offset.val(), 8))
	}
	setRegister := func(reg U64, v U64) {
		//fmt.Printf("write reg %2d: %016x   value: %016x\n", reg, state.Registers[reg], v)
		if iszero64(reg) { // reg 0 must stay 0
			// v is a HINT, but no hints are specified by standard spec, or used by us.
			return
		}
		if gt64(reg, byteToU64(31)) != (U64{}) {
			revertWithCode(riscv.ErrInvalidRegister, fmt.Errorf("unknown register %d, cannot write %x", reg.val(), v.val()))
		}
		offset := add64(byteToU64(stateOffsetRegisters), mul64(reg, byteToU64(8)))
		writeState(offset.val(), 8, encodeU64BE(v))
	}

	//
	// State output
	//
	vmStatus := func() (status uint8) {
		switch getExited() {
		case true:
			switch getExitCode() {
			case 0:
				status = 0 // VMStatusValid
			case 1:
				status = 1 // VMStatusInvalid
			default:
				status = 2 // VMStatusPanic
			}
		default:
			status = 3 // VMStatusUnfinished
		}
		return
	}

	computeStateHash := func() (out [32]byte) {
		out = crypto.Keccak256Hash(stateData)
		out[0] = vmStatus()
		return
	}

	//
	// Parse - functions to parse RISC-V instructions - see parse.go
	//

	//
	// Memory functions
	//
	proofOffset := func(proofIndex uint8) (offset U64) {
		// proof size: 64-5+1=60 (a 64-bit mem-address branch to 32 byte leaf, incl leaf itself), all 32 bytes
		offset = mul64(mul64(byteToU64(proofIndex), byteToU64(60)), byteToU64(32))
		offset = add64(offset, proofContentOffset)
		return
	}

	hashPair := func(a [32]byte, b [32]byte) (h [32]byte) {
		return crypto.Keccak256Hash(a[:], b[:])
	}

	getMemoryB32 := func(addr U64, proofIndex uint8) (out [32]byte) {
		if and64(addr, byteToU64(31)) != (U64{}) { // quick addr alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("addr %d not aligned with 32 bytes", addr))
		}
		offset := proofOffset(proofIndex)
		leaf := calldataload(offset)
		offset = add64(offset, byteToU64(32))

		path := shr64(byteToU64(5), addr) // 32 bytes of memory per leaf
		node := leaf                      // starting from the leaf node, work back up by combining with siblings, to reconstruct the root
		for i := uint8(0); i < 64-5; i++ {
			sibling := calldataload(offset)
			offset = add64(offset, byteToU64(32))
			switch and64(shr64(byteToU64(i), path), byteToU64(1)).val() {
			case 0:
				node = hashPair(node, sibling)
			case 1:
				node = hashPair(sibling, node)
			}
		}
		memRoot := getMemRoot()
		if iszero(eq(b32asBEWord(node), b32asBEWord(memRoot))) { // verify the root matches
			revertWithCode(riscv.ErrBadMemoryProof, fmt.Errorf("bad memory proof, got mem root: %x, expected %x", node, memRoot))
		}
		out = leaf
		return
	}

	// warning: setMemoryB32 does not verify the proof,
	// it assumes the same memory proof has been verified with getMemoryB32
	setMemoryB32 := func(addr U64, v [32]byte, proofIndex uint8) {
		if and64(addr, byteToU64(31)) != (U64{}) {
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("addr %d not aligned with 32 bytes", addr))
		}
		offset := proofOffset(proofIndex)
		leaf := v
		offset = add64(offset, byteToU64(32))
		path := shr64(byteToU64(5), addr) // 32 bytes of memory per leaf
		node := leaf                      // starting from the leaf node, work back up by combining with siblings, to reconstruct the root
		for i := uint8(0); i < 64-5; i++ {
			sibling := calldataload(offset)
			offset = add64(offset, byteToU64(32))

			switch and64(shr64(byteToU64(i), path), byteToU64(1)).val() {
			case 0:
				node = hashPair(node, sibling)
			case 1:
				node = hashPair(sibling, node)
			}
		}
		setMemRoot(node) // store new memRoot
	}

	// load unaligned, optionally signed, little-endian, integer of 1 ... 8 bytes from memory
	loadMem := func(addr U64, size U64, signed bool, proofIndexL uint8, proofIndexR uint8) (out U64) {
		if size.val() > 8 {
			revertWithCode(riscv.ErrLoadExceeds8Bytes, fmt.Errorf("cannot load more than 8 bytes: %d", size))
		}
		// load/verify left part
		leftAddr := and64(addr, not64(byteToU64(31)))
		left := b32asBEWord(getMemoryB32(leftAddr, proofIndexL))
		alignment := sub64(addr, leftAddr)

		right := U256{}
		rightAddr := and64(add64(addr, sub64(size, byteToU64(1))), not64(byteToU64(31)))
		leftShamt := sub64(sub64(byteToU64(32), alignment), size)
		rightShamt := byteToU64(0)
		if iszero64(eq64(leftAddr, rightAddr)) {
			// if unaligned, use second proof for the right part
			if proofIndexR == 0xff {
				revertWithCode(riscv.ErrUnexpectedRProofLoad, fmt.Errorf("unexpected need for right-side proof %d in loadMem", proofIndexR))
			}
			// load/verify right part
			right = b32asBEWord(getMemoryB32(rightAddr, proofIndexR))
			// left content is aligned to right of 32 bytes
			leftShamt = byteToU64(0)
			rightShamt = sub64(sub64(byteToU64(64), alignment), size)
		}

		// left: prepare for byte-taking by right-aligning
		left = shr(u64ToU256(shl64(byteToU64(3), leftShamt)), left)
		// right: right-align for byte-taking by right-aligning
		right = shr(u64ToU256(shl64(byteToU64(3), rightShamt)), right)
		// loop:
		for i := uint8(0); i < uint8(size.val()); i++ {
			// translate to reverse byte lookup, since we are reading little-endian memory, and need the highest byte first.
			// effAddr := (addr + size - 1 - i) &^ 31
			effAddr := and64(sub64(sub64(add64(addr, size), byteToU64(1)), byteToU64(i)), not64(byteToU64(31)))
			// take a byte from either left or right, depending on the effective address
			b := byteToU256(0)
			switch eq64(effAddr, leftAddr).val() {
			case 1:
				b = and(left, byteToU256(0xff))
				left = shr(byteToU256(8), left)
			case 0:
				b = and(right, byteToU256(0xff))
				right = shr(byteToU256(8), right)
			}
			// append it to the output
			out = or64(shl64(byteToU64(8), out), u256ToU64(b))
		}

		if signed {
			signBitShift := sub64(shl64(byteToU64(3), size), byteToU64(1))
			out = signExtend64(out, signBitShift)
		}
		return
	}

	// Splits the value into a left and a right part, each with a mask (identify data) and a patch (diff content).
	leftAndRight := func(alignment U64, size U64, value U256) (leftMask, rightMask, leftPatch, rightPatch U256) {
		start := alignment
		end := add64(alignment, size)
		for i := uint8(0); i < 64; i++ {
			index := byteToU64(i)
			leftSide := lt64(index, byteToU64(32))
			switch leftSide.val() {
			case 1:
				leftPatch = shl(byteToU256(8), leftPatch)
				leftMask = shl(byteToU256(8), leftMask)
			case 0:
				rightPatch = shl(byteToU256(8), rightPatch)
				rightMask = shl(byteToU256(8), rightMask)
			}
			if and64(eq64(lt64(index, start), byteToU64(0)), lt64(index, end)) != (U64{}) { // if alignment <= i < alignment+size
				b := and(shr(u64ToU256(shl64(byteToU64(3), sub64(index, alignment))), value), byteToU256(0xff))
				switch leftSide.val() {
				case 1:
					leftPatch = or(leftPatch, b)
					leftMask = or(leftMask, byteToU256(0xff))
				case 0:
					rightPatch = or(rightPatch, b)
					rightMask = or(rightMask, byteToU256(0xff))
				}
			}
		}
		return
	}

	storeMemUnaligned := func(addr U64, size U64, value U256, proofIndexL uint8, proofIndexR uint8) {
		if size.val() > 32 {
			revertWithCode(riscv.ErrStoreExceeds32Bytes, fmt.Errorf("cannot store more than 32 bytes: %d", size))
		}

		leftAddr := and64(addr, not64(byteToU64(31)))
		rightAddr := and64(add64(addr, sub64(size, byteToU64(1))), not64(byteToU64(31)))
		alignment := sub64(addr, leftAddr)
		leftMask, rightMask, leftPatch, rightPatch := leftAndRight(alignment, size, value)

		// load the left base
		left := b32asBEWord(getMemoryB32(leftAddr, proofIndexL))
		// apply the left patch
		left = or(and(left, not(leftMask)), leftPatch)
		// write the left
		setMemoryB32(leftAddr, beWordAsB32(left), proofIndexL)

		// if aligned: nothing more to do here
		if eq64(leftAddr, rightAddr) != (U64{}) {
			return
		}
		if proofIndexR == 0xff {
			revertWithCode(riscv.ErrUnexpectedRProofStoreUnaligned, fmt.Errorf("unexpected need for right-side proof %d in storeMemUnaligned", proofIndexR))
		}
		// load the right base (with updated mem root)
		right := b32asBEWord(getMemoryB32(rightAddr, proofIndexR))
		// apply the right patch
		right = or(and(right, not(rightMask)), rightPatch)
		// write the right (with updated mem root)
		setMemoryB32(rightAddr, beWordAsB32(right), proofIndexR)
	}
	storeMem := func(addr U64, size U64, value U64, proofIndexL uint8, proofIndexR uint8) {
		if size.val() > 8 {
			revertWithCode(riscv.ErrStoreExceeds8Bytes, fmt.Errorf("cannot store more than 8 bytes: %d", size))
		}

		storeMemUnaligned(addr, size, u64ToU256(value), proofIndexL, proofIndexR)
	}

	//
	// Preimage oracle interactions
	//
	writePreimageKey := func(addr U64, count U64) (out U64) {
		// adjust count down, so we only have to read a single 32 byte leaf of memory
		alignment := and64(addr, byteToU64(31))
		maxData := sub64(byteToU64(32), alignment)
		if gt64(count, maxData) != (U64{}) {
			count = maxData
		}

		dat := b32asBEWord(getMemoryB32(sub64(addr, alignment), 1))
		// shift out leading bits
		dat = shl(u64ToU256(shl64(byteToU64(3), alignment)), dat)
		// shift to right end, remove trailing bits
		dat = shr(u64ToU256(shl64(byteToU64(3), sub64(byteToU64(32), count))), dat)

		bits := shl(byteToU256(3), u64ToU256(count))

		preImageKey := getPreimageKey()

		// Append to key content by bit-shifting
		key := b32asBEWord(preImageKey)
		key = shl(bits, key)
		key = or(key, dat)

		// We reset the pre-image value offset back to 0 (the right part of the merkle pair)
		setPreimageKey(beWordAsB32(key))
		setPreimageOffset(byteToU64(0))
		out = count
		return
	}

	readPreimagePart := func(key [32]byte, offset U64) (dat [32]byte, datlen U64) {
		d, l, err := po.ReadPreimagePart(key, offset.val())
		if err == nil {
			dat = d
			datlen = byteToU64(l)
			return
		}
		revertWithCode(riscv.ErrFailToReadPreimage, err)
		return
	}

	readPreimageValue := func(addr U64, count U64) (out U64) {
		preImageKey := getPreimageKey()
		offset := getPreimageOffset()

		// make call to pre-image oracle contract
		pdatB32, pdatlen := readPreimagePart(preImageKey, offset)
		if iszero64(pdatlen) { // EOF
			out = byteToU64(0)
			return
		}
		alignment := and64(addr, byteToU64(31))    // how many bytes addr is offset from being left-aligned
		maxData := sub64(byteToU64(32), alignment) // higher alignment leaves less room for data this step
		if gt64(count, maxData) != (U64{}) {
			count = maxData
		}
		if gt64(count, pdatlen) != (U64{}) { // cannot read more than pdatlen
			count = pdatlen
		}

		bits := shl64(byteToU64(3), sub64(byteToU64(32), count))             // 32-count, in bits
		mask := not(sub(shl(u64ToU256(bits), byteToU256(1)), byteToU256(1))) // left-aligned mask for count bytes
		alignmentBits := u64ToU256(shl64(byteToU64(3), alignment))
		mask = shr(alignmentBits, mask)                  // mask of count bytes, shifted by alignment
		pdat := shr(alignmentBits, b32asBEWord(pdatB32)) // pdat, shifted by alignment

		// update pre-image reader with updated offset
		newOffset := add64(offset, count)
		setPreimageOffset(newOffset)

		node := getMemoryB32(sub64(addr, alignment), 1)
		dat := and(b32asBEWord(node), not(mask)) // keep old bytes outside of mask
		dat = or(dat, and(pdat, mask))           // fill with bytes from pdat
		setMemoryB32(sub64(addr, alignment), beWordAsB32(dat), 1)
		out = count
		return
	}

	//
	// Syscall handling
	//
	sysCall := func() {
		a7 := getRegister(byteToU64(17))
		switch a7.val() {
		case riscv.SysExit: // exit the calling thread. No multi-thread support yet, so just exit.
			a0 := getRegister(byteToU64(10))
			setExitCode(uint8(a0.val()))
			setExited()
			// program stops here, no need to change registers.
		case riscv.SysExitGroup: // exit-group
			a0 := getRegister(byteToU64(10))
			setExitCode(uint8(a0.val()))
			setExited()
		case riscv.SysBrk: // brk
			// Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

			// brk(0) changes nothing about the memory, and returns the current page break
			v := shl64(byteToU64(30), byteToU64(1)) // set program break at 1 GiB
			setRegister(byteToU64(10), v)
			setRegister(byteToU64(11), byteToU64(0)) // no error
		case riscv.SysMmap: // mmap
			// A0 = addr (hint)
			addr := getRegister(byteToU64(10))
			// A1 = n (length)
			length := getRegister(byteToU64(11))
			// A2 = prot (memory protection type, can ignore)
			// A3 = flags (shared with other process and or written back to file)
			flags := getRegister(byteToU64(13))
			// A4 = fd (file descriptor, can ignore because we support anon memory only)
			fd := getRegister(byteToU64(14))
			// A5 = offset (offset in file, we don't support any non-anon memory, so we can ignore this)

			errCode := byteToU64(0)

			// ensure MAP_ANONYMOUS is set and fd == -1
			if (flags.val()&0x20) == 0 || fd != u64Mask() {
				addr = u64Mask()
				errCode = byteToU64(0x4d) // no error
			} else {
				// ignore: prot, flags, fd, offset
				switch addr.val() {
				case 0:
					// No hint, allocate it ourselves, by as much as the requested length.
					// Increase the length to align it with desired page size if necessary.
					align := and64(length, shortToU64(4095))
					if align != (U64{}) {
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
			setRegister(byteToU64(10), addr)
			setRegister(byteToU64(11), errCode)
		case riscv.SysRead: // read
			fd := getRegister(byteToU64(10))    // A0 = fd
			addr := getRegister(byteToU64(11))  // A1 = *buf addr
			count := getRegister(byteToU64(12)) // A2 = count
			var n U64
			var errCode U64
			switch fd.val() {
			case riscv.FdStdin: // stdin
				n = byteToU64(0) // never read anything from stdin
				errCode = byteToU64(0)
			case riscv.FdHintRead: // hint-read
				// say we read it all, to continue execution after reading the hint-write ack response
				n = count
				errCode = byteToU64(0)
			case riscv.FdPreimageRead: // preimage read
				n = readPreimageValue(addr, count)
				errCode = byteToU64(0)
			default:
				n = u64Mask()             //  -1 (reading error)
				errCode = byteToU64(0x4d) // EBADF
			}
			setRegister(byteToU64(10), n)
			setRegister(byteToU64(11), errCode)
		case riscv.SysWrite: // write
			fd := getRegister(byteToU64(10))    // A0 = fd
			addr := getRegister(byteToU64(11))  // A1 = *buf addr
			count := getRegister(byteToU64(12)) // A2 = count
			var n U64
			var errCode U64
			switch fd.val() {
			case riscv.FdStdout: // stdout
				n = count // write completes fully in single instruction step
				errCode = byteToU64(0)
			case riscv.FdStderr: // stderr
				n = count // write completes fully in single instruction step
				errCode = byteToU64(0)
			case riscv.FdHintWrite: // hint-write
				n = count
				errCode = byteToU64(0)
			case riscv.FdPreimageWrite: // pre-image key write
				n = writePreimageKey(addr, count)
				errCode = byteToU64(0) // no error
			default: // any other file, including (3) hint read (5) preimage read
				n = u64Mask()             //  -1 (writing error)
				errCode = byteToU64(0x4d) // EBADF
			}
			setRegister(byteToU64(10), n)
			setRegister(byteToU64(11), errCode)
		case riscv.SysFcntl: // fcntl - file descriptor manipulation / info lookup
			fd := getRegister(byteToU64(10))  // A0 = fd
			cmd := getRegister(byteToU64(11)) // A1 = cmd
			var out U64
			var errCode U64
			switch cmd.val() {
			case 0x1: // F_GETFD: get file descriptor flags
				switch fd.val() {
				case 0: // stdin
					out = byteToU64(0) // no flag set
				case 1: // stdout
					out = byteToU64(0) // no flag set
				case 2: // stderr
					out = byteToU64(0) // no flag set
				case 3: // hint-read
					out = byteToU64(0) // no flag set
				case 4: // hint-write
					out = byteToU64(0) // no flag set
				case 5: // pre-image read
					out = byteToU64(0) // no flag set
				case 6: // pre-image write
					out = byteToU64(0) // no flag set
				default:
					out = u64Mask()
					errCode = byteToU64(0x4d) //EBADF
				}
			case 0x3: // F_GETFL: get file descriptor flags
				switch fd.val() {
				case 0: // stdin
					out = byteToU64(0) // O_RDONLY
				case 1: // stdout
					out = byteToU64(1) // O_WRONLY
				case 2: // stderr
					out = byteToU64(1) // O_WRONLY
				case 3: // hint-read
					out = byteToU64(0) // O_RDONLY
				case 4: // hint-write
					out = byteToU64(1) // O_WRONLY
				case 5: // pre-image read
					out = byteToU64(0) // O_RDONLY
				case 6: // pre-image write
					out = byteToU64(1) // O_WRONLY
				default:
					out = u64Mask()
					errCode = byteToU64(0x4d) // EBADF
				}
			default: // no other commands: don't allow changing flags, duplicating FDs, etc.
				out = u64Mask()
				errCode = byteToU64(0x16) // EINVAL (cmd not recognized by this kernel)
			}
			setRegister(byteToU64(10), out)
			setRegister(byteToU64(11), errCode) // EBADF
		case riscv.SysOpenat: // openat - the Go linux runtime will try to open optional /sys/kernel files for performance hints
			setRegister(byteToU64(10), u64Mask())
			setRegister(byteToU64(11), byteToU64(0xd)) // EACCES - no access allowed
		case riscv.SysClockGettime: // clock_gettime
			addr := getRegister(byteToU64(11)) // addr of timespec struct
			// write 1337s + 42ns as time
			value := or(shortToU256(1337), shl(shortToU256(64), byteToU256(42)))
			storeMemUnaligned(addr, byteToU64(16), value, 1, 2)
			setRegister(byteToU64(10), byteToU64(0))
			setRegister(byteToU64(11), byteToU64(0))
		case riscv.SysClone: // clone - not supported
			setRegister(byteToU64(10), byteToU64(1))
			setRegister(byteToU64(11), byteToU64(0))
		case riscv.SysGetrlimit: // getrlimit
			res := getRegister(byteToU64(10))
			addr := getRegister(byteToU64(11))
			switch res.val() {
			case 0x7: // RLIMIT_NOFILE
				// first 8 bytes: soft limit. 1024 file handles max open
				// second 8 bytes: hard limit
				storeMemUnaligned(addr, byteToU64(16), or(shortToU256(1024), shl(byteToU256(64), shortToU256(1024))), 1, 2)
				setRegister(byteToU64(10), byteToU64(0))
				setRegister(byteToU64(11), byteToU64(0))
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
			setRegister(byteToU64(10), byteToU64(0))
			setRegister(byteToU64(11), byteToU64(0))
		}
	}

	//
	// Instruction execution
	//

	if getExited() { // early exit if we can
		return computeStateHash(), nil
	}
	setStep(add64(getStep(), byteToU64(1)))

	pc := getPC()
	instr := loadMem(pc, byteToU64(4), false, 0, 0xff) // raw instruction

	// these fields are ignored if not applicable to the instruction type / opcode
	opcode := parseOpcode(instr)
	rd := parseRd(instr) // destination register index
	funct3 := parseFunct3(instr)
	rs1 := parseRs1(instr) // source register 1 index
	rs2 := parseRs2(instr) // source register 2 index
	funct7 := parseFunct7(instr)

	switch opcode.val() {
	case 0x03: // 000_0011: memory loading
		// LB, LH, LW, LD, LBU, LHU, LWU

		// bits[14:12] set to 111 are reserved
		if eq64(funct3, byteToU64(0x7)) != (U64{}) {
			revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction %d: reserved instruction encoding", instr))
		}

		imm := parseImmTypeI(instr)
		signed := iszero64(and64(funct3, byteToU64(4)))          // 4 = 100 -> bitflag
		size := shl64(and64(funct3, byteToU64(3)), byteToU64(1)) // 3 = 11 -> 1, 2, 4, 8 bytes size
		rs1Value := getRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, byteToU64(11)))
		rdValue := loadMem(memIndex, size, signed, 1, 2)
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x23: // 010_0011: memory storing
		// SB, SH, SW, SD
		imm := parseImmTypeS(instr)
		size := shl64(funct3, byteToU64(1))
		value := getRegister(rs2)
		rs1Value := getRegister(rs1)
		memIndex := add64(rs1Value, signExtend64(imm, byteToU64(11)))
		storeMem(memIndex, size, value, 1, 2)
		setPC(add64(pc, byteToU64(4)))
	case 0x63: // 110_0011: branching
		rs1Value := getRegister(rs1)
		rs2Value := getRegister(rs2)
		branchHit := byteToU64(0)
		switch funct3.val() {
		case 0: // 000 = BEQ
			branchHit = eq64(rs1Value, rs2Value)
		case 1: // 001 = BNE
			branchHit = and64(not64(eq64(rs1Value, rs2Value)), byteToU64(1))
		case 4: // 100 = BLT
			branchHit = slt64(rs1Value, rs2Value)
		case 5: // 101 = BGE
			branchHit = and64(not64(slt64(rs1Value, rs2Value)), byteToU64(1))
		case 6: // 110 = BLTU
			branchHit = lt64(rs1Value, rs2Value)
		case 7: // 111 = BGEU
			branchHit = and64(not64(lt64(rs1Value, rs2Value)), byteToU64(1))
		}
		switch branchHit.val() {
		case 0:
			pc = add64(pc, byteToU64(4))
		default:
			imm := parseImmTypeB(instr)
			// imm is a signed offset, in multiples of 2 bytes.
			// So it's really 13 bits with a hardcoded 0 bit.
			pc = add64(pc, imm)
		}

		// The PC must be aligned to 4 bytes.
		if and64(pc, byteToU64(3)) != (U64{}) {
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("pc %d not aligned with 4 bytes", pc))
		}

		// not like the other opcodes: nothing to write to rd register, and PC has already changed
		setPC(pc)
	case 0x13: // 001_0011: immediate arithmetic and logic
		rs1Value := getRegister(rs1)
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3.val() {
		case 0: // 000 = ADDI
			rdValue = add64(rs1Value, imm)
		case 1: // 001 = SLLI
			rdValue = shl64(and64(imm, byteToU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
		case 2: // 010 = SLTI
			rdValue = slt64(rs1Value, imm)
		case 3: // 011 = SLTIU
			rdValue = lt64(rs1Value, imm)
		case 4: // 100 = XORI
			rdValue = xor64(rs1Value, imm)
		case 5: // 101 = SR~
			switch shr64(byteToU64(6), imm).val() { // in rv64i the top 6 bits select the shift type
			case 0x00: // 000000 = SRLI
				rdValue = shr64(and64(imm, byteToU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
			case 0x10: // 010000 = SRAI
				rdValue = sar64(and64(imm, byteToU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid imm value %d for opcode 0x13", imm.val()))
			}
		case 6: // 110 = ORI
			rdValue = or64(rs1Value, imm)
		case 7: // 111 = ANDI
			rdValue = and64(rs1Value, imm)
		default:
			revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x13", funct3.val()))
		}
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x1B: // 001_1011: immediate arithmetic and logic signed 32 bit
		rs1Value := getRegister(rs1)
		imm := parseImmTypeI(instr)
		var rdValue U64
		switch funct3.val() {
		case 0: // 000 = ADDIW
			rdValue = mask32Signed64(add64(rs1Value, imm))
		case 1: // 001 = SLLIW
			// SLLIW where imm[5] != 0 is reserved
			if and64(imm, byteToU64(0x20)) != (U64{}) {
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction %d: reserved instruction encoding", instr))
			}
			rdValue = mask32Signed64(shl64(and64(imm, byteToU64(0x1F)), rs1Value))
		case 5: // 101 = SR~
			// SRLIW and SRAIW imm[5] != 0 is reserved
			if and64(imm, byteToU64(0x20)) != (U64{}) {
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction %d: reserved instruction encoding", instr))
			}
			shamt := and64(imm, byteToU64(0x1F))
			switch shr64(byteToU64(5), imm).val() { // top 7 bits select the shift type
			case 0x00: // 0000000 = SRLIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), byteToU64(31))
			case 0x20: // 0100000 = SRAIW
				rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(byteToU64(31), shamt))
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid imm value %d for opcode 0x1B", imm.val()))
			}
		default:
			revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x1B", funct3.val()))
		}
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x33: // 011_0011: register arithmetic and logic
		rs1Value := getRegister(rs1)
		rs2Value := getRegister(rs2)
		var rdValue U64
		switch funct7.val() {
		case 1: // RV M extension
			switch funct3.val() {
			case 0: // 000 = MUL: signed x signed
				rdValue = mul64(rs1Value, rs2Value)
			case 1: // 001 = MULH: upper bits of signed x signed
				rdValue = u256ToU64(shr(byteToU256(64), mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value))))
			case 2: // 010 = MULHSU: upper bits of signed x unsigned
				rdValue = u256ToU64(shr(byteToU256(64), mul(signExtend64To256(rs1Value), u64ToU256(rs2Value))))
			case 3: // 011 = MULHU: upper bits of unsigned x unsigned
				rdValue = u256ToU64(shr(byteToU256(64), mul(u64ToU256(rs1Value), u64ToU256(rs2Value))))
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
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x33", funct3.val()))
			}
		default:
			switch funct3.val() {
			case 0: // 000 = ADD/SUB
				switch funct7.val() {
				case 0x00: // 0000000 = ADD
					rdValue = add64(rs1Value, rs2Value)
				case 0x20: // 0100000 = SUB
					rdValue = sub64(rs1Value, rs2Value)
				default:
					revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct7 value %d for funct3 %d", funct7.val(), funct3.val()))
				}
			case 1: // 001 = SLL
				rdValue = shl64(and64(rs2Value, byteToU64(0x3F)), rs1Value) // only the low 6 bits are consider in RV6VI
			case 2: // 010 = SLT
				rdValue = slt64(rs1Value, rs2Value)
			case 3: // 011 = SLTU
				rdValue = lt64(rs1Value, rs2Value)
			case 4: // 100 = XOR
				rdValue = xor64(rs1Value, rs2Value)
			case 5: // 101 = SR~
				switch funct7.val() {
				case 0x00: // 0000000 = SRL
					rdValue = shr64(and64(rs2Value, byteToU64(0x3F)), rs1Value) // logical: fill with zeroes
				case 0x20: // 0100000 = SRA
					rdValue = sar64(and64(rs2Value, byteToU64(0x3F)), rs1Value) // arithmetic: sign bit is extended
				default:
					revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct7 value %d for funct3 %d", funct7.val(), funct3.val()))
				}
			case 6: // 110 = OR
				rdValue = or64(rs1Value, rs2Value)
			case 7: // 111 = AND
				rdValue = and64(rs1Value, rs2Value)
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x3B", funct3.val()))
			}
		}
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x3B: // 011_1011: register arithmetic and logic in 32 bits
		rs1Value := getRegister(rs1)
		rs2Value := and64(getRegister(rs2), u32Mask())
		var rdValue U64
		switch funct7.val() {
		case 1: // RV M extension
			switch funct3.val() {
			case 0: // 000 = MULW
				rdValue = mask32Signed64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
			case 4: // 100 = DIVW
				switch rs2Value.val() {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = mask32Signed64(sdiv64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
				}
			case 5: // 101 = DIVUW
				switch rs2Value.val() {
				case 0:
					rdValue = u64Mask()
				default:
					rdValue = mask32Signed64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				}
			case 6: // 110 = REMW
				switch rs2Value.val() {
				case 0:
					rdValue = mask32Signed64(rs1Value)
				default:
					rdValue = mask32Signed64(smod64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
				}
			case 7: // 111 = REMUW
				switch rs2Value.val() {
				case 0:
					rdValue = mask32Signed64(rs1Value)
				default:
					rdValue = mask32Signed64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				}
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x3B", funct3.val()))
			}
		default:
			switch funct3.val() {
			case 0: // 000 = ADDW/SUBW
				switch funct7.val() {
				case 0x00: // 0000000 = ADDW
					rdValue = mask32Signed64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				case 0x20: // 0100000 = SUBW
					rdValue = mask32Signed64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
				default:
					revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct7 value %d for funct3 %d", funct7.val(), funct3.val()))
				}
			case 1: // 001 = SLLW
				rdValue = mask32Signed64(shl64(and64(rs2Value, byteToU64(0x1F)), rs1Value))
			case 5: // 101 = SR~
				shamt := and64(rs2Value, byteToU64(0x1F))
				switch funct7.val() {
				case 0x00: // 0000000 = SRLW
					rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), byteToU64(31))
				case 0x20: // 0100000 = SRAW
					rdValue = signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(byteToU64(31), shamt))
				default:
					revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct7 value %d for funct3 %d", funct7.val(), funct3.val()))
				}
			default:
				revertWithCode(riscv.ErrIllegalInstruction, fmt.Errorf("illegal instruction: invalid funct3 value %d for opcode 0x3B", funct3.val()))
			}
		}
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x37: // 011_0111: LUI = Load upper immediate
		imm := parseImmTypeU(instr)
		rdValue := shl64(byteToU64(12), imm)
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x17: // 001_0111: AUIPC = Add upper immediate to PC
		imm := parseImmTypeU(instr)
		rdValue := add64(pc, signExtend64(shl64(byteToU64(12), imm), byteToU64(31)))
		setRegister(rd, rdValue)
		setPC(add64(pc, byteToU64(4)))
	case 0x6F: // 110_1111: JAL = Jump and link
		imm := parseImmTypeJ(instr)
		rdValue := add64(pc, byteToU64(4))
		setRegister(rd, rdValue)

		newPC := add64(pc, signExtend64(shl64(byteToU64(1), imm), byteToU64(20)))
		if and64(newPC, byteToU64(3)) != (U64{}) { // quick target alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("pc %d not aligned with 4 bytes", newPC))
		}
		setPC(newPC) // signed offset in multiples of 2 bytes (last bit is there, but ignored)
	case 0x67: // 110_0111: JALR = Jump and link register
		rs1Value := getRegister(rs1)
		imm := parseImmTypeI(instr)
		rdValue := add64(pc, byteToU64(4))
		setRegister(rd, rdValue)

		newPC := and64(add64(rs1Value, signExtend64(imm, byteToU64(11))), xor64(u64Mask(), byteToU64(1)))
		if and64(newPC, byteToU64(3)) != (U64{}) { // quick target alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("pc %d not aligned with 4 bytes", newPC))
		}
		setPC(newPC) // least significant bit is set to 0
	case 0x73: // 111_0011: environment things
		switch funct3.val() {
		case 0: // 000 = ECALL/EBREAK
			switch shr64(byteToU64(20), instr).val() { // I-type, top 12 bits
			case 0: // imm12 = 000000000000 ECALL
				sysCall()
				setPC(add64(pc, byteToU64(4)))
			default: // imm12 = 000000000001 EBREAK
				setPC(add64(pc, byteToU64(4))) // ignore breakpoint
			}
		default: // ignore CSR instructions
			setRegister(rd, byteToU64(0)) // ignore CSR instructions
			setPC(add64(pc, byteToU64(4)))
		}
	case 0x2F: // 010_1111: RV32A and RV32A atomic operations extension
		// acquire and release bits:
		//   aq := and64(shr64(byteToU64(1), funct7), byteToU64(1))
		//   rl := and64(funct7, byteToU64(1))
		// if none set: unordered
		// if aq is set: no following mem ops observed before acquire mem op
		// if rl is set: release mem op not observed before earlier mem ops
		// if both set: sequentially consistent
		// These are no-op here because there is no pipeline of mem ops to acquire/release.

		// 0b010 == RV32A W variants
		// 0b011 == RV64A D variants
		size := shl64(funct3, byteToU64(1))
		if or64(lt64(size, byteToU64(4)), gt64(size, byteToU64(8))) != (U64{}) {
			revertWithCode(riscv.ErrBadAMOSize, fmt.Errorf("bad AMO size: %d", size))
		}
		addr := getRegister(rs1)
		if mod64(addr, size) != (U64{}) { // quick addr alignment check
			revertWithCode(riscv.ErrNotAlignedAddr, fmt.Errorf("addr %d not aligned with 4 bytes", addr))
		}

		op := shr64(byteToU64(2), funct7)
		switch op.val() {
		case 0x2: // 00010 = LR = Load Reserved
			v := loadMem(addr, size, true, 1, 2)
			setRegister(rd, v)
			setLoadReservation(addr)
		case 0x3: // 00011 = SC = Store Conditional
			rdValue := byteToU64(1)
			if eq64(addr, getLoadReservation()) != (U64{}) {
				rs2Value := getRegister(rs2)
				storeMem(addr, size, rs2Value, 1, 2)
				rdValue = byteToU64(0)
			}
			setRegister(rd, rdValue)
			setLoadReservation(byteToU64(0))
		default: // AMO: Atomic Memory Operation
			rs2Value := getRegister(rs2)
			if eq64(size, byteToU64(4)) != (U64{}) {
				rs2Value = mask32Signed64(rs2Value)
			}
			value := rs2Value
			v := loadMem(addr, size, true, 1, 2)
			rdValue := v
			switch op.val() {
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
				if slt64(value, v) != (U64{}) {
					v = value
				}
			case 0x14: // 10100 = AMOMAX = max signed
				if sgt64(value, v) != (U64{}) {
					v = value
				}
			case 0x18: // 11000 = AMOMINU = min unsigned
				if lt64(value, v) != (U64{}) {
					v = value
				}
			case 0x1c: // 11100 = AMOMAXU = max unsigned
				if gt64(value, v) != (U64{}) {
					v = value
				}
			default:
				revertWithCode(riscv.ErrUnknownAtomicOperation, fmt.Errorf("unknown atomic operation %d", op))
			}
			storeMem(addr, size, v, 1, 3) // after overwriting 1, proof 2 is no longer valid
			setRegister(rd, rdValue)
		}
		setPC(add64(pc, byteToU64(4)))
	case 0x0F: // 000_1111: fence
		// Used to impose additional ordering constraints; flushing the mem operation pipeline.
		// This VM doesn't have a pipeline, nor additional harts, so this is a no-op.
		// FENCE / FENCE.TSO / FENCE.I all no-op: there's nothing to synchronize.
		setPC(add64(pc, byteToU64(4)))
	case 0x07: // FLW/FLD: floating point load word/double
		setPC(add64(pc, byteToU64(4))) // no-op this.
	case 0x27: // FSW/FSD: floating point store word/double
		setPC(add64(pc, byteToU64(4))) // no-op this.
	case 0x53: // FADD etc. no-op is enough to pass Go runtime check
		setPC(add64(pc, byteToU64(4))) // no-op this.
	default:
		revertWithCode(riscv.ErrUnknownOpCode, fmt.Errorf("unknown instruction opcode: %d", opcode))
	}
	return computeStateHash(), nil
}
