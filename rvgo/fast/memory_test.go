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

func TestMemoryMerkleProof(t *testing.T) {
	t.Run("nearly empty tree", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0x10000, []byte{0xaa, 0xbb, 0xcc, 0xdd})
		proof := m.MerkleProof(0x10000)
		require.Equal(t, uint32(0xaabbccdd), binary.BigEndian.Uint32(proof[:4]))
		for i := 0; i < 32-5; i++ {
			require.Equal(t, zeroHashes[i][:], proof[32+i*32:32+i*32+32], "empty siblings")
		}
	})

	t.Run("fuller tree", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0x10000, []byte{0xaa, 0xbb, 0xcc, 0xdd})
		m.SetUnaligned(0x80004, []byte{42})
		m.SetUnaligned(0x13370000, []byte{123})
		root := m.MerkleRoot()
		proof := m.MerkleProof(0x80004)
		require.Equal(t, uint32(42<<24), binary.BigEndian.Uint32(proof[4:8]))
		node := *(*[32]byte)(proof[:32])
		path := 0x80004 >> 5
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

	t.Run("consistency test", func(t *testing.T) {
		m := NewMemory()
		addr := uint64(0x1234560000000)
		m.SetUnaligned(addr, []byte{1})
		proof1 := m.MerkleProof(addr)
		proof2 := m.MerkleProof(addr)
		require.Equal(t, proof1, proof2, "Proofs for the same address should be consistent")
	})

	t.Run("stress test", func(t *testing.T) {
		m := NewMemory()
		var addresses []uint64
		for i := uint64(0); i < 10000; i++ {
			addr := i * 0x1000000 // Spread out addresses
			addresses = append(addresses, addr)
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})
	t.Run("boundary addresses", func(t *testing.T) {
		m := NewMemory()
		addresses := []uint64{
			//0x0000000000000 - 1, // Just before first level
			0x0000000000000,     // Start of first level
			0x0400000000000 - 1, // End of first level
			0x0400000000000,     // Start of second level
			0x3C00000000000 - 1, // End of fourth level
			0x3C00000000000,     // Start of fifth level
			0x3FFFFFFFFFFF,      // Maximum address
		}
		for i, addr := range addresses {
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})
	t.Run("multiple levels", func(t *testing.T) {
		m := NewMemory()
		addresses := []uint64{
			0x0000000000000,
			0x0400000000000,
			0x0800000000000,
			0x0C00000000000,
			0x1000000000000,
			0x1400000000000,
		}
		for i, addr := range addresses {
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})

	t.Run("sparse tree", func(t *testing.T) {
		m := NewMemory()
		addresses := []uint64{
			0x0000000000000,
			0x0000400000000,
			0x0004000000000,
			0x0040000000000,
			0x0400000000000,
			0x3C00000000000,
		}
		for i, addr := range addresses {
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})

	t.Run("adjacent addresses", func(t *testing.T) {
		m := NewMemory()
		baseAddr := uint64(0x0400000000000)
		for i := uint64(0); i < 16; i++ {
			m.SetUnaligned(baseAddr+i, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for i := uint64(0); i < 16; i++ {
			proof := m.MerkleProof(baseAddr + i)
			verifyProof(t, root, proof, baseAddr+i)
		}
	})

	t.Run("cross-page addresses", func(t *testing.T) {
		m := NewMemory()
		pageSize := uint64(4096)
		addresses := []uint64{
			pageSize - 2,
			pageSize - 1,
			pageSize,
			pageSize + 1,
			2*pageSize - 2,
			2*pageSize - 1,
			2 * pageSize,
			2*pageSize + 1,
		}
		for i, addr := range addresses {
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})

	t.Run("large addresses", func(t *testing.T) {
		m := NewMemory()
		addresses := []uint64{
			0x10_00_00_00_00_00_00_00,
			0x10_00_00_00_00_00_00_02,
			0x10_00_00_00_00_00_00_04,
			0x10_00_00_00_00_00_00_06,
		}
		for i, addr := range addresses {
			m.SetUnaligned(addr, []byte{byte(i + 1)})
		}
		root := m.MerkleRoot()
		for _, addr := range addresses {
			proof := m.MerkleProof(addr)
			verifyProof(t, root, proof, addr)
		}
	})
}
func TestMerkleProofWithPartialPaths(t *testing.T) {
	testCases := []struct {
		name        string
		setupMemory func(*Memory)
		proofAddr   uint64
	}{
		{
			name: "Path ends at level 1",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_00_00_00_00_00_00_00, []byte{1})
			},
			proofAddr: 0x20_00_00_00_00_00_00_00,
		},
		{
			name: "Path ends at level 2",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_00_00_00_00_00_00_00, []byte{1})
			},
			proofAddr: 0x11_00_00_00_00_00_00_00,
		},
		{
			name: "Path ends at level 3",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_10_00_00_00_00_00_00, []byte{1})
			},
			proofAddr: 0x10_11_00_00_00_00_00_00,
		},
		{
			name: "Path ends at level 4",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_10_10_00_00_00_00_00, []byte{1})
			},
			proofAddr: 0x10_10_11_00_00_00_00_00,
		},
		{
			name: "Full path to level 5, page doesn't exist",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_10_10_10_00_00_00_00, []byte{1})
			},
			proofAddr: 0x10_10_10_10_10_00_00_00, // Different page in the same level 5 node
		},
		{
			name: "Path ends at level 3, check different page offsets",
			setupMemory: func(m *Memory) {
				m.SetUnaligned(0x10_10_00_00_00_00_00_00, []byte{1})
				m.SetUnaligned(0x10_10_00_00_00_00_10_00, []byte{2})
			},
			proofAddr: 0x10_10_00_00_00_00_20_00, // Different offset in the same page
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMemory()
			tc.setupMemory(m)

			proof := m.MerkleProof(tc.proofAddr)

			// Check that the proof is filled correctly
			verifyProof(t, m.MerkleRoot(), proof, tc.proofAddr)
			//checkProof(t, proof, tc.expectedDepth)
		})
	}
}

func verifyProof(t *testing.T, expectedRoot [32]byte, proof [ProofLen * 32]byte, addr uint64) {
	node := *(*[32]byte)(proof[:32])
	path := addr >> 5
	for i := 32; i < len(proof); i += 32 {
		sib := *(*[32]byte)(proof[i : i+32])
		if path&1 != 0 {
			node = HashPair(sib, node)
		} else {
			node = HashPair(node, sib)
		}
		path >>= 1
	}
	require.Equal(t, expectedRoot, node, "proof must verify for address 0x%x", addr)
}

func TestMemoryMerkleRoot(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m := NewMemory()
		root := m.MerkleRoot()
		require.Equal(t, zeroHashes[64-5], root, "fully zeroed memory should have expected zero hash")
	})
	t.Run("empty page", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0xF000, []byte{0})
		root := m.MerkleRoot()
		require.Equal(t, zeroHashes[64-5], root, "fully zeroed memory should have expected zero hash")
	})
	t.Run("single page", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0xF000, []byte{1})
		root := m.MerkleRoot()
		require.NotEqual(t, zeroHashes[64-5], root, "non-zero memory")
	})
	t.Run("repeat zero", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0xF000, []byte{0})
		m.SetUnaligned(0xF004, []byte{0})
		root := m.MerkleRoot()
		require.Equal(t, zeroHashes[64-5], root, "zero still")
	})
	t.Run("two empty pages", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(PageSize*3, []byte{0})
		m.SetUnaligned(PageSize*10, []byte{0})
		root := m.MerkleRoot()
		require.Equal(t, zeroHashes[64-5], root, "zero still")
	})

	t.Run("random few pages", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(PageSize*3, []byte{1})
		m.SetUnaligned(PageSize*5, []byte{42})
		m.SetUnaligned(PageSize*6, []byte{123})

		p0 := m.MerkleizeNodeLevel1(m.radix, 0, 8)
		p1 := m.MerkleizeNodeLevel1(m.radix, 0, 9)
		p2 := m.MerkleizeNodeLevel1(m.radix, 0, 10)
		p3 := m.MerkleizeNodeLevel1(m.radix, 0, 11)
		p4 := m.MerkleizeNodeLevel1(m.radix, 0, 12)
		p5 := m.MerkleizeNodeLevel1(m.radix, 0, 13)
		p6 := m.MerkleizeNodeLevel1(m.radix, 0, 14)
		p7 := m.MerkleizeNodeLevel1(m.radix, 0, 15)

		r1 := HashPair(
			HashPair(
				HashPair(p0, p1), // 0,1
				HashPair(p2, p3), // 2,3
			),
			HashPair(
				HashPair(p4, p5), // 4,5
				HashPair(p6, p7), // 6,7
			),
		)
		r2 := m.MerkleizeNodeLevel1(m.radix, 0, 1)
		require.Equal(t, r1, r2, "expecting manual page combination to match subtree merkle func")
	})

	t.Run("invalidate page", func(t *testing.T) {
		m := NewMemory()
		m.SetUnaligned(0xF000, []byte{0})
		require.Equal(t, zeroHashes[64-5], m.MerkleRoot(), "zero at first")
		m.SetUnaligned(0xF004, []byte{1})
		require.NotEqual(t, zeroHashes[64-5], m.MerkleRoot(), "non-zero")
		m.SetUnaligned(0xF004, []byte{0})
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

func TestMemoryBinary(t *testing.T) {
	m := NewMemory()
	m.SetUnaligned(8, []byte{123})
	ser := new(bytes.Buffer)
	err := m.Serialize(ser)
	require.NoError(t, err, "must serialize state")
	m2 := NewMemory()
	err = m2.Deserialize(ser)
	require.NoError(t, err, "must deserialize state")
	var dest [1]byte
	m.GetUnaligned(8, dest[:])
	require.Equal(t, uint8(123), dest[0])
}
