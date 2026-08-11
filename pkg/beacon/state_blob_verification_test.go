package beacon

import (
	"testing"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// minimalBeaconState builds a structurally-valid (all ssz-size slices at spec
// length) phase0 BeaconState so HashTreeRoot() succeeds.
func minimalBeaconState(slot uint64) *spec.VersionedBeaconState {
	return &spec.VersionedBeaconState{
		Version: spec.DataVersionPhase0,
		Phase0: &phase0.BeaconState{
			Slot:                        phase0.Slot(slot),
			Fork:                        &phase0.Fork{},
			LatestBlockHeader:           &phase0.BeaconBlockHeader{},
			BlockRoots:                  make([]phase0.Root, 8192),
			StateRoots:                  make([]phase0.Root, 8192),
			HistoricalRoots:             []phase0.Root{},
			ETH1Data:                    &phase0.ETH1Data{BlockHash: make([]byte, 32)},
			ETH1DataVotes:               []*phase0.ETH1Data{},
			Validators:                  []*phase0.Validator{},
			Balances:                    []phase0.Gwei{},
			RANDAOMixes:                 make([]phase0.Root, 65536),
			Slashings:                   make([]phase0.Gwei, 8192),
			JustificationBits:           make([]byte, 1),
			PreviousJustifiedCheckpoint: &phase0.Checkpoint{},
			CurrentJustifiedCheckpoint:  &phase0.Checkpoint{},
			FinalizedCheckpoint:         &phase0.Checkpoint{},
		},
	}
}

func TestVerifyBeaconStateRoot_AcceptsMatchingRoot(t *testing.T) {
	state := minimalBeaconState(12345)

	realRoot, err := state.HashTreeRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := verifyBeaconStateRoot(state, phase0.Root(realRoot)); err != nil {
		t.Fatalf("expected a state matching its own hash tree root to verify, got: %v", err)
	}
}

func TestVerifyBeaconStateRoot_RejectsMismatchedRoot(t *testing.T) {
	state := minimalBeaconState(12345)

	var claimedRoot phase0.Root
	claimedRoot[0] = 0xDE
	claimedRoot[1] = 0xAD

	if err := verifyBeaconStateRoot(state, claimedRoot); err == nil {
		t.Fatal("expected verification to fail for a state that does not hash to the claimed root")
	}
}

func kzgCommitment(b byte) deneb.KZGCommitment {
	var c deneb.KZGCommitment
	c[0] = b

	return c
}

func TestVerifyBlobSidecarCommitments_AcceptsMatchingCommitments(t *testing.T) {
	commitments := []deneb.KZGCommitment{kzgCommitment(0x01), kzgCommitment(0x02)}
	sidecars := []*deneb.BlobSidecar{
		{Index: 0, KZGCommitment: commitments[0]},
		{Index: 1, KZGCommitment: commitments[1]},
	}

	if err := verifyBlobSidecarCommitments(commitments, sidecars); err != nil {
		t.Fatalf("expected matching commitments to verify, got: %v", err)
	}
}

func TestVerifyBlobSidecarCommitments_RejectsMismatchedCommitment(t *testing.T) {
	commitments := []deneb.KZGCommitment{kzgCommitment(0x01)}
	sidecars := []*deneb.BlobSidecar{
		{Index: 0, KZGCommitment: kzgCommitment(0xFF)}, // fabricated, does not match the block
	}

	if err := verifyBlobSidecarCommitments(commitments, sidecars); err == nil {
		t.Fatal("expected verification to fail for a sidecar whose commitment does not match the block")
	}
}

func TestVerifyBlobSidecarCommitments_RejectsOutOfRangeIndex(t *testing.T) {
	commitments := []deneb.KZGCommitment{kzgCommitment(0x01)}
	sidecars := []*deneb.BlobSidecar{
		{Index: 5, KZGCommitment: kzgCommitment(0x01)}, // block only committed to 1 blob
	}

	if err := verifyBlobSidecarCommitments(commitments, sidecars); err == nil {
		t.Fatal("expected verification to fail for a sidecar index beyond the block's committed blobs")
	}
}

func TestVerifyBlobSidecarCommitments_AcceptsEmpty(t *testing.T) {
	if err := verifyBlobSidecarCommitments(nil, nil); err != nil {
		t.Fatalf("expected no commitments and no sidecars to verify trivially, got: %v", err)
	}
}
