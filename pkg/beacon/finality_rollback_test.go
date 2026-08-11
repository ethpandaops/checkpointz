package beacon

import (
	"context"
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus/hooks/test"
)

func checkpointAt(epoch uint64, root byte) *v1.Finality {
	r := phase0.Root{root}

	return &v1.Finality{
		Finalized:         &phase0.Checkpoint{Epoch: phase0.Epoch(epoch), Root: r},
		Justified:         &phase0.Checkpoint{Epoch: phase0.Epoch(epoch), Root: r},
		PreviousJustified: &phase0.Checkpoint{Epoch: phase0.Epoch(epoch), Root: r},
	}
}

func TestNextHeadCheckpoint_RejectsEpochRegression(t *testing.T) {
	current := checkpointAt(100, 0xAA)
	regressed := checkpointAt(50, 0xBB) // older epoch, different fork

	next, updated, rejected := nextHeadCheckpoint(current, regressed)

	if !rejected {
		t.Fatal("expected a lower-epoch majority to be rejected")
	}

	if updated {
		t.Fatal("a rejected checkpoint must not be reported as updated")
	}

	if next != current {
		t.Fatalf("expected head to remain unchanged at epoch %d, got %d", current.Finalized.Epoch, next.Finalized.Epoch)
	}
}

func TestNextHeadCheckpoint_AcceptsAdvancingEpoch(t *testing.T) {
	current := checkpointAt(100, 0xAA)
	advanced := checkpointAt(101, 0xCC)

	next, updated, rejected := nextHeadCheckpoint(current, advanced)

	if rejected {
		t.Fatal("an advancing epoch must not be rejected")
	}

	if !updated || next != advanced {
		t.Fatal("expected head to move to the advancing checkpoint")
	}
}

func TestNextHeadCheckpoint_NoOpWhenRootUnchanged(t *testing.T) {
	current := checkpointAt(100, 0xAA)
	same := checkpointAt(100, 0xAA)

	next, updated, rejected := nextHeadCheckpoint(current, same)

	if rejected {
		t.Fatal("an unchanged checkpoint must not be rejected")
	}

	if updated {
		t.Fatal("an unchanged root must not be reported as an update")
	}

	if next != current {
		t.Fatal("expected the existing head pointer to be preserved")
	}
}

func TestNextHeadCheckpoint_AcceptsFirstCheckpoint(t *testing.T) {
	first := checkpointAt(0, 0xAA)

	next, updated, rejected := nextHeadCheckpoint(nil, first)

	if rejected || !updated || next != first {
		t.Fatal("expected the first-ever checkpoint to be accepted")
	}
}

// TestCheckForNewServingCheckpoint_DoesNotDownloadOnRegressedEpoch drives the
// real production entrypoint (not just the pure helper) to prove a regressed or
// equal head epoch never reaches downloadServingCheckpoint. If it did, this test
// would hang/fail trying to reach a nil node set; instead it must return cleanly
// with the serving bundle untouched.
func TestCheckForNewServingCheckpoint_DoesNotDownloadOnRegressedEpoch(t *testing.T) {
	logger, _ := test.NewNullLogger()

	servingBundle := checkpointAt(100, 0xAA)

	d := &Default{
		log:           logger,
		head:          checkpointAt(50, 0xBB), // regressed relative to serving
		servingBundle: servingBundle,
	}

	if err := d.checkForNewServingCheckpoint(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.servingBundle != servingBundle {
		t.Fatal("serving bundle must not change when the head epoch has not advanced")
	}
}

func TestCheckForNewServingCheckpoint_DoesNotDownloadOnSameEpoch(t *testing.T) {
	logger, _ := test.NewNullLogger()

	servingBundle := checkpointAt(100, 0xAA)

	d := &Default{
		log:           logger,
		head:          checkpointAt(100, 0xCC), // same epoch, different root
		servingBundle: servingBundle,
	}

	if err := d.checkForNewServingCheckpoint(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.servingBundle != servingBundle {
		t.Fatal("serving bundle must not change on a same-epoch root change")
	}
}
