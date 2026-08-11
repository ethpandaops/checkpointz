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

	finality, err := majority.Decide(payload, len(payload))
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

	_, err := majority.Decide(payload, len(payload))
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}

func TestSplitMajority(t *testing.T) {
	payload := []*v1.Finality{
		finalityA,
		finalityB,
	}

	_, err := majority.Decide(payload, len(payload))
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}

// A lone responder must never be treated as a majority when other upstreams are
// configured but not currently ready. The threshold has to be measured against
// the configured upstream count, not against how many nodes happened to answer.
func TestLoneResponderIsNotMajorityWhenMoreUpstreamsAreConfigured(t *testing.T) {
	payload := []*v1.Finality{finalityA}

	_, err := majority.Decide(payload, 4)
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}

// With four configured upstreams, three agreeing responses clears the bar.
func TestThreeOfFourUpstreamsFormMajority(t *testing.T) {
	payload := []*v1.Finality{finalityA, finalityA, finalityA}

	finality, err := majority.Decide(payload, 4)
	if err != nil {
		t.Fatal(err)
	}

	if finality.Finalized.Root != finalityA.Finalized.Root {
		t.Errorf("Expected %v, got %v", finalityA, finality)
	}
}

// With only one upstream configured, that upstream's answer is by definition the
// only opinion available and must still be accepted.
func TestSingleConfiguredUpstreamFormsMajority(t *testing.T) {
	payload := []*v1.Finality{finalityA}

	finality, err := majority.Decide(payload, 1)
	if err != nil {
		t.Fatal(err)
	}

	if finality.Finalized.Root != finalityA.Finalized.Root {
		t.Errorf("Expected %v, got %v", finalityA, finality)
	}
}

// With two configured upstreams, a single response cannot form a majority; both
// must agree.
func TestTwoConfiguredUpstreamsRequireBothToAgree(t *testing.T) {
	payload := []*v1.Finality{finalityA}

	_, err := majority.Decide(payload, 2)
	if err != ErrNoMajorityFound {
		t.Errorf("Expected %v, got %v", ErrNoMajorityFound, err)
	}
}
