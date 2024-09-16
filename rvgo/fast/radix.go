package fast

import (
	"math/bits"
)

type RadixNode interface {
	InvalidateNode(addr uint64)
	GenerateProof(addr uint64) [][32]byte
	MerkleizeNode(addr, gindex uint64) [32]byte
}

type SmallRadixNode[C RadixNode] struct {
	Children   [1 << 4]*C
	Hashes     [1 << 4][32]byte
	HashExists uint16
	HashValid  uint16
	Depth      uint16
}

type LargeRadixNode[C RadixNode] struct {
	Children   [1 << 8]*C
	Hashes     [1 << 8][32]byte
	HashExists [(1 << 8) / 64]uint64
	HashValid  [(1 << 8) / 64]uint64
	Depth      uint16
}

type L1 = SmallRadixNode[L2]
type L2 = *SmallRadixNode[L3]
type L3 = *SmallRadixNode[L4]
type L4 = *SmallRadixNode[L5]
type L5 = *SmallRadixNode[L6]
type L6 = *SmallRadixNode[L7]
type L7 = *SmallRadixNode[L8]
type L8 = *LargeRadixNode[L9]
type L9 = *LargeRadixNode[L10]
type L10 = *LargeRadixNode[L11]
type L11 = *Memory

func (n *SmallRadixNode[C]) InvalidateNode(addr uint64) {
	childIdx := addressToRadixPath(addr, n.Depth, 4)

	branchIdx := (childIdx + 1<<4) / 2
	for index := branchIdx; index > 0; index >>= 1 {
		hashBit := index & 15
		n.HashExists |= 1 << hashBit
		n.HashValid &= ^(1 << hashBit)
	}

	if n.Children[childIdx] != nil {
		(*n.Children[childIdx]).InvalidateNode(addr)
	}
}

func (n *LargeRadixNode[C]) InvalidateNode(addr uint64) {
	childIdx := addressToRadixPath(addr, n.Depth, 8)

	branchIdx := (childIdx + 1<<8) / 2

	for index := branchIdx; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)
	}

	if n.Children[childIdx] != nil {
		(*n.Children[childIdx]).InvalidateNode(addr)
	}
}

func (m *Memory) InvalidateNode(addr uint64) {
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
}

func (n *SmallRadixNode[C]) GenerateProof(addr uint64) [][32]byte {
	var proofs [][32]byte
	path := addressToRadixPath(addr, n.Depth, 4)

	if n.Children[path] == nil {
		proofs = zeroHashRange(0, 60-n.Depth-4)
	} else {
		proofs = (*n.Children[path]).GenerateProof(addr)
	}
	for idx := path + 1<<4; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, n.MerkleizeNode(addr>>(64-n.Depth), sibling))
	}

	return proofs
}

func (n *LargeRadixNode[C]) GenerateProof(addr uint64) [][32]byte {
	var proofs [][32]byte
	path := addressToRadixPath(addr, n.Depth, 8)

	if n.Children[path] == nil {
		proofs = zeroHashRange(0, 60-n.Depth-8)
	} else {
		proofs = (*n.Children[path]).GenerateProof(addr)
	}

	for idx := path + 1<<8; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, n.MerkleizeNode(addr>>(64-n.Depth), sibling))
	}
	return proofs
}

func (m *Memory) GenerateProof(addr uint64) [][32]byte {
	pageIndex := addr >> PageAddrSize

	if p, ok := m.pages[pageIndex]; ok {
		return p.GenerateProof(addr)
	} else {
		return zeroHashRange(0, 8)
	}
}

func (n *SmallRadixNode[C]) MerkleizeNode(addr, gindex uint64) [32]byte {
	depth := uint16(bits.Len64(gindex))

	if depth <= 4 {
		hashBit := gindex & 15

		if (n.HashExists & (1 << hashBit)) != 0 {
			if (n.HashValid & (1 << hashBit)) != 0 {
				return n.Hashes[gindex]
			} else {
				left := n.MerkleizeNode(addr, gindex<<1)
				right := n.MerkleizeNode(addr, (gindex<<1)|1)

				r := HashPair(left, right)
				n.Hashes[gindex] = r
				n.HashValid |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-(depth+n.Depth)]
		}
	}

	if depth > 5 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<4
	if n.Children[childIndex] == nil {
		return zeroHashes[64-5+1-(depth+n.Depth)]
	}
	addr <<= 4
	addr |= childIndex
	return (*n.Children[childIndex]).MerkleizeNode(addr, 1)
}

func (n *LargeRadixNode[C]) MerkleizeNode(addr, gindex uint64) [32]byte {
	depth := uint16(bits.Len64(gindex))

	if depth <= 8 {
		hashIndex := gindex >> 6
		hashBit := gindex & 63
		if (n.HashExists[hashIndex] & (1 << hashBit)) != 0 {
			if (n.HashValid[hashIndex] & (1 << hashBit)) != 0 {
				return n.Hashes[gindex]
			} else {
				left := n.MerkleizeNode(addr, gindex<<1)
				right := n.MerkleizeNode(addr, (gindex<<1)|1)

				r := HashPair(left, right)
				n.Hashes[gindex] = r
				n.HashValid[hashIndex] |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-(depth+n.Depth)]
		}
	}

	if depth > 8<<1 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<8
	if n.Children[int(childIndex)] == nil {
		return zeroHashes[64-5+1-(depth+n.Depth)]
	}

	addr <<= 8
	addr |= childIndex
	return (*n.Children[childIndex]).MerkleizeNode(addr, 1)
}

func (m *Memory) MerkleizeNode(addr, gindex uint64) [32]byte {
	depth := uint64(bits.Len64(gindex))

	pageIndex := addr
	if p, ok := m.pages[pageIndex]; ok {
		return p.MerkleRoot()
	} else {
		return zeroHashes[64-5+1-(depth-1+52)]
	}
}

func (m *Memory) MerkleRoot() [32]byte {
	return (*m.radix).MerkleizeNode(0, 1)
}

func (m *Memory) MerkleProof(addr uint64) [ProofLen * 32]byte {
	proofs := m.radix.GenerateProof(addr)

	return encodeProofs(proofs)
}

func zeroHashRange(start, end uint16) [][32]byte {
	proofs := make([][32]byte, end-start)
	if start == 0 {
		proofs[0] = zeroHashes[0]
		start++
	}
	for i := start; i < end; i++ {
		proofs[i] = zeroHashes[i-1]
	}
	return proofs
}

func encodeProofs(proofs [][32]byte) [ProofLen * 32]byte {
	var out [ProofLen * 32]byte
	for i := 0; i < ProofLen; i++ {
		copy(out[i*32:(i+1)*32], proofs[i][:])
	}
	return out
}

func addressToRadixPath(addr uint64, position, count uint16) uint64 {
	// Calculate the total shift amount
	totalShift := PageAddrSize + 52 - position - count

	// Shift the address to bring the desired bits to the LSB
	addr >>= totalShift

	// Extract the desired bits using a mask
	return addr & ((1 << count) - 1)
}

func (m *Memory) addressToBranchPath(addr uint64) []uint64 {
	addr >>= PageAddrSize

	path := make([]uint64, len(m.branchFactors))
	for i := len(m.branchFactors) - 1; i >= 0; i-- {
		bits := m.branchFactors[i]
		mask := (1 << bits) - 1       // Create a mask for the current segment
		path[i] = addr & uint64(mask) // Extract the segment using the mask
		addr >>= bits                 // Shift the gindex to the right by the number of bits processed
	}
	return path
}

func (m *Memory) AllocPage(pageIndex uint64) *CachedPage {
	p := &CachedPage{Data: new(Page)}
	m.pages[pageIndex] = p

	branchPaths := m.addressToBranchPath(pageIndex << PageAddrSize)
	currentLevel1 := m.radix
	branch1 := branchPaths[0]
	if (*currentLevel1).Children[branch1] == nil {
		node := &SmallRadixNode[L3]{Depth: 4}
		(*currentLevel1).Children[branch1] = &node
	}
	currentLevel2 := (*currentLevel1).Children[branch1]

	branch2 := branchPaths[1]
	if (*currentLevel2).Children[branch2] == nil {
		node := &SmallRadixNode[L4]{Depth: 8}
		(*currentLevel2).Children[branch2] = &node
	}
	currentLevel3 := (*currentLevel2).Children[branch2]

	branch3 := branchPaths[2]
	if (*currentLevel3).Children[branch3] == nil {
		node := &SmallRadixNode[L5]{Depth: 12}
		(*currentLevel3).Children[branch3] = &node
	}
	currentLevel4 := (*currentLevel3).Children[branch3]

	branch4 := branchPaths[3]
	if (*currentLevel4).Children[branch4] == nil {
		node := &SmallRadixNode[L6]{Depth: 16}
		(*currentLevel4).Children[branch4] = &node
	}
	currentLevel5 := (*currentLevel4).Children[branch4]

	branch5 := branchPaths[4]
	if (*currentLevel5).Children[branch5] == nil {
		node := &SmallRadixNode[L7]{Depth: 20}
		(*currentLevel5).Children[branch5] = &node
	}
	currentLevel6 := (*currentLevel5).Children[branch5]

	branch6 := branchPaths[5]
	if (*currentLevel6).Children[branch6] == nil {
		node := &SmallRadixNode[L8]{Depth: 24}
		(*currentLevel6).Children[branch6] = &node
	}
	currentLevel7 := (*currentLevel6).Children[branch6]

	branch7 := branchPaths[6]
	if (*currentLevel7).Children[branch7] == nil {
		node := &LargeRadixNode[L9]{Depth: 28}
		(*currentLevel7).Children[branch7] = &node
	}
	currentLevel8 := (*currentLevel7).Children[branch7]

	branch8 := branchPaths[7]
	if (*currentLevel8).Children[branch8] == nil {
		node := &LargeRadixNode[L10]{Depth: 36}
		(*currentLevel8).Children[branch8] = &node
	}
	currentLevel9 := (*currentLevel8).Children[branch8]

	branch9 := branchPaths[8]
	if (*currentLevel9).Children[branch9] == nil {
		node := &LargeRadixNode[L11]{Depth: 44}
		(*currentLevel9).Children[branch9] = &node
	}
	currentLevel10 := (*currentLevel9).Children[branch9]

	branch10 := branchPaths[9]

	(*currentLevel10).Children[branch10] = &m

	m.Invalidate(pageIndex << PageAddrSize)

	return p
}

func (m *Memory) Invalidate(addr uint64) {
	m.radix.InvalidateNode(addr)
}
