package ssz

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

// createTestState creates a minimal beacon state for testing.
// In real usage, states are ~100-300MB. We use smaller states for benchmarks
// to keep CI fast, but the relative performance difference still applies.
func createTestState(validatorCount int) *spec.VersionedBeaconState {
	validators := make([]*phase0.Validator, validatorCount)
	balances := make([]phase0.Gwei, validatorCount)
	prevParticipation := make([]altair.ParticipationFlags, validatorCount)
	currParticipation := make([]altair.ParticipationFlags, validatorCount)
	inactivityScores := make([]uint64, validatorCount)

	for i := 0; i < validatorCount; i++ {
		validators[i] = &phase0.Validator{
			PublicKey:             phase0.BLSPubKey{},
			WithdrawalCredentials: make([]byte, 32),
			EffectiveBalance:      32000000000,
			Slashed:               false,
			ActivationEpoch:       0,
			ExitEpoch:             ^phase0.Epoch(0),
			WithdrawableEpoch:     ^phase0.Epoch(0),
		}
		balances[i] = 32000000000
		prevParticipation[i] = 0
		currParticipation[i] = 0
		inactivityScores[i] = 0
	}

	// Create sync committee with proper pubkeys
	syncCommitteePubkeys := make([]phase0.BLSPubKey, 512)
	for i := range syncCommitteePubkeys {
		syncCommitteePubkeys[i] = phase0.BLSPubKey{}
	}

	syncCommittee := &altair.SyncCommittee{
		Pubkeys:         syncCommitteePubkeys,
		AggregatePubkey: phase0.BLSPubKey{},
	}

	state := &deneb.BeaconState{
		GenesisTime:           1606824023,
		GenesisValidatorsRoot: phase0.Root{},
		Slot:                  1000000,
		Fork: &phase0.Fork{
			PreviousVersion: phase0.Version{0x04, 0x00, 0x00, 0x00},
			CurrentVersion:  phase0.Version{0x04, 0x00, 0x00, 0x00},
			Epoch:           0,
		},
		LatestBlockHeader: &phase0.BeaconBlockHeader{
			Slot:          999999,
			ProposerIndex: 0,
			ParentRoot:    phase0.Root{},
			StateRoot:     phase0.Root{},
			BodyRoot:      phase0.Root{},
		},
		BlockRoots:      make([]phase0.Root, 8192),
		StateRoots:      make([]phase0.Root, 8192),
		HistoricalRoots: []phase0.Root{},
		ETH1Data: &phase0.ETH1Data{
			DepositRoot:  phase0.Root{},
			DepositCount: 0,
			BlockHash:    make([]byte, 32),
		},
		ETH1DataVotes:               []*phase0.ETH1Data{},
		ETH1DepositIndex:            0,
		Validators:                  validators,
		Balances:                    balances,
		RANDAOMixes:                 make([]phase0.Root, 65536),
		Slashings:                   make([]phase0.Gwei, 8192),
		PreviousEpochParticipation:  prevParticipation,
		CurrentEpochParticipation:   currParticipation,
		JustificationBits:           []byte{0},
		PreviousJustifiedCheckpoint: &phase0.Checkpoint{},
		CurrentJustifiedCheckpoint:  &phase0.Checkpoint{},
		FinalizedCheckpoint:         &phase0.Checkpoint{},
		InactivityScores:            inactivityScores,
		CurrentSyncCommittee:        syncCommittee,
		NextSyncCommittee:           syncCommittee,
		LatestExecutionPayloadHeader: &deneb.ExecutionPayloadHeader{
			ParentHash:       phase0.Hash32{},
			FeeRecipient:     [20]byte{},
			StateRoot:        phase0.Root{},
			ReceiptsRoot:     phase0.Root{},
			LogsBloom:        [256]byte{},
			PrevRandao:       [32]byte{},
			BlockNumber:      0,
			GasLimit:         0,
			GasUsed:          0,
			Timestamp:        0,
			ExtraData:        []byte{},
			BaseFeePerGas:    uint256.NewInt(0),
			BlockHash:        phase0.Hash32{},
			TransactionsRoot: phase0.Root{},
			WithdrawalsRoot:  phase0.Root{},
			BlobGasUsed:      0,
			ExcessBlobGas:    0,
		},
		NextWithdrawalIndex:          0,
		NextWithdrawalValidatorIndex: 0,
		HistoricalSummaries:          []*capella.HistoricalSummary{},
	}

	return &spec.VersionedBeaconState{
		Version: spec.DataVersionDeneb,
		Deneb:   state,
	}
}

func TestEncodeStateSSZ(t *testing.T) {
	encoder := NewEncoder(false, 0)
	state := createTestState(100)

	data, err := encoder.EncodeStateSSZ(state)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}

func TestWriteStateSSZ(t *testing.T) {
	encoder := NewEncoder(false, 0)
	state := createTestState(100)

	var buf bytes.Buffer

	n, err := encoder.WriteStateSSZ(context.Background(), &buf, state)
	require.NoError(t, err)
	require.Greater(t, n, int64(0))

	// Verify output matches EncodeStateSSZ
	expected, err := encoder.EncodeStateSSZ(state)
	require.NoError(t, err)
	require.Equal(t, expected, buf.Bytes())
}

func TestStateSizeSSZ(t *testing.T) {
	encoder := NewEncoder(false, 0)
	state := createTestState(100)

	size, err := encoder.StateSizeSSZ(state)
	require.NoError(t, err)
	require.Greater(t, size, 0)

	// Verify size matches actual encoded size
	data, err := encoder.EncodeStateSSZ(state)
	require.NoError(t, err)
	require.Equal(t, size, len(data))
}

func TestWriteStateSSZ_MemoryBounded(t *testing.T) {
	state := createTestState(100)

	// Get the size of this state
	unboundedEncoder := NewEncoder(false, 0)

	size, err := unboundedEncoder.StateSizeSSZ(state)
	require.NoError(t, err)

	// Create encoder with memory budget that allows exactly 2 concurrent encodings
	memoryBudget := int64(size * 2)
	encoder := NewEncoder(false, memoryBudget)

	// Test basic encoding works
	var buf bytes.Buffer

	n, err := encoder.WriteStateSSZ(context.Background(), &buf, state)
	require.NoError(t, err)
	require.Equal(t, int64(size), n)
}

func TestWriteStateSSZ_MemoryBounded_ContextCancellation(t *testing.T) {
	state := createTestState(100)

	// Create encoder with very small memory budget (smaller than state size)
	// This will cause the semaphore to block
	encoder := NewEncoder(false, 1) // 1 byte budget

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail immediately due to cancelled context
	_, err := encoder.WriteStateSSZ(ctx, io.Discard, state)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

// BenchmarkEncodeStateSSZ benchmarks the original non-streaming encoder.
// This allocates a new buffer for each call.
func BenchmarkEncodeStateSSZ(b *testing.B) {
	benchmarks := []struct {
		name           string
		validatorCount int
	}{
		{"100_validators", 100},
		{"1000_validators", 1000},
		{"10000_validators", 10000},
	}

	encoder := NewEncoder(false, 0)

	for _, bm := range benchmarks {
		state := createTestState(bm.validatorCount)

		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				data, err := encoder.EncodeStateSSZ(state)
				if err != nil {
					b.Fatal(err)
				}

				// Prevent compiler optimization
				if len(data) == 0 {
					b.Fatal("empty data")
				}
			}
		})
	}
}

// BenchmarkWriteStateSSZ benchmarks the streaming encoder with pooled buffers.
// This reuses buffers across calls, reducing allocations.
func BenchmarkWriteStateSSZ(b *testing.B) {
	benchmarks := []struct {
		name           string
		validatorCount int
	}{
		{"100_validators", 100},
		{"1000_validators", 1000},
		{"10000_validators", 10000},
	}

	encoder := NewEncoder(false, 0)

	for _, bm := range benchmarks {
		state := createTestState(bm.validatorCount)

		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				n, err := encoder.WriteStateSSZ(context.Background(), io.Discard, state)
				if err != nil {
					b.Fatal(err)
				}

				// Prevent compiler optimization
				if n == 0 {
					b.Fatal("no bytes written")
				}
			}
		})
	}
}

// BenchmarkWriteStateSSZ_Parallel tests buffer pool contention under parallel load.
func BenchmarkWriteStateSSZ_Parallel(b *testing.B) {
	encoder := NewEncoder(false, 0)
	state := createTestState(1000)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n, err := encoder.WriteStateSSZ(context.Background(), io.Discard, state)
			if err != nil {
				b.Fatal(err)
			}

			if n == 0 {
				b.Fatal("no bytes written")
			}
		}
	})
}

// BenchmarkStateSizeSSZ benchmarks size calculation (no encoding).
func BenchmarkStateSizeSSZ(b *testing.B) {
	encoder := NewEncoder(false, 0)
	state := createTestState(10000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		size, err := encoder.StateSizeSSZ(state)
		if err != nil {
			b.Fatal(err)
		}

		if size == 0 {
			b.Fatal("zero size")
		}
	}
}
