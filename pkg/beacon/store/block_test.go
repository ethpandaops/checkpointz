package store

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/holiman/uint256"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// benchNamespaceCounter ensures unique namespaces for benchmarks to avoid Prometheus metric conflicts.
var benchNamespaceCounter atomic.Uint64

func uniqueNamespace(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, benchNamespaceCounter.Add(1))
}

func createTestBlock(slot phase0.Slot) *spec.VersionedSignedBeaconBlock {
	return &spec.VersionedSignedBeaconBlock{
		Version: spec.DataVersionDeneb,
		Deneb: &deneb.SignedBeaconBlock{
			Message: &deneb.BeaconBlock{
				Slot:          slot,
				ProposerIndex: phase0.ValidatorIndex(slot % 1000),
				ParentRoot:    phase0.Root{byte(slot)},
				StateRoot:     phase0.Root{byte(slot), byte(slot >> 8)},
				Body: &deneb.BeaconBlockBody{
					RANDAOReveal:      phase0.BLSSignature{},
					ETH1Data:          &phase0.ETH1Data{},
					Graffiti:          [32]byte{},
					ProposerSlashings: []*phase0.ProposerSlashing{},
					AttesterSlashings: []*phase0.AttesterSlashing{},
					Attestations:      []*phase0.Attestation{},
					Deposits:          []*phase0.Deposit{},
					VoluntaryExits:    []*phase0.SignedVoluntaryExit{},
					SyncAggregate: &altair.SyncAggregate{
						SyncCommitteeBits:      bitfield.NewBitvector512(),
						SyncCommitteeSignature: phase0.BLSSignature{},
					},
					ExecutionPayload: &deneb.ExecutionPayload{
						ParentHash:    phase0.Hash32{},
						FeeRecipient:  [20]byte{},
						StateRoot:     phase0.Root{},
						ReceiptsRoot:  phase0.Root{},
						LogsBloom:     [256]byte{},
						PrevRandao:    [32]byte{},
						BlockNumber:   uint64(slot),
						GasLimit:      30000000,
						GasUsed:       15000000,
						Timestamp:     uint64(slot) * 12,
						ExtraData:     []byte{},
						BaseFeePerGas: uint256.NewInt(0),
						BlockHash:     phase0.Hash32{byte(slot)},
						Transactions:  []bellatrix.Transaction{},
						Withdrawals:   []*capella.Withdrawal{},
						BlobGasUsed:   0,
						ExcessBlobGas: 0,
					},
					BLSToExecutionChanges: []*capella.SignedBLSToExecutionChange{},
					BlobKZGCommitments:    []deneb.KZGCommitment{},
				},
			},
			Signature: phase0.BLSSignature{},
		},
	}
}

func TestBlockAddAndGet(t *testing.T) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 10}
	namespace := "test_block_a"
	blockStore := NewBlock(logger, config, namespace)

	slot := phase0.Slot(100)
	block := createTestBlock(slot)
	expiresAt := time.Now().Add(10 * time.Minute)

	// Get the root for this block
	root := phase0.Root{0x01, 0x02, 0x03}

	err := blockStore.Add(root, block, expiresAt)
	require.NoError(t, err)

	// Get by root
	retrievedBlock, err := blockStore.GetByRoot(root)
	require.NoError(t, err)
	require.NotNil(t, retrievedBlock)

	retrievedSlot, err := retrievedBlock.Slot()
	require.NoError(t, err)
	assert.Equal(t, slot, retrievedSlot)

	// Get by slot
	retrievedBlock2, err := blockStore.GetBySlot(slot)
	require.NoError(t, err)
	require.NotNil(t, retrievedBlock2)
}

func TestBlockGetBySlotNotFound(t *testing.T) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 10}
	namespace := "test_block_b"
	blockStore := NewBlock(logger, config, namespace)

	slot := phase0.Slot(200)

	retrievedBlock, err := blockStore.GetBySlot(slot)
	assert.Error(t, err)
	assert.Nil(t, retrievedBlock)
}

func TestBlockCleanup(t *testing.T) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 3}
	namespace := "test_block_cleanup"
	blockStore := NewBlock(logger, config, namespace)

	// Add 4 blocks (max is 3, so first should be evicted)
	for i := 0; i < 4; i++ {
		slot := phase0.Slot(100 + i) //nolint:gosec // test code with bounded values
		block := createTestBlock(slot)
		root := phase0.Root{byte(i)}
		expiresAt := time.Now().Add(time.Duration(i+1) * time.Minute)

		err := blockStore.Add(root, block, expiresAt)
		require.NoError(t, err)
	}

	// Give the eviction callback time to run
	time.Sleep(100 * time.Millisecond)

	// The first block (slot 100) should have been evicted (closest to expiry)
	_, err := blockStore.GetBySlot(phase0.Slot(100))
	assert.Error(t, err, "slot 100 should have been evicted")

	// Slots 101, 102, 103 should still exist
	for i := 1; i < 4; i++ {
		_, err := blockStore.GetBySlot(phase0.Slot(100 + i)) //nolint:gosec // test code with bounded values
		assert.NoError(t, err, "slot %d should still exist", 100+i)
	}
}

// BenchmarkBlockAdd benchmarks adding blocks to the store.
func BenchmarkBlockAdd(b *testing.B) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 1000}
	blockStore := NewBlock(logger, config, uniqueNamespace("bench_block_add"))

	blocks := make([]*spec.VersionedSignedBeaconBlock, b.N)
	roots := make([]phase0.Root, b.N)

	for i := 0; i < b.N; i++ {
		blocks[i] = createTestBlock(phase0.Slot(i)) //nolint:gosec // test code with bounded values
		roots[i] = phase0.Root{byte(i), byte(i >> 8), byte(i >> 16)}
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := blockStore.Add(roots[i], blocks[i], expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlockGetBySlot benchmarks slot lookups.
func BenchmarkBlockGetBySlot(b *testing.B) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 1000}
	blockStore := NewBlock(logger, config, uniqueNamespace("bench_block_get_slot"))

	// Pre-populate with blocks
	numBlocks := 100

	for i := 0; i < numBlocks; i++ {
		block := createTestBlock(phase0.Slot(i)) //nolint:gosec // test code with bounded values
		root := phase0.Root{byte(i), byte(i >> 8)}
		expiresAt := time.Now().Add(1 * time.Hour)

		err := blockStore.Add(root, block, expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slot := phase0.Slot(i % numBlocks) //nolint:gosec // test code with bounded values

		_, err := blockStore.GetBySlot(slot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlockGetByRoot benchmarks root lookups.
func BenchmarkBlockGetByRoot(b *testing.B) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 1000}
	blockStore := NewBlock(logger, config, uniqueNamespace("bench_block_get_root"))

	// Pre-populate with blocks
	numBlocks := 100
	roots := make([]phase0.Root, numBlocks)

	for i := 0; i < numBlocks; i++ {
		block := createTestBlock(phase0.Slot(i)) //nolint:gosec // test code with bounded values
		roots[i] = phase0.Root{byte(i), byte(i >> 8)}
		expiresAt := time.Now().Add(1 * time.Hour)

		err := blockStore.Add(roots[i], block, expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		root := roots[i%numBlocks]

		_, err := blockStore.GetByRoot(root)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlockEviction benchmarks the eviction behavior when cache is full.
func BenchmarkBlockEviction(b *testing.B) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 10} // Small cache to force evictions
	blockStore := NewBlock(logger, config, uniqueNamespace("bench_block_evict"))

	blocks := make([]*spec.VersionedSignedBeaconBlock, b.N)
	roots := make([]phase0.Root, b.N)

	for i := 0; i < b.N; i++ {
		blocks[i] = createTestBlock(phase0.Slot(i)) //nolint:gosec // test code with bounded values
		roots[i] = phase0.Root{byte(i), byte(i >> 8), byte(i >> 16)}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Each add after the 10th will trigger eviction
		expiresAt := time.Now().Add(time.Duration(i) * time.Millisecond)

		err := blockStore.Add(roots[i], blocks[i], expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBlockGetBySlot_Parallel benchmarks concurrent slot lookups.
func BenchmarkBlockGetBySlot_Parallel(b *testing.B) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 1000}
	blockStore := NewBlock(logger, config, uniqueNamespace("bench_block_parallel"))

	// Pre-populate with blocks
	numBlocks := 100

	for i := 0; i < numBlocks; i++ {
		block := createTestBlock(phase0.Slot(i)) //nolint:gosec // test code with bounded values
		root := phase0.Root{byte(i), byte(i >> 8)}
		expiresAt := time.Now().Add(1 * time.Hour)

		err := blockStore.Add(root, block, expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			slot := phase0.Slot(i % numBlocks) //nolint:gosec // test code with bounded values
			i++

			_, err := blockStore.GetBySlot(slot)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
