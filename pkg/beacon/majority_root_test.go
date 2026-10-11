package beacon

import (
	"testing"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func rootWithByte(b byte) phase0.Root {
	var r phase0.Root
	r[0] = b

	return r
}

func TestMajorityRoot_LoneResponderIsNotMajorityWhenMoreUpstreamsAreConfigured(t *testing.T) {
	// One node answered (e.g. a byzantine upstream), but four are configured.
	// The other three being unreachable/quiet must not let the lone responder
	// unilaterally decide genesis.
	_, err := majorityRoot([]phase0.Root{rootWithByte(0xAA)}, 4)
	if err == nil {
		t.Fatal("expected a lone responder not to form a majority when more upstreams are configured")
	}
}

func TestMajorityRoot_ThreeOfFourAgreeingRootsFormMajority(t *testing.T) {
	root := rootWithByte(0xAA)

	got, err := majorityRoot([]phase0.Root{root, root, root}, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != root {
		t.Fatalf("expected %#x, got %#x", root, got)
	}
}

func TestMajorityRoot_SingleConfiguredUpstreamFormsMajority(t *testing.T) {
	root := rootWithByte(0xAA)

	got, err := majorityRoot([]phase0.Root{root}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != root {
		t.Fatalf("expected %#x, got %#x", root, got)
	}
}

func TestMajorityRoot_TwoConfiguredUpstreamsRequireBothToAgree(t *testing.T) {
	_, err := majorityRoot([]phase0.Root{rootWithByte(0xAA)}, 2)
	if err == nil {
		t.Fatal("expected a single response out of two configured upstreams not to form a majority")
	}
}

func TestMajorityRoot_DisagreeingRootsFormNoMajority(t *testing.T) {
	_, err := majorityRoot([]phase0.Root{rootWithByte(0xAA), rootWithByte(0xBB), rootWithByte(0xCC)}, 3)
	if err == nil {
		t.Fatal("expected three distinct roots to form no majority")
	}
}

func TestMajorityRoot_NoResponsesFormNoMajority(t *testing.T) {
	_, err := majorityRoot(nil, 4)
	if err == nil {
		t.Fatal("expected no responses to form no majority")
	}
}
