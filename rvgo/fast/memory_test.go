package fast

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMemoryMerkleProof tests the Merkle proof generation for memory trees.
func TestMemoryMerkleProof(t *testing.T) {
	// Test case for a nearly empty tree.
	t.Run("nearly empty tree", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set a value at an unaligned address.
		m.SetUnaligned(0x10000, []byte{0xaa, 0xbb, 0xcc, 0xdd})
		// Generate a Merkle proof for the set address.
		proof := m.MerkleProof(0x10000)
		// Verify the proof contains the expected value.
		require.Equal(t, uint32(0xaabbccdd), binary.BigEndian.Uint32(proof[:4]))
		// Check that the rest of the proof is filled with zero hashes.
		for i := 0; i < 32-5; i++ {
			require.Equal(t, zeroHashes[i][:], proof[32+i*32:32+i*32+32], "empty siblings")
		}
	})

	// Test case for a tree with more entries.
	t.Run("fuller tree", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set multiple values at unaligned addresses.
		m.SetUnaligned(0x10000, []byte{0xaa, 0xbb, 0xcc, 0xdd})
		m.SetUnaligned(0x80004, []byte{42})
		m.SetUnaligned(0x13370000, []byte{123})
		// Calculate the Merkle root of the tree.
		root := m.MerkleRoot()
		// Generate a Merkle proof for one of the addresses.
		proof := m.MerkleProof(0x80004)
		// Verify the proof contains the expected value for the set address.
		require.Equal(t, uint32(42<<24), binary.BigEndian.Uint32(proof[4:8]))
		// Reconstruct the node from the proof and verify it matches the root.
		node := *(*[32]byte)(proof[:32])
		path := uint32(0x80004) >> 5
		for i := 32; i < len(proof); i += 32 {
			sib := *(*[32]byte)(proof[i : i+32])
			if path&1 != 0 {
				node = HashPair(sib, node)
			} else {
				node = HashPair(node, sib)
			}
			path >>= 1
		}
		require.Equal(t, root, node, "proof must verify")
	})
}

// TestMemoryMerkleRoot tests the calculation of Merkle roots for memory trees.
func TestMemoryMerkleRoot(t *testing.T) {
	// Test case for an empty memory.
	t.Run("empty", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Calculate the Merkle root of the empty memory.
		root := m.MerkleRoot()
		// Verify the root matches the expected zero hash.
		require.Equal(t, zeroHashes[64-5], root, "fully zeroed memory should have expected zero hash")
	})

	// Test case for a memory with a single zeroed page.
	t.Run("empty page", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set a zero value at an unaligned address.
		m.SetUnaligned(0xF000, []byte{0})
		// Calculate the Merkle root of the memory.
		root := m.MerkleRoot()
		// Verify the root matches the expected zero hash.
		require.Equal(t, zeroHashes[64-5], root, "fully zeroed memory should have expected zero hash")
	})

	// Test case for a memory with a single non-zero page.
	t.Run("single page", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set a non-zero value at an unaligned address.
		m.SetUnaligned(0xF000, []byte{1})
		// Calculate the Merkle root of the memory.
		root := m.MerkleRoot()
		// Verify the root does not match the expected zero hash.
		require.NotEqual(t, zeroHashes[64-5], root, "non-zero memory")
	})

	// Test case for a memory with repeated zeroes.
	t.Run("repeat zero", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set zero values at two unaligned addresses.
		m.SetUnaligned(0xF000, []byte{0})
		m.SetUnaligned(0xF004, []byte{0})
		// Calculate the Merkle root of the memory.
		root := m.MerkleRoot()
		// Verify the root matches the expected zero hash.
		require.Equal(t, zeroHashes[64-5], root, "zero still")
	})

	// Test case for a memory with two empty pages.
	t.Run("two empty pages", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set zero values at two unaligned addresses.
		m.SetUnaligned(PageSize*3, []byte{0})
		m.SetUnaligned(PageSize*10, []byte{0})
		// Calculate the Merkle root of the memory.
		root := m.MerkleRoot()
		// Verify the root matches the expected zero hash.
		require.Equal(t, zeroHashes[64-5], root, "zero still")
	})

	// Test case for a memory with random few pages.
	t.Run("random few pages", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set values at unaligned addresses.
		m.SetUnaligned(PageSize*3, []byte{1})
		m.SetUnaligned(PageSize*5, []byte{42})
		m.SetUnaligned(PageSize*6, []byte{123})
		// Calculate the Merkle root of the subtrees.
		p3 := m.MerkleizeSubtree((1 << PageKeySize) | 3)
		p5 := m.MerkleizeSubtree((1 << PageKeySize) | 5)
		p6 := m.MerkleizeSubtree((1 << PageKeySize) | 6)
		// Manually combine the subtrees and compare with the subtree merkle function.
		z := zeroHashes[PageAddrSize-5]
		r1 := HashPair(
			HashPair(
				HashPair(z, z), // 0,1
				HashPair(z, p3), // 2,3
			),
			HashPair(
				HashPair(z, p5), // 4,5
				HashPair(p6, z), // 6,7
			),
		)
		r2 := m.MerkleizeSubtree(1 << (PageKeySize - 3))
		// Verify the manually combined subtrees match the subtree merkle function.
		require.Equal(t, r1, r2, "expecting manual page combination to match subtree merkle func")
	})

	// Test case for invalidating a page.
	t.Run("invalidate page", func(t *testing.T) {
		// Initialize a new memory instance.
		m := NewMemory()
		// Set a zero value at an unaligned address.
		m.SetUnaligned(0xF000, []byte{0})
		// Verify the root matches the expected zero hash.
		require.Equal(t, zeroHashes[64-5], m.MerkleRoot(), "zero at first")
		// Set a non-zero value at an unaligned address.
		m.SetUnaligned(0xF004, []byte{1})
		// Verify the root does not match the expected zero hash.
		require.NotEqual(t, zeroHashes[64-5], m.MerkleRoot(), "non-zero")
		// Set the value back to zero.
		m.SetUnaligned(0xF004, []byte{0})
		// Verify the root matches the expected zero hash again.
		require.Equal(t, zeroHashes[64-5], m.MerkleRoot(), "zero again")
	})
}
func TestMemoryReadWrite(t *testing.T) {
	t.Run("large random", func(t *testing.T) {
		m := NewMemory()
		data := make([]byte, 20_000)
		_, err := rand.Read(data[:])
		require.NoError(t, err)
		require.NoError(t, m.SetMemoryRange(0, bytes.NewReader(data)))
		for _, i := range []uint64{0, 1, 2, 3, 4, 5, 6, 7, 1000, 3333, 4095, 4096, 4097, 20_000 - 32} {
			for s := uint64(1); s <= 32; s++ {
				var res [32]byte
				m.GetUnaligned(i, res[:s])
				var expected [32]byte
				copy(expected[:s], data[i:i+s])
				require.Equalf(t, expected, res, "read %d at %d", s, i)
			}
		}
	})

	t.Run("repeat range", func(t *testing.T) {
		m := NewMemory()
		data := []byte(strings.Repeat("under the big bright yellow sun ", 40))
		require.NoError(t, m.SetMemoryRange(0x1337, bytes.NewReader(data)))
		res, err := io.ReadAll(m.ReadMemoryRange(0x1337-10, uint64(len(data)+20)))
		require.NoError(t, err)
		require.Equal(t, make([]byte, 10), res[:10], "empty start")
		require.Equal(t, data, res[10:len(res)-10], "result")
		require.Equal(t, make([]byte, 10), res[len(res)-10:], "empty end")
	})

	t.Run("read-write", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(12, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE})
		var tmp [5]byte
		m.GetUnaligned(12, tmp[:])
		require.Equal(t, [5]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}, tmp)
		m.SetUnaligned(12, []byte{0xAA, 0xBB, 0x1C, 0xDD, 0xEE})
		m.GetUnaligned(12, tmp[:])
		require.Equal(t, [5]byte{0xAA, 0xBB, 0x1C, 0xDD, 0xEE}, tmp)
	})

	t.Run("read-write-unaligned", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(13, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE})
		var tmp [5]byte
		m.GetUnaligned(13, tmp[:])
		require.Equal(t, [5]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}, tmp)
		m.SetUnaligned(13, []byte{0xAA, 0xBB, 0x1C, 0xDD, 0xEE})
		m.GetUnaligned(13, tmp[:])
		require.Equal(t, [5]byte{0xAA, 0xBB, 0x1C, 0xDD, 0xEE}, tmp)
	})
}

func TestMemoryJSON(t *testing.T) {
	m := NewMemory()
	m.SetUnaligned(8, []byte{123})
	dat, err := json.Marshal(m)
	require.NoError(t, err)
	var res Memory
	require.NoError(t, json.Unmarshal(dat, &res))
	var dest [1]byte
	m.GetUnaligned(8, dest[:])
	require.Equal(t, uint8(123), dest[0])
}
