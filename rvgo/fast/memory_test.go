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
		m.SetUnaligned(0x1002221234200, []byte{0xaa, 0xbb, 0xcc, 0xdd})
		m.SetUnaligned(0x8002212342204, []byte{42})
		m.SetUnaligned(0x1337022212342000, []byte{123})
		root := m.MerkleRoot()
		proof := m.MerkleProof(0x8002212342204)
		require.Equal(t, uint32(42<<24), binary.BigEndian.Uint32(proof[4:8]))
		node := *(*[32]byte)(proof[:32])
		path := 0x8002212342204 >> 5
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

	//t.Run("random few pages", func(t *testing.T) {
	//	m := NewMemory()
	//	m.SetUnaligned(PageSize*3, []byte{1})
	//	m.SetUnaligned(PageSize*5, []byte{42})
	//	m.SetUnaligned(PageSize*6, []byte{123})
	//	p3 := m.MerkleizeNode(m.radix, (1<<PageKeySize)|3, 0)
	//	p5 := m.MerkleizeNode(m.radix, (1<<PageKeySize)|5, 0)
	//	p6 := m.MerkleizeNode(m.radix, (1<<PageKeySize)|6, 0)
	//	z := zeroHashes[PageAddrSize-5]
	//	r1 := HashPair(
	//		HashPair(
	//			HashPair(z, z),  // 0,1
	//			HashPair(z, p3), // 2,3
	//		),
	//		HashPair(
	//			HashPair(z, p5), // 4,5
	//			HashPair(p6, z), // 6,7
	//		),
	//	)
	//	r2 := m.MerkleizeNode(m.radix, 1<<(PageKeySize-3), 0)
	//	r3 := m.MerkleizeNode3(m.radix, 1, 0)
	//	require.Equal(t, r1, r2, "expecting manual page combination to match subtree merkle func")
	//	require.Equal(t, r3, r2, "expecting manual page combination to match subtree merkle func")
	//})
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
