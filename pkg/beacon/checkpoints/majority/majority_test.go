package majority

import (
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

var (
	checkpointA = &phase0.Checkpoint{
		Epoch: 100,
		Root:  phase0.Root{0x01},
	}

	checkpointB = &phase0.Checkpoint{
		Epoch: 101,
		Root:  phase0.Root{0x02},
	}

	checkpointC = &phase0.Checkpoint{
		Epoch: 102,
		Root:  phase0.Root{0x03},
	}

	checkpointD = &phase0.Checkpoint{
		Epoch: 103,
		Root:  phase0.Root{0x04},
	}

	finalityA = &v1.Finality{
		Finalized:         checkpointA,
		Justified:         checkpointB,
		PreviousJustified: checkpointB,
	}
	finalityB = &v1.Finality{
		Finalized:         checkpointB,
		Justified:         checkpointC,
		PreviousJustified: checkpointC,
	}
	finalityC = &v1.Finality{
		Finalized:         checkpointC,
		Justified:         checkpointD,
		PreviousJustified: checkpointD,
	}

	majority = New()
)

func TestBasicMajority(t *testing.T) {
	payload := []*v1.Finality{
		finalityA,
		finalityB,
		finalityA,
	}

	finality, err := majority.Decide(payload)
	if err != nil {
		t.Fatal(err)
	}

	if finality.Finalized.Root != finalityA.Finalized.Root {
		t.Errorf("Expected %v, got %v", finalityA, finality)
	}
}

func TestNonMajority(t *testing.T) {
	payload := []*v1.Finality{
		finalityA,
		finalityB,
		finalityC,
	}

	_, err := majority.Decide(payload)
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}

func TestSplitMajority(t *testing.T) {
	payload := []*v1.Finality{
		finalityA,
		finalityB,
	}

	_, err := majority.Decide(payload)
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}
