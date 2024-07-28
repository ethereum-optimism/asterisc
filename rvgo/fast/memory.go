package fast

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
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

type Memory struct {
	// generalized index -> merkle root or nil if invalidated
	// pageIndex -> cached page

	pages map[uint64]*CachedPage

	radix         *RadixNodeLevel1
	branchFactors [5]uint64

	// Note: since we don't de-alloc pages, we don't do ref-counting.
	// Once a page exists, it doesn't leave memory

	// two caches: we often read instructions from one page, and do memory things with another page.
	// this prevents map lookups each instruction
	lastPageKeys [2]uint64
	lastPage     [2]*CachedPage
}

func NewMemory() *Memory {
	return &Memory{
		//nodes:         make(map[uint64]*[32]byte),
		radix:         &RadixNodeLevel1{},
		pages:         make(map[uint64]*CachedPage),
		branchFactors: [5]uint64{BF1, BF2, BF3, BF4, BF5},
		lastPageKeys:  [2]uint64{^uint64(0), ^uint64(0)}, // default to invalid keys, to not match any pages
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
		copy(dest[d:], zeroes[:l])
	} else {
		copy(dest[d:], p.Data[pageAddr:])
	}
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
	//m.nodes = make(map[uint64]*[32]byte)
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

// Serialize writes the memory in a simple binary format which can be read again using Deserialize
// The format is a simple concatenation of fields, with prefixed item count for repeating items and using big endian
// encoding for numbers.
//
// len(PageCount)    uint64
// For each page (order is arbitrary):
//
//	page index          uint64
//	page Data           [PageSize]byte
func (m *Memory) Serialize(out io.Writer) error {
	if err := binary.Write(out, binary.BigEndian, uint64(m.PageCount())); err != nil {
		return err
	}
	for pageIndex, page := range m.pages {
		if err := binary.Write(out, binary.BigEndian, pageIndex); err != nil {
			return err
		}
		if _, err := out.Write(page.Data[:]); err != nil {
			return err
		}
	}
	return nil
}

func (m *Memory) Deserialize(in io.Reader) error {
	var pageCount uint64
	if err := binary.Read(in, binary.BigEndian, &pageCount); err != nil {
		return err
	}
	for i := uint64(0); i < pageCount; i++ {
		var pageIndex uint64
		if err := binary.Read(in, binary.BigEndian, &pageIndex); err != nil {
			return err
		}
		page := m.AllocPage(pageIndex)
		if _, err := io.ReadFull(in, page.Data[:]); err != nil {
			return err
		}
	}
	return nil
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
