package fast

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/protolambda/asterisc/rvgo/oracle"
)

// page size must be at least 32 bytes (one merkle node)
// memory merkleization will look the same regardless of page size past 32.
const (
	pageAddrSize = 10
	pageKeySize  = 64 - pageAddrSize
	pageSize     = 1 << pageAddrSize
	pageAddrMask = pageSize - 1
	maxPageCount = 1 << pageKeySize
)

type VMState struct {
	PC uint64
	// sparse memory, pages of 1KB, keyed by page number: start memory address truncated by 10 bits.
	Memory    map[uint64]*[pageSize]byte
	Registers [32]uint64

	// 0xF14: mhartid  - riscv tests use this. Always hart 0, no parallelism supported
	CSR [4096]uint64 // 12 bit addressing space

	Exit   uint64
	Exited bool
	Heap   uint64 // for mmap to keep allocating new anon memory

	LoadReservation uint64

	PreimageKey         [2][32]byte // 0: type, 1: hash
	PreimageValueOffset uint64

	PreimageOracle func(typ [32]byte, key [32]byte) ([]byte, error)
}

func NewVMState() *VMState {
	return &VMState{
		Memory: make(map[uint64]*[1024]byte),
		Heap:   1 << 28, // 0.25 GiB of program code space
	}
}

func (state *VMState) Merkleize(so oracle.VMStateOracle) [32]byte {
	var zeroHashes [256][32]byte
	for i := 1; i < 256; i++ {
		zeroHashes[i] = so.Remember(zeroHashes[i-1], zeroHashes[i-1])
	}
	pageBranches := make(map[uint64]struct{})
	for pageKey := range state.Memory {
		pageGindex := (1 << pageKeySize) | pageKey
		for i := 0; i < pageKeySize; i++ {
			gindex := pageGindex >> i
			pageBranches[gindex] = struct{}{}
		}
	}
	merkleize := func(stackDepth uint64, getItem func(index uint64) [32]byte) [32]byte {
		stack := make([][32]byte, stackDepth+1)
		for i := uint64(0); i < (1 << stackDepth); i++ {
			v := getItem(i)
			for j := uint64(0); j <= stackDepth; j++ {
				if i&(1<<j) == 0 {
					stack[j] = v
					break
				} else {
					v = so.Remember(stack[j], v)
				}
			}
		}
		return stack[stackDepth]
	}
	uint64AsBytes32 := func(v uint64) (out [32]byte) {
		binary.LittleEndian.PutUint64(out[:8], v)
		return
	}
	merkleizePage := func(page *[1024]byte) [32]byte {
		return merkleize(pageAddrSize-5, func(index uint64) [32]byte { // 32 byte leaf values (5 bits)
			return *(*[32]byte)(page[index*32 : index*32+32])
		})
	}
	var merkleizeMemory func(gindex uint64, depth uint64) [32]byte
	merkleizeMemory = func(gindex uint64, depth uint64) [32]byte {
		if depth == pageKeySize {
			pageKey := gindex & ((1 << pageKeySize) - 1)
			return merkleizePage(state.Memory[pageKey])
		}
		left := gindex << 1
		right := left | 1
		var leftRoot, rightRoot [32]byte
		if _, ok := pageBranches[left]; ok {
			leftRoot = merkleizeMemory(left, depth+1)
		} else {
			leftRoot = zeroHashes[pageKeySize-depth+(pageAddrSize-5)]
		}
		if _, ok := pageBranches[right]; ok {
			rightRoot = merkleizeMemory(right, depth+1)
		} else {
			rightRoot = zeroHashes[pageKeySize-depth+(pageAddrSize-5)]
		}
		return so.Remember(leftRoot, rightRoot)
	}

	registersRoot := merkleize(5, func(index uint64) [32]byte {
		return uint64AsBytes32(state.Registers[index])
	})
	memoryRoot := merkleizeMemory(1, 0)
	csrRoot := merkleize(12, func(index uint64) [32]byte {
		return uint64AsBytes32(state.CSR[index])
	})
	return so.Remember(
		so.Remember(
			so.Remember(uint64AsBytes32(state.PC), memoryRoot), // 8, 9
			so.Remember(registersRoot, csrRoot),                // 10, 11
		),
		so.Remember(
			so.Remember(uint64AsBytes32(state.Exit), uint64AsBytes32(state.Heap)),     // 12, 13
			so.Remember(uint64AsBytes32(state.LoadReservation), state.PreimageKey[0]), // TODO pre-image state merkleization
		),
	)
}

func (state *VMState) loadOrCreatePage(pageIndex uint64) *[pageSize]byte {
	if pageIndex >= maxPageCount {
		panic("invalid page key")
	}
	p, ok := state.Memory[pageIndex]
	if ok {
		return p
	}
	p = &[pageSize]byte{}
	state.Memory[pageIndex] = p
	return p
}

func (state *VMState) loadMem(addr uint64, size uint64, signed bool) uint64 {
	if size > 8 {
		panic(fmt.Errorf("cannot load more than 8 bytes: %d", size))
	}
	var out [8]byte
	pageIndex := addr >> pageAddrSize
	p, ok := state.Memory[pageIndex]
	if !ok { // if page does not exist, then it's a 0 value
		return 0
	}
	copy(out[:], p[addr&pageAddrMask:])
	end := addr + size - 1 // Can also wrap around total memory.
	endPage := end >> pageAddrSize
	if pageIndex != endPage { // if it spans across two pages.
		if p, ok := state.Memory[endPage]; ok { // only if page exists, 0 otherwise
			remaining := (end & pageAddrMask) + 1
			copy(out[size-remaining:], p[:remaining])
		}
	}
	v := binary.LittleEndian.Uint64(out[:]) & ((1 << (size * 8)) - 1)
	if signed && v&((1<<(size<<3))>>1) != 0 { // if the last bit is set, then extend it to the full 64 bits
		v |= 0xFFFF_FFFF_FFFF_FFFF << (size << 3)
	} // otherwise just leave it zeroed
	//fmt.Printf("load mem: %016x  size: %d  value: %016x  signed: %v\n", addr, size, v, signed)
	return v
}

func (state *VMState) storeMem(addr uint64, size uint64, value uint64) {
	if size > 8 {
		panic(fmt.Errorf("cannot store more than 8 bytes: %d", size))
	}
	var bytez [8]byte
	binary.LittleEndian.PutUint64(bytez[:], value)
	pageIndex := addr >> pageAddrSize
	p, ok := state.Memory[pageIndex]
	if !ok { // create page if it does not exist
		p = &[pageSize]byte{}
		state.Memory[pageIndex] = p
	}
	copy(p[addr&pageAddrMask:], bytez[:size])
	end := addr + size - 1
	endPage := end >> pageAddrSize
	if pageIndex != endPage { // if it spans across two pages. Can also wrap around total memory.
		p, ok := state.Memory[endPage]
		if !ok { // create page if it does not exist
			p = &[pageSize]byte{}
			state.Memory[endPage] = p
		}
		remaining := (end & pageAddrMask) + 1
		copy(p[:remaining], bytez[size-remaining:])
	}
	//fmt.Printf("store mem: %016x  size: %d  value: %016x\n", addr, size, bytez[:size])
}

const (
	destADD uint64 = iota
	destSWAP
	destXOR
	destOR
	destAND
	destMIN
	destMAX
	destMINU
	destMAXU
)

func (state *VMState) opMem(op uint64, addr uint64, size uint64, value uint64) uint64 {
	v := state.loadMem(addr, size, true)
	out := v
	switch op {
	case destADD:
		v = add64(v, value)
	case destSWAP:
		v = value
	case destXOR:
		v = xor64(v, value)
	case destOR:
		v = or64(v, value)
	case destAND:
		v = and64(v, value)
	case destMIN:
		if slt64(value, v) != 0 {
			v = value
		}
	case destMAX:
		if sgt64(value, v) != 0 {
			v = value
		}
	case destMINU:
		if lt64(value, v) != 0 {
			v = value
		}
	case destMAXU:
		if gt64(value, v) != 0 {
			v = value
		}
	default:
		panic(fmt.Errorf("unrecognized mem op: %d", op))
	}
	state.storeMem(addr, size, v)
	return out
}

func (state *VMState) getPC() uint64 {
	return state.PC
}

func (state *VMState) setPC(pc uint64) {
	state.PC = pc
}

func (state *VMState) loadRegister(reg uint64) uint64 {
	//fmt.Printf("load reg %2d: %016x\n", reg, state.Registers[reg])
	return state.Registers[reg]
}

func (state *VMState) writeCSR(num uint64, v uint64) {
	state.CSR[num] = v
}

func (state *VMState) readCSR(num uint64) uint64 {
	return state.CSR[num]
}

func (state *VMState) setLoadReservation(addr uint64) {
	state.LoadReservation = addr
}

func (state *VMState) getLoadReservation() uint64 {
	return state.LoadReservation
}

func (state *VMState) writeRegister(reg uint64, v uint64) {
	//fmt.Printf("write reg %2d: %016x   value: %016x\n", reg, state.Registers[reg], v)
	if reg == 0 { // reg 0 must stay 0
		// v is a HINT, but no hints are specified by standard spec, or used by us.
		return
	}
	if reg >= 32 {
		panic(fmt.Errorf("unknown register %d, cannot write %x", reg, v))
	}
	state.Registers[reg] = v
}

// readU256AlignedMemory reads bytes starting at addr from memory, no more than count are read.
// At most 32 bytes are read.
// If addr is not aligned to a multiple of 32 bytes, less bytes may be read,
// to contain reading to a single 32-byte leaf.
// A container with the data (zeroed right-padding if necessary),
// and the length of the read data in bits (!!!), is returned.
func (state *VMState) readU256AlignedMemory(addr uint64, count uint64) (dat U256, bits uint64) {
	// find alignment, and reduce work if it is not aligned
	alignment := addr % 32
	// how many bytes we can read from this bytes32
	maxData := 32 - alignment
	// reduce count accordingly, if necessary
	if count > maxData {
		count = maxData
	}
	bits = shl64(toU64(8), count)

	// make sure addr is aligned with 32 bits
	addr = addr & ^uint64(0x1f)

	// load the key data
	pageIndex := addr >> pageAddrSize
	p, ok := state.Memory[pageIndex]
	if !ok { // default to zeroed data if page does not exist
		return U256{}, bits
	}

	// Load the relevant bytes32
	pageAddr := addr & pageAddrMask
	dat = b32asBEWord(*(*[32]byte)(p[pageAddr : pageAddr+32]))

	// shift out prefix and align with start of output
	alignmentInBits := u64ToU256(shl64(toU64(3), alignment))
	dat = shl(alignmentInBits, dat)
	// remove suffix
	shamt := u64ToU256(sub64(U64(256), bits))
	dat = shl(shamt, shr(shamt, dat))
	return
}

// writeU256AlignedMemory writes up to count bytes of the given data to the memory at the address.
// At most 32 bytes are written.
// If addr is not aligned to a multiple of 32 bytes, less bytes may be written,
// to contain writing to a single 32-byte leaf.
// The length of the written data in bytes is returned.
func (state *VMState) writeU256AlignedMemory(addr uint64, count uint64, dat [32]byte) uint64 {
	// find alignment, and reduce work if it is not aligned
	alignment := addr % 32
	// how many bytes we can write of this bytes32
	maxData := 32 - alignment
	// reduce count accordingly, if necessary
	if count > maxData {
		count = maxData
	}

	// make sure addr is aligned with 32 bits
	addr = addr & ^uint64(0x1f)

	// load the key data
	pageIndex := addr >> pageAddrSize
	p, ok := state.Memory[pageIndex]
	if !ok { // create page if it doesn't exist yet
		p = &[pageSize]byte{}
		state.Memory[pageIndex] = p
	}

	// overwrite the leaf part
	pageAddr := addr & pageAddrMask
	prev := p[pageAddr : pageAddr+32]
	copy(prev[alignment:alignment+count], dat[:count])
	return count
}

func (state *VMState) writePreimageKey(addr uint64, count uint64) uint64 {
	dat, bits := state.readU256AlignedMemory(addr, count)

	// Append to key type, key content using bitshifts
	key0 := b32asBEWord(state.PreimageKey[0])
	key1 := b32asBEWord(state.PreimageKey[1])
	key1 = shl(u64ToU256(bits), key1)
	key1 = or(key1, shr(u64ToU256(sub64(U64(256), bits)), key0)) // bits overflow from key0 to key1
	key0 = shl(u64ToU256(bits), key0)
	key0 = or(key0, dat)
	state.PreimageKey[0] = beWordAsB32(key0)
	state.PreimageKey[1] = beWordAsB32(key1)
	state.PreimageValueOffset = 0
	return shr64(toU64(3), bits)
}

func (state *VMState) readPreimageValue(addr uint64, size uint64) (uint64, error) {
	preimage, err := state.PreimageOracle(state.PreimageKey[0], state.PreimageKey[1])
	if err != nil {
		return 0, fmt.Errorf("failed to get preimage (%x, %x): %w", state.PreimageKey[0], state.PreimageKey[1], err)
	}
	preimageSize := uint64(len(preimage))
	remaining := preimageSize - state.PreimageValueOffset
	n := size
	if n > 32 {
		n = 32
	}
	if n > remaining {
		n = remaining
	}
	var x [32]byte
	copy(x[:], preimage[state.PreimageValueOffset:state.PreimageValueOffset+n])
	n = state.writeU256AlignedMemory(addr, n, x)
	state.PreimageValueOffset += n
	return n, nil
}

type memReader struct {
	state *VMState
	addr  uint64
	count uint64
}

func (r *memReader) Read(dest []byte) (n int, err error) {
	if r.count == 0 {
		return 0, io.EOF
	}

	// Keep iterating over memory until we have all our data.
	// It may wrap around the address range, and may not be aligned
	endAddr := r.addr + r.count

	pageIndex := r.addr >> pageAddrSize
	start := r.addr & pageAddrMask
	end := uint64(pageSize)

	if pageIndex == (endAddr >> pageAddrSize) {
		end = endAddr & pageAddrMask
	}
	p, ok := r.state.Memory[pageIndex]
	if ok {
		n = copy(dest, p[start:end])
	} else {
		n = copy(dest, make([]byte, end-start)) // default to zeroes
	}
	r.addr += uint64(n)
	r.count -= uint64(n)
	return n, nil
}

func (state *VMState) memRange(addr uint64, count uint64) io.Reader {
	return &memReader{state: state, addr: addr, count: count}
}

func (state *VMState) setRange(addr uint64, count uint64, r io.Reader) error {
	end := addr + count
	for addr := addr; addr < end; {
		// map address to page index, and start within page
		page := state.loadOrCreatePage(addr >> pageAddrSize)
		pageStart := addr & pageAddrMask
		// copy till end of page
		pageEnd := uint64(pageSize)
		// unless we reached the end
		if (addr&^pageAddrMask)+pageSize > end {
			pageEnd = end & pageAddrMask
		}
		if _, err := io.ReadFull(r, page[pageStart:pageEnd]); err != nil {
			return fmt.Errorf("failed to read data into memory %d: %w", pageStart, err)
		}
		addr += pageEnd - pageStart
	}
	return nil
}

func (state *VMState) Instr() uint32 {
	var out [4]byte
	_, _ = io.ReadFull(state.memRange(state.PC, 4), out[:])
	return binary.LittleEndian.Uint32(out[:])
}
