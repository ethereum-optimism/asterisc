package fast

import (
	"math/bits"
)

const (
	// Define branching factors for each level
	BF1 = 10
	BF2 = 10
	BF3 = 10
	BF4 = 10
	BF5 = 12
)

type RadixNodeLevel1 struct {
	Children   [1 << BF1]*RadixNodeLevel2
	Hashes     [1 << BF1][32]byte
	HashExists [(1 << BF1) / 64]uint64
	HashValid  [(1 << BF1) / 64]uint64
}

type RadixNodeLevel2 struct {
	Children   [1 << BF2]*RadixNodeLevel3
	Hashes     [1 << BF2][32]byte
	HashExists [(1 << BF2) / 64]uint64
	HashValid  [(1 << BF2) / 64]uint64
}

type RadixNodeLevel3 struct {
	Children   [1 << BF3]*RadixNodeLevel4
	Hashes     [1 << BF3][32]byte
	HashExists [(1 << BF3) / 64]uint64
	HashValid  [(1 << BF3) / 64]uint64
}

type RadixNodeLevel4 struct {
	Children   [1 << BF4]*RadixNodeLevel5
	Hashes     [1 << BF4][32]byte
	HashExists [(1 << BF4) / 64]uint64
	HashValid  [(1 << BF4) / 64]uint64
}

type RadixNodeLevel5 struct {
	Hashes     [1 << BF5][32]byte
	HashExists [(1 << BF5) / 64]uint64
	HashValid  [(1 << BF5) / 64]uint64
}

func (n *RadixNodeLevel1) invalidateHashes(branch uint64) {
	branch = (branch + 1<<BF1) / 2
	for index := branch; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)
	}
}
func (n *RadixNodeLevel2) invalidateHashes(branch uint64) {
	branch = (branch + 1<<BF2) / 2
	for index := branch; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)
	}
}
func (n *RadixNodeLevel3) invalidateHashes(branch uint64) {
	branch = (branch + 1<<BF3) / 2
	for index := branch; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)

	}
}
func (n *RadixNodeLevel4) invalidateHashes(branch uint64) {
	branch = (branch + 1<<BF4) / 2
	for index := branch; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)

	}
}

func (n *RadixNodeLevel5) invalidateHashes(branch uint64) {
	branch = (branch + 1<<BF5) / 2
	for index := branch; index > 0; index >>= 1 {
		hashIndex := index >> 6
		hashBit := index & 63
		n.HashExists[hashIndex] |= 1 << hashBit
		n.HashValid[hashIndex] &= ^(1 << hashBit)

	}
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

	branchPaths := m.addressToBranchPath(addr)

	currentLevel1 := m.radix

	currentLevel1.invalidateHashes(branchPaths[0])
	if currentLevel1.Children[branchPaths[0]] == nil {
		return
	}

	currentLevel2 := currentLevel1.Children[branchPaths[0]]
	currentLevel2.invalidateHashes(branchPaths[1])
	if currentLevel2.Children[branchPaths[1]] == nil {
		return
	}

	currentLevel3 := currentLevel2.Children[branchPaths[1]]
	currentLevel3.invalidateHashes(branchPaths[2])
	if currentLevel3.Children[branchPaths[2]] == nil {
		return
	}

	currentLevel4 := currentLevel3.Children[branchPaths[2]]
	currentLevel4.invalidateHashes(branchPaths[3])
	if currentLevel4.Children[branchPaths[3]] == nil {
		return
	}

	currentLevel5 := currentLevel4.Children[branchPaths[3]]
	currentLevel5.invalidateHashes(branchPaths[4])
}

func (m *Memory) MerkleizeNodeLevel1(node *RadixNodeLevel1, addr, gindex uint64) [32]byte {
	depth := uint64(bits.Len64(gindex))

	if depth <= BF1 {
		hashIndex := gindex >> 6
		hashBit := gindex & 63

		if (node.HashExists[hashIndex] & (1 << hashBit)) != 0 {
			if (node.HashValid[hashIndex] & (1 << hashBit)) != 0 {
				return node.Hashes[gindex]
			} else {
				left := m.MerkleizeNodeLevel1(node, addr, gindex<<1)
				right := m.MerkleizeNodeLevel1(node, addr, (gindex<<1)|1)

				r := HashPair(left, right)
				node.Hashes[gindex] = r
				node.HashValid[hashIndex] |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-depth]
		}
	}

	if depth > BF1<<1 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<BF1
	if node.Children[childIndex] == nil {
		return zeroHashes[64-5+1-depth]
	}
	addr <<= BF1
	addr |= childIndex
	return m.MerkleizeNodeLevel2(node.Children[childIndex], addr, 1)
}

func (m *Memory) MerkleizeNodeLevel2(node *RadixNodeLevel2, addr, gindex uint64) [32]byte {

	depth := uint64(bits.Len64(gindex))

	if depth <= BF2 {
		hashIndex := gindex >> 6
		hashBit := gindex & 63

		if (node.HashExists[hashIndex] & (1 << hashBit)) != 0 {
			if (node.HashValid[hashIndex] & (1 << hashBit)) != 0 {
				return node.Hashes[gindex]
			} else {
				left := m.MerkleizeNodeLevel2(node, addr, gindex<<1)
				right := m.MerkleizeNodeLevel2(node, addr, (gindex<<1)|1)

				r := HashPair(left, right)
				node.Hashes[gindex] = r
				node.HashValid[hashIndex] |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-(depth+BF1)]
		}
	}

	if depth > BF2<<1 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<BF2
	if node.Children[childIndex] == nil {
		return zeroHashes[64-5+1-(depth+BF1)]
	}

	addr <<= BF2
	addr |= childIndex
	return m.MerkleizeNodeLevel3(node.Children[childIndex], addr, 1)
}
func (m *Memory) MerkleizeNodeLevel3(node *RadixNodeLevel3, addr, gindex uint64) [32]byte {

	depth := uint64(bits.Len64(gindex))

	if depth <= BF3 {
		hashIndex := gindex >> 6
		hashBit := gindex & 63

		if (node.HashExists[hashIndex] & (1 << hashBit)) != 0 {
			if (node.HashValid[hashIndex] & (1 << hashBit)) != 0 {
				return node.Hashes[gindex]
			} else {
				left := m.MerkleizeNodeLevel3(node, addr, gindex<<1)
				right := m.MerkleizeNodeLevel3(node, addr, (gindex<<1)|1)
				r := HashPair(left, right)
				node.Hashes[gindex] = r
				node.HashValid[hashIndex] |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-(depth+BF1+BF2)]
		}
	}

	if depth > BF3<<1 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<BF3
	if node.Children[childIndex] == nil {
		return zeroHashes[64-5+1-(depth+BF1+BF2)]
	}

	addr <<= BF3
	addr |= childIndex
	return m.MerkleizeNodeLevel4(node.Children[childIndex], addr, 1)
}

func (m *Memory) MerkleizeNodeLevel4(node *RadixNodeLevel4, addr, gindex uint64) [32]byte {

	depth := uint64(bits.Len64(gindex))

	if depth <= BF4 {
		hashIndex := gindex >> 6
		hashBit := gindex & 63
		if (node.HashExists[hashIndex] & (1 << hashBit)) != 0 {
			if (node.HashValid[hashIndex] & (1 << hashBit)) != 0 {
				return node.Hashes[gindex]
			} else {
				left := m.MerkleizeNodeLevel4(node, addr, gindex<<1)
				right := m.MerkleizeNodeLevel4(node, addr, (gindex<<1)|1)

				r := HashPair(left, right)
				node.Hashes[gindex] = r
				node.HashValid[hashIndex] |= 1 << hashBit
				return r
			}
		} else {
			return zeroHashes[64-5+1-(depth+BF1+BF2+BF3)]
		}
	}

	if depth > BF4<<1 {
		panic("gindex too deep")
	}

	childIndex := gindex - 1<<BF4
	if node.Children[childIndex] == nil {
		return zeroHashes[64-5+1-(depth+BF1+BF2+BF3)]
	}

	addr <<= BF4
	addr |= childIndex
	return m.MerkleizeNodeLevel5(node.Children[childIndex], addr, 1)
}

func (m *Memory) MerkleizeNodeLevel5(node *RadixNodeLevel5, addr, gindex uint64) [32]byte {
	depth := uint64(bits.Len64(gindex))

	if depth > BF5 {
		pageIndex := (addr << BF5) | (gindex - (1 << BF5))
		if p, ok := m.pages[pageIndex]; ok {
			return p.MerkleRoot()
		} else {
			return zeroHashes[64-5+1-(depth+40)]
		}
	}

	hashIndex := gindex >> 6
	hashBit := gindex & 63

	if (node.HashExists[hashIndex] & (1 << hashBit)) != 0 {
		if (node.HashValid[hashIndex] & (1 << hashBit)) != 0 {
			return node.Hashes[gindex]
		} else {
			left := m.MerkleizeNodeLevel5(node, addr, gindex<<1)
			right := m.MerkleizeNodeLevel5(node, addr, (gindex<<1)|1)
			r := HashPair(left, right)
			node.Hashes[gindex] = r
			node.HashValid[hashIndex] |= 1 << hashBit
			return r
		}
	} else {
		return zeroHashes[64-5+1-(depth+40)]
	}
}

func (m *Memory) GenerateProof1(node *RadixNodeLevel1, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF1; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel1(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof2(node *RadixNodeLevel2, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF2; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel2(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof3(node *RadixNodeLevel3, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF3; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel3(node, addr, sibling))
	}

	return proofs
}
func (m *Memory) GenerateProof4(node *RadixNodeLevel4, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF4; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel4(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof5(node *RadixNodeLevel5, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF5; idx > 1; idx >>= 1 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel5(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) MerkleProof(addr uint64) [ProofLen * 32]byte {
	var proofs [60][32]byte

	branchPaths := m.addressToBranchPath(addr)

	// Level 1
	proofIndex := BF1
	currentLevel1 := m.radix
	branch1 := branchPaths[0]

	levelProofs := m.GenerateProof1(currentLevel1, 0, branch1)
	copy(proofs[60-proofIndex:60], levelProofs)

	// Level 2
	currentLevel2 := m.radix.Children[branchPaths[0]]
	if currentLevel2 != nil {
		branch2 := branchPaths[1]
		proofIndex += BF2
		levelProofs := m.GenerateProof2(currentLevel2, addr>>(PageAddrSize+BF5+BF4+BF3+BF2), branch2)
		copy(proofs[60-proofIndex:60-proofIndex+BF2], levelProofs)
	} else {
		fillZeroHashes(proofs[:], 0, 60-proofIndex)
		return encodeProofs(proofs)
	}

	// Level 3
	currentLevel3 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]]
	if currentLevel3 != nil {
		branch3 := branchPaths[2]
		proofIndex += BF3
		levelProofs := m.GenerateProof3(currentLevel3, addr>>(PageAddrSize+BF5+BF4+BF3), branch3)
		copy(proofs[60-proofIndex:60-proofIndex+BF3], levelProofs)
	} else {
		fillZeroHashes(proofs[:], 0, 60-proofIndex)
		return encodeProofs(proofs)
	}

	// Level 4
	currentLevel4 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]].Children[branchPaths[2]]
	if currentLevel4 != nil {
		branch4 := branchPaths[3]
		levelProofs := m.GenerateProof4(currentLevel4, addr>>(PageAddrSize+BF5+BF4), branch4)
		proofIndex += BF4
		copy(proofs[60-proofIndex:60-proofIndex+BF4], levelProofs)
	} else {
		fillZeroHashes(proofs[:], 0, 60-proofIndex)
		return encodeProofs(proofs)
	}

	// Level 5
	currentLevel5 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]].Children[branchPaths[2]].Children[branchPaths[3]]
	if currentLevel5 != nil {
		branch5 := branchPaths[4]
		levelProofs := m.GenerateProof5(currentLevel5, addr>>(PageAddrSize+BF5), branch5)
		proofIndex += BF5
		copy(proofs[60-proofIndex:60-proofIndex+BF5], levelProofs)
	} else {
		fillZeroHashes(proofs[:], 0, 60-proofIndex)
		return encodeProofs(proofs)
	}

	// Page-level proof
	pageGindex := PageSize>>5 + (addr&PageAddrMask)>>5
	pageIndex := addr >> PageAddrSize

	proofIndex = 0
	if p, ok := m.pages[pageIndex]; ok {
		proofs[proofIndex] = p.MerkleizeSubtree(pageGindex)
		for idx := pageGindex; idx > 1; idx >>= 1 {
			sibling := idx ^ 1
			proofIndex++
			proofs[proofIndex] = p.MerkleizeSubtree(uint64(sibling))
		}
	} else {
		fillZeroHashes(proofs[:], 0, 7)
	}

	return encodeProofs(proofs)
}

func fillZeroHashes(proofs [][32]byte, start, end int) {
	if start == 0 {
		proofs[0] = zeroHashes[0]
		start++
	}
	for i := start; i <= end; i++ {
		proofs[i] = zeroHashes[i-1]
	}
}

func encodeProofs(proofs [60][32]byte) [ProofLen * 32]byte {
	var out [ProofLen * 32]byte
	for i := 0; i < ProofLen; i++ {
		copy(out[i*32:(i+1)*32], proofs[i][:])
	}
	return out
}

func (m *Memory) MerkleRoot() [32]byte {
	return m.MerkleizeNodeLevel1(m.radix, 0, 1)
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
	if currentLevel1.Children[branch1] == nil {
		node := &RadixNodeLevel2{}
		currentLevel1.Children[branch1] = node

	}
	currentLevel1.invalidateHashes(branchPaths[0])
	currentLevel2 := currentLevel1.Children[branch1]

	branch2 := branchPaths[1]
	if currentLevel2.Children[branch2] == nil {
		node := &RadixNodeLevel3{}
		currentLevel2.Children[branch2] = node
	}
	currentLevel2.invalidateHashes(branchPaths[1])
	currentLevel3 := currentLevel2.Children[branch2]

	branch3 := branchPaths[2]
	if currentLevel3.Children[branch3] == nil {
		node := &RadixNodeLevel4{}
		currentLevel3.Children[branch3] = node
	}
	currentLevel3.invalidateHashes(branchPaths[2])
	currentLevel4 := currentLevel3.Children[branch3]

	branch4 := branchPaths[3]
	if currentLevel4.Children[branch4] == nil {
		node := &RadixNodeLevel5{}
		currentLevel4.Children[branch4] = node
	}
	currentLevel4.invalidateHashes(branchPaths[3])

	currentLevel5 := currentLevel4.Children[branchPaths[3]]
	currentLevel5.invalidateHashes(branchPaths[4])

	// For Level 5, we don't need to allocate a child node

	return p
}
