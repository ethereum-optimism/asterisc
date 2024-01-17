package fast

import (
	"encoding/json"
	"fmt"
	"io"
	"math/bits"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
)

// Note: 2**12 = 4 KiB, the minimum page-size in Unicorn for mmap
// as well as the Go runtime min phys page size.
const (
	PageAddrSize = 12
	PageKeySize  = 64 - PageAddrSize
	PageSize     = 1 << PageAddrSize
	PageAddrMask = PageSize - 1
	MaxPageCount = 1 << PageKeySize
	PageKeyMask  = MaxPageCount - 1
	ProofLen     = 64 - 4
	branchDepth  = 4
	branchMask   = (1 << branchDepth) - 1
	branchFactor = 1 << branchDepth
	halfBranch   = branchFactor >> 1
	levels       = 13
	// log2(16) * 13 = 4 * 13 = 52 = PageKeySize
)

func HashPair(left, right [32]byte) [32]byte {
	out := crypto.Keccak256Hash(left[:], right[:])
	//fmt.Printf("0x%x 0x%x -> 0x%x\n", left, right, out)
	return out
}

var zeroHashes = func() [256][32]byte {
	// empty parts of the tree are all zero. Precompute the hash of each full-zero range sub-tree level.
	var out [256][32]byte
	for i := 1; i < 256; i++ {
		out[i] = HashPair(out[i-1], out[i-1])
	}
	return out
}()

type MemLevel interface {
	Invalidate(addr uint64)
	MerkleRoot() [32]byte
	ForEachPage(fn func(pageIndex uint64, page *Page) error) error
	Prove(addr uint64) (proof [][32]byte)
}

type MemNode[C MemLevel] struct {
	Radix  [branchFactor]C
	Merkle [branchFactor][32]byte
	Cache  uint16
	Depth  uint8 // bits, from bottom to top
}

func (mn *MemNode[C]) Invalidate(addr uint64) {
	i := (addr >> mn.Depth) & branchMask
	mn.Radix[i].Invalidate(addr)
	// TODO: replace this with constant lookup (since there are only 16 different i values)
	gi := (1 << branchDepth) | i
	for gi > 0 {
		gi >>= 1
		mn.Cache &^= 1 << gi
	}
}

func (mn *MemNode[C]) MerkleRoot() [32]byte {
	for i := 0; i < halfBranch; i++ {
		if mn.Cache&(1<<(halfBranch|i)) == 0 {
			mn.Merkle[halfBranch|i] = HashPair(mn.Radix[i<<1].MerkleRoot(), mn.Radix[(i<<1)|1].MerkleRoot())
		}
	}
	for i := halfBranch; i > 0; i-- {
		if mn.Cache&(1<<i) == 0 {
			mn.Merkle[i] = HashPair(mn.Merkle[i<<1], mn.Merkle[(i<<1)|1])
		}
	}
	return mn.Merkle[1]
}

func (mn *MemNode[C]) ForEachPage(fn func(pageIndex uint64, page *Page) error) error {
	for i, c := range mn.Radix {
		if err := c.ForEachPage(fn); err != nil {
			return fmt.Errorf("node %d failed ForEachPage: %w", i, err)
		}
	}
	return nil
}

func (mn *MemNode[C]) Prove(addr uint64) (proof [][32]byte) {
	i := (addr >> mn.Depth) & branchMask
	proof = mn.Radix[i].Prove(addr) // sub-proof at the target sub-tree
	_ = mn.MerkleRoot() // update cache
	proof = append(proof, mn.Radix[i ^ 1].MerkleRoot()) // append sibling
	gi := branchFactor | i
	for gi > 1 { // add remaining nodes from cache
		gi >>= 1
		proof = append(proof, mn.Merkle[gi ^ 1])
	}
	return
}

type L1 = MemNode[*CachedPage]
type L2 = MemNode[*L1]
type L3 = MemNode[*L2]
type L4 = MemNode[*L3]

type Memory struct {
	// generalized index -> merkle root or nil if invalidated
	nodes map[uint64]*[32]byte

	// pageIndex -> cached page
	pages map[uint64]*CachedPage

	// Note: since we don't de-alloc pages, we don't do ref-counting.
	// Once a page exists, it doesn't leave memory

	// two caches: we often read instructions from one page, and do memory things with another page.
	// this prevents map lookups each instruction
	lastPageKeys [2]uint64
	lastPage     [2]*CachedPage
}

func NewMemory() *Memory {
	return &Memory{
		nodes:        make(map[uint64]*[32]byte),
		pages:        make(map[uint64]*CachedPage),
		lastPageKeys: [2]uint64{^uint64(0), ^uint64(0)}, // default to invalid keys, to not match any pages
	}
}

func (m *Memory) PageCount() int {
	return len(m.pages)
}

func (m *Memory) ForEachPage(fn func(pageIndex uint64, page *Page) error) error {
	for pageIndex, cachedPage := range m.pages {
		if err := fn(pageIndex, cachedPage.Data); err != nil {
			return err
		}
	}
	return nil
}

func (m *Memory) Invalidate(addr uint64) {
	// find page, and invalidate addr within it
	if p, ok := m.pageLookup(addr >> PageAddrSize); ok {
		prevValid := p.Ok[1]
		p.Invalidate(addr & PageAddrMask)
		if !prevValid { // if the page was already invalid before, then nodes to mem-root will also still be.
			return
		}
	} else { // no page? nothing to invalidate
		return
	}

	// find the gindex of the first page covering the address
	gindex := (uint64(addr) >> PageAddrSize) | (1 << (64 - PageAddrSize))

	for gindex > 0 {
		m.nodes[gindex] = nil
		gindex >>= 1
	}
}

func (m *Memory) MerkleizeSubtree(gindex uint64) [32]byte {
	l := uint64(bits.Len64(gindex))
	if l > ProofLen {
		panic("gindex too deep")
	}
	if l > PageKeySize {
		depthIntoPage := l - 1 - PageKeySize
		pageIndex := (gindex >> depthIntoPage) & PageKeyMask
		if p, ok := m.pages[uint64(pageIndex)]; ok {
			pageGindex := (1 << depthIntoPage) | (gindex & ((1 << depthIntoPage) - 1))
			return p.MerkleizeSubtree(pageGindex)
		} else {
			return zeroHashes[64-5+1-l] // page does not exist
		}
	}
	if l > PageKeySize+1 {
		panic("cannot jump into intermediate node of page")
	}
	n, ok := m.nodes[gindex]
	if !ok {
		// if the node doesn't exist, the whole sub-tree is zeroed
		return zeroHashes[64-5+1-l]
	}
	if n != nil {
		return *n
	}
	left := m.MerkleizeSubtree(gindex << 1)
	right := m.MerkleizeSubtree((gindex << 1) | 1)
	r := HashPair(left, right)
	m.nodes[gindex] = &r
	return r
}

func (m *Memory) MerkleProof(addr uint64) (out [ProofLen * 32]byte) {
	proof := m.traverseBranch(1, addr, 0)
	// encode the proof
	for i := 0; i < ProofLen; i++ {
		copy(out[i*32:(i+1)*32], proof[i][:])
	}
	return out
}

func (m *Memory) traverseBranch(parent uint64, addr uint64, depth uint8) (proof [][32]byte) {
	if depth == ProofLen-1 {
		proof = make([][32]byte, 0, ProofLen)
		proof = append(proof, m.MerkleizeSubtree(parent))
		return
	}
	if depth > ProofLen-1 {
		panic("traversed too deep")
	}
	self := parent << 1
	sibling := self | 1
	if addr&(1<<(63-depth)) != 0 {
		self, sibling = sibling, self
	}
	proof = m.traverseBranch(self, addr, depth+1)
	siblingNode := m.MerkleizeSubtree(sibling)
	proof = append(proof, siblingNode)
	return
}

func (m *Memory) MerkleRoot() [32]byte {
	return m.MerkleizeSubtree(1)
}

func (m *Memory) pageLookup(pageIndex uint64) (*CachedPage, bool) {
	// hit caches
	if pageIndex == m.lastPageKeys[0] {
		return m.lastPage[0], true
	}
	if pageIndex == m.lastPageKeys[1] {
		return m.lastPage[1], true
	}
	p, ok := m.pages[pageIndex]

	// only cache existing pages.
	if ok {
		m.lastPageKeys[1] = m.lastPageKeys[0]
		m.lastPage[1] = m.lastPage[0]
		m.lastPageKeys[0] = pageIndex
		m.lastPage[0] = p
	}

	return p, ok
}

// TODO: we never do unaligned writes, this should be simplified
func (m *Memory) SetUnaligned(addr uint64, dat []byte) {
	if len(dat) > 32 {
		panic("cannot set more than 32 bytes")
	}
	pageIndex := addr >> PageAddrSize
	pageAddr := addr & PageAddrMask
	p, ok := m.pageLookup(pageIndex)
	if !ok {
		// allocate the page if we have not already.
		// Go may mmap relatively large ranges, but we only allocate the pages just in time.
		p = m.AllocPage(pageIndex)
	} else {
		m.Invalidate(addr) // invalidate this branch of memory, now that the value changed
	}

	d := copy(p.Data[pageAddr:], dat)
	if d == len(dat) {
		return // if all the data fitted in the page, we're done
	}

	// continue to remaining part
	addr += uint64(d)
	pageIndex = addr >> PageAddrSize
	pageAddr = addr & PageAddrMask
	p, ok = m.pageLookup(pageIndex)
	if !ok {
		// allocate the page if we have not already.
		// Go may mmap relatively large ranges, but we only allocate the pages just in time.
		p = m.AllocPage(pageIndex)
	} else {
		m.Invalidate(addr) // invalidate this branch of memory, now that the value changed
	}

	copy(p.Data[pageAddr:], dat)
}

func (m *Memory) GetUnaligned(addr uint64, dest []byte) {
	if len(dest) > 32 {
		panic("cannot get more than 32 bytes")
	}
	pageIndex := addr >> PageAddrSize
	pageAddr := addr & PageAddrMask
	p, ok := m.pageLookup(pageIndex)
	var d int
	if !ok {
		l := pageSize - pageAddr
		if l > 32 {
			l = 32
		}
		var zeroes [32]byte
		d = copy(dest, zeroes[:l])
	} else {
		d = copy(dest, p.Data[pageAddr:])
	}

	if d == len(dest) {
		return // if all the data fitted in the page, we're done
	}

	// continue to remaining part
	addr += uint64(d)
	pageIndex = addr >> PageAddrSize
	pageAddr = addr & PageAddrMask
	p, ok = m.pageLookup(pageIndex)
	if !ok {
		l := pageSize - pageAddr
		if l > 32 {
			l = 32
		}
		var zeroes [32]byte
		d = copy(dest[d:], zeroes[:l])
	} else {
		copy(dest[d:], p.Data[pageAddr:])
	}
}

func (m *Memory) AllocPage(pageIndex uint64) *CachedPage {
	p := &CachedPage{Data: new(Page), PageIndex: pageIndex}
	m.pages[pageIndex] = p
	// make nodes to root
	k := (1 << PageKeySize) | uint64(pageIndex)
	for k > 0 {
		m.nodes[k] = nil
		k >>= 1
	}
	return p
}

type pageEntry struct {
	Index uint64 `json:"index"`
	Data  *Page  `json:"data"`
}

func (m *Memory) MarshalJSON() ([]byte, error) {
	pages := make([]pageEntry, 0, len(m.pages))
	for k, p := range m.pages {
		pages = append(pages, pageEntry{
			Index: k,
			Data:  p.Data,
		})
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Index < pages[j].Index
	})
	return json.Marshal(pages)
}

func (m *Memory) UnmarshalJSON(data []byte) error {
	var pages []pageEntry
	if err := json.Unmarshal(data, &pages); err != nil {
		return err
	}
	m.nodes = make(map[uint64]*[32]byte)
	m.pages = make(map[uint64]*CachedPage)
	m.lastPageKeys = [2]uint64{^uint64(0), ^uint64(0)}
	m.lastPage = [2]*CachedPage{nil, nil}
	for i, p := range pages {
		if _, ok := m.pages[p.Index]; ok {
			return fmt.Errorf("cannot load duplicate page, entry %d, page index %d", i, p.Index)
		}
		m.AllocPage(p.Index).Data = p.Data
	}
	return nil
}

func (m *Memory) SetMemoryRange(addr uint64, r io.Reader) error {
	for {
		pageIndex := addr >> PageAddrSize
		pageAddr := addr & PageAddrMask
		p, ok := m.pageLookup(pageIndex)
		if !ok {
			p = m.AllocPage(pageIndex)
		}
		p.InvalidateFull()
		n, err := r.Read(p.Data[pageAddr:])
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		addr += uint64(n)
	}
}

type memReader struct {
	m     *Memory
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

	pageIndex := r.addr >> PageAddrSize
	start := r.addr & PageAddrMask
	end := uint64(PageSize)

	if pageIndex == (endAddr >> PageAddrSize) {
		end = endAddr & PageAddrMask
	}
	p, ok := r.m.pageLookup(pageIndex)
	if ok {
		n = copy(dest, p.Data[start:end])
	} else {
		n = copy(dest, make([]byte, end-start)) // default to zeroes
	}
	r.addr += uint64(n)
	r.count -= uint64(n)
	return n, nil
}

func (m *Memory) ReadMemoryRange(addr uint64, count uint64) io.Reader {
	return &memReader{m: m, addr: addr, count: count}
}

func (m *Memory) Usage() string {
	total := uint64(len(m.pages)) * PageSize
	const unit = 1024
	if total < unit {
		return fmt.Sprintf("%d B", total)
	}
	div, exp := uint64(unit), 0
	for n := total / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// KiB, MiB, GiB, TiB, ...
	return fmt.Sprintf("%.1f %ciB", float64(total)/float64(div), "KMGTPE"[exp])
}
