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

type RadixNode interface {
	merkleize(m *Memory, addr, gindex uint64) [32]byte
	//getChild(index uint64) RadixNode
	//setChild(index uint64, child RadixNode)
	invalidateHashes(branch uint64)
}

//func (n *baseRadixNode) invalidateHashes(branch uint64) {
//	for index := branch + (1 << 10); index > 0; index /= 2 {
//		n.HashCache[index] = false
//		n.Hashes[index] = [32]byte{}
//	}
//}

type RadixNodeLevel1 struct {
	Children  [1 << BF1]*RadixNodeLevel2
	Hashes    [2 * 1 << BF1][32]byte
	HashCache [2 * 1 << BF1]bool
}

type RadixNodeLevel2 struct {
	Children  [1 << BF2]*RadixNodeLevel3
	Hashes    [2 * 1 << BF2][32]byte
	HashCache [2 * 1 << BF2]bool
}

type RadixNodeLevel3 struct {
	Children  [1 << BF3]*RadixNodeLevel4
	Hashes    [2 * 1 << BF3][32]byte
	HashCache [2 * 1 << BF3]bool
}

type RadixNodeLevel4 struct {
	Children  [1 << BF4]*RadixNodeLevel5
	Hashes    [2 * 1 << BF4][32]byte
	HashCache [2 * 1 << BF4]bool
}

type RadixNodeLevel5 struct {
	Hashes    [2 * 1 << BF5][32]byte
	HashCache [2 * 1 << BF5]bool
}

func (n *RadixNodeLevel1) invalidateHashes(branch uint64) {
	for index := branch + (1 << BF1); index > 0; index /= 2 {
		n.HashCache[index] = false
		n.Hashes[index] = [32]byte{}
	}
}
func (n *RadixNodeLevel2) invalidateHashes(branch uint64) {
	for index := branch + (1 << BF2); index > 0; index /= 2 {
		n.HashCache[index] = false
		n.Hashes[index] = [32]byte{}
	}
}
func (n *RadixNodeLevel3) invalidateHashes(branch uint64) {
	for index := branch + (1 << BF3); index > 0; index /= 2 {
		n.HashCache[index] = false
		n.Hashes[index] = [32]byte{}
	}
}
func (n *RadixNodeLevel4) invalidateHashes(branch uint64) {
	for index := branch + (1 << BF4); index > 0; index /= 2 {
		n.HashCache[index] = false
		n.Hashes[index] = [32]byte{}
	}
}

func (n *RadixNodeLevel5) invalidateHashes(branch uint64) {
	for index := branch + (1 << BF5); index > 0; index /= 2 {
		n.HashCache[index] = false
		n.Hashes[index] = [32]byte{}
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
	if gindex > 2*1<<BF1-1 {
		return [32]byte{}
	}

	depth := uint64(bits.Len64(gindex))

	if node.HashCache[gindex] {
		if node.Hashes[gindex] == [32]byte{} {
			return zeroHashes[64-5+1-depth]
		} else {
			return node.Hashes[gindex]
		}
	}

	if gindex < 1<<BF1 {
		left := m.MerkleizeNodeLevel1(node, addr, gindex<<1)
		right := m.MerkleizeNodeLevel1(node, addr, (gindex<<1)|1)

		r := HashPair(left, right)
		node.Hashes[gindex] = r
		node.HashCache[gindex] = true
		return r
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
	if gindex > 2*1<<BF2 {
		return [32]byte{}
	}

	depth := uint64(bits.Len64(gindex))

	if node.HashCache[gindex] {
		if node.Hashes[gindex] == [32]byte{} {
			return zeroHashes[64-5+1-depth]
		} else {
			return node.Hashes[gindex]
		}
	}

	if gindex < 1<<BF2 {
		left := m.MerkleizeNodeLevel2(node, addr, gindex<<1)
		right := m.MerkleizeNodeLevel2(node, addr, (gindex<<1)|1)

		r := HashPair(left, right)
		node.Hashes[gindex] = r
		node.HashCache[gindex] = true
		return r
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
	if gindex > 2*1<<BF3 {
		return [32]byte{}
	}

	depth := uint64(bits.Len64(gindex))

	if node.HashCache[gindex] {
		if node.Hashes[gindex] == [32]byte{} {
			return zeroHashes[64-5+1-depth]
		} else {
			return node.Hashes[gindex]
		}
	}

	if gindex < 1<<BF3 {
		left := m.MerkleizeNodeLevel3(node, addr, gindex<<1)
		right := m.MerkleizeNodeLevel3(node, addr, (gindex<<1)|1)
		r := HashPair(left, right)
		node.Hashes[gindex] = r
		node.HashCache[gindex] = true
		return r
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
	if gindex > 2*1<<BF4 {
		return [32]byte{}
	}

	depth := uint64(bits.Len64(gindex))

	if node.HashCache[gindex] {
		if node.Hashes[gindex] == [32]byte{} {
			return zeroHashes[64-5+1-depth]
		} else {
			return node.Hashes[gindex]
		}
	}

	if gindex < 1<<BF4 {
		left := m.MerkleizeNodeLevel4(node, addr, gindex<<1)
		right := m.MerkleizeNodeLevel4(node, addr, (gindex<<1)|1)

		r := HashPair(left, right)
		node.Hashes[gindex] = r
		node.HashCache[gindex] = true
		return r
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

	if gindex >= (1 << BF5) {
		pageIndex := (addr << BF5) | (gindex - (1 << BF5))
		if p, ok := m.pages[pageIndex]; ok {
			return p.MerkleRoot()
		} else {
			return zeroHashes[64-5+1-(depth+40)]
		}
	}

	if node.HashCache[gindex] {
		if node.Hashes[gindex] == [32]byte{} {
			return zeroHashes[64-5+1-depth]
		} else {
			return node.Hashes[gindex]
		}
	}

	left := m.MerkleizeNodeLevel5(node, addr, gindex<<1)
	right := m.MerkleizeNodeLevel5(node, addr, (gindex<<1)|1)
	r := HashPair(left, right)
	node.Hashes[gindex] = r
	node.HashCache[gindex] = true
	return r

}

func (m *Memory) GenerateProof1(node *RadixNodeLevel1, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF1; idx > 1; idx /= 2 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel1(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof2(node *RadixNodeLevel2, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF2; idx > 1; idx /= 2 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel2(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof3(node *RadixNodeLevel3, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF3; idx > 1; idx /= 2 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel3(node, addr, sibling))
	}

	return proofs
}
func (m *Memory) GenerateProof4(node *RadixNodeLevel4, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF4; idx > 1; idx /= 2 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel4(node, addr, sibling))
	}

	return proofs
}

func (m *Memory) GenerateProof5(node *RadixNodeLevel5, addr, target uint64) [][32]byte {
	var proofs [][32]byte

	for idx := target + 1<<BF5; idx > 1; idx /= 2 {
		sibling := idx ^ 1
		proofs = append(proofs, m.MerkleizeNodeLevel5(node, addr, sibling))
	}

	return proofs
}
func (m *Memory) MerkleProof(addr uint64) [ProofLen * 32]byte {
	var proofs [60][32]byte
	proofIndex := 0 // Start from the beginning, as we're building the proof from page to root

	branchPaths := m.addressToBranchPath(addr)

	// Page-level proof
	pageGindex := PageSize>>5 + (addr&PageAddrMask)>>5
	pageIndex := addr >> PageAddrSize

	if p, ok := m.pages[pageIndex]; ok {
		proofs[proofIndex] = p.MerkleizeSubtree(pageGindex)
		proofIndex++
		for idx := pageGindex; idx > 1; idx /= 2 {
			sibling := idx ^ 1
			proofs[proofIndex] = p.MerkleizeSubtree(uint64(sibling))
			proofIndex++
		}
	} else {
		fillZeroHashes(proofs[:], proofIndex, proofIndex+7, 12)
		proofIndex += 8
	}

	// Level 5
	currentLevel5 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]].Children[branchPaths[2]].Children[branchPaths[3]]
	if currentLevel5 != nil {
		branch5 := branchPaths[4]
		levelProofs := m.GenerateProof5(currentLevel5, addr>>(pageKeySize-BF1-BF2-BF3-BF4), branch5)
		copy(proofs[proofIndex:proofIndex+12], levelProofs)
		proofIndex += 12
	} else {
		fillZeroHashes(proofs[:], proofIndex, proofIndex+9, 22)
		return encodeProofs(proofs)
	}

	// Level 4
	currentLevel4 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]].Children[branchPaths[2]]
	if currentLevel4 != nil {
		branch4 := branchPaths[3]
		levelProofs := m.GenerateProof4(currentLevel4, addr>>(pageKeySize-BF1-BF2-BF3), branch4)
		copy(proofs[proofIndex:proofIndex+10], levelProofs)
		proofIndex += 10
	} else {
		fillZeroHashes(proofs[:], proofIndex, proofIndex+9, 32)
		return encodeProofs(proofs)
	}

	// Level 3
	currentLevel3 := m.radix.Children[branchPaths[0]].Children[branchPaths[1]]
	if currentLevel3 != nil {
		branch3 := branchPaths[2]
		levelProofs := m.GenerateProof3(currentLevel3, addr>>(pageKeySize-BF1-BF2), branch3)
		copy(proofs[proofIndex:proofIndex+10], levelProofs)
		proofIndex += 10
	} else {
		fillZeroHashes(proofs[:], proofIndex, proofIndex+9, 42)
		return encodeProofs(proofs)
	}

	// Level 2
	currentLevel2 := m.radix.Children[branchPaths[0]]
	if currentLevel2 != nil {
		branch2 := branchPaths[1]
		levelProofs := m.GenerateProof2(currentLevel2, addr>>(pageKeySize-BF1), branch2)
		copy(proofs[proofIndex:proofIndex+10], levelProofs)
		proofIndex += 10
	} else {
		fillZeroHashes(proofs[:], proofIndex, proofIndex+9, 52)
		return encodeProofs(proofs)
	}

	// Level 1
	currentLevel1 := m.radix
	branch1 := branchPaths[0]
	levelProofs := m.GenerateProof1(currentLevel1, 0, branch1)
	copy(proofs[proofIndex:proofIndex+10], levelProofs)

	return encodeProofs(proofs)
}

func fillZeroHashes(proofs [][32]byte, start, end int, startingBitDepth int) {
	for i := start; i <= end; i++ {
		proofs[i] = zeroHashes[startingBitDepth-(i-start)]
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
		currentLevel1.Children[branch1] = &RadixNodeLevel2{}
	}
	currentLevel2 := currentLevel1.Children[branch1]

	branch2 := branchPaths[1]
	if currentLevel2.Children[branch2] == nil {
		currentLevel2.Children[branch2] = &RadixNodeLevel3{}
	}
	currentLevel3 := currentLevel2.Children[branch2]

	branch3 := branchPaths[2]
	if currentLevel3.Children[branch3] == nil {
		currentLevel3.Children[branch3] = &RadixNodeLevel4{}
	}
	currentLevel4 := currentLevel3.Children[branch3]

	branch4 := branchPaths[3]
	if currentLevel4.Children[branch4] == nil {
		currentLevel4.Children[branch4] = &RadixNodeLevel5{}
	}

	// For Level 5, we don't need to allocate a child node

	return p
}
