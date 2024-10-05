package fast

import (
	"math/rand"
	"testing"
)

const (
	smallDataset  = 1_000
	mediumDataset = 100_000
	largeDataset  = 1_000_000
)

func BenchmarkMemoryOperations(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func(b *testing.B, m *Memory)
	}{
		{"RandomReadWrite_Small", benchRandomReadWrite(smallDataset)},
		{"RandomReadWrite_Medium", benchRandomReadWrite(mediumDataset)},
		{"RandomReadWrite_Large", benchRandomReadWrite(largeDataset)},
		{"SequentialReadWrite_Small", benchSequentialReadWrite(smallDataset)},
		{"SequentialReadWrite_Large", benchSequentialReadWrite(largeDataset)},
		{"SparseMemoryUsage", benchSparseMemoryUsage},
		{"DenseMemoryUsage", benchDenseMemoryUsage},
		{"SmallFrequentUpdates", benchSmallFrequentUpdates},
		{"MerkleProofGeneration_Small", benchMerkleProofGeneration(smallDataset)},
		{"MerkleProofGeneration_Large", benchMerkleProofGeneration(largeDataset)},
		{"MerkleRootCalculation_Small", benchMerkleRootCalculation(smallDataset)},
		{"MerkleRootCalculation_Large", benchMerkleRootCalculation(largeDataset)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			m := NewMemory()
			b.ResetTimer()
			bm.fn(b, m)
		})
	}
}

func benchRandomReadWrite(size int) func(b *testing.B, m *Memory) {
	return func(b *testing.B, m *Memory) {
		addresses := make([]uint64, size)
		for i := range addresses {
			addresses[i] = rand.Uint64()
		}
		data := make([]byte, 8)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addr := addresses[i%len(addresses)]
			if i%2 == 0 {
				m.SetUnaligned(addr, data)
			} else {
				m.GetUnaligned(addr, data)
			}
		}
	}
}

func benchSequentialReadWrite(size int) func(b *testing.B, m *Memory) {
	return func(b *testing.B, m *Memory) {
		data := make([]byte, 8)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addr := uint64(i % size)
			if i%2 == 0 {
				m.SetUnaligned(addr, data)
			} else {
				m.GetUnaligned(addr, data)
			}
		}
	}
}

func benchSparseMemoryUsage(b *testing.B, m *Memory) {
	data := make([]byte, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := uint64(i) * 10_000_000 // Large gaps between addresses
		m.SetUnaligned(addr, data)
	}
}

func benchDenseMemoryUsage(b *testing.B, m *Memory) {
	data := make([]byte, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := uint64(i) * 8 // Contiguous 8-byte allocations
		m.SetUnaligned(addr, data)
	}
}

func benchSmallFrequentUpdates(b *testing.B, m *Memory) {
	data := make([]byte, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := uint64(rand.Intn(1000000)) // Confined to a smaller range
		m.SetUnaligned(addr, data)
	}
}

func benchMerkleProofGeneration(size int) func(b *testing.B, m *Memory) {
	return func(b *testing.B, m *Memory) {
		// Setup: allocate some memory
		for i := 0; i < size; i++ {
			m.SetUnaligned(uint64(i)*8, []byte{byte(i)})
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			addr := uint64(rand.Intn(size) * 8)
			_ = m.MerkleProof(addr)
		}
	}
}

func benchMerkleRootCalculation(size int) func(b *testing.B, m *Memory) {
	return func(b *testing.B, m *Memory) {
		// Setup: allocate some memory
		for i := 0; i < size; i++ {
			m.SetUnaligned(uint64(i)*8, []byte{byte(i)})
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = m.MerkleRoot()
		}
	}
}
