package beacon

import (
	"fmt"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

var (
	defaultSecondsPerSlot = time.Second * 12
	defaultSlotsPerEpoch  = uint64(32)
)

func TestExpireBlockAdd(t *testing.T) {
	genesis, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	daysOfHistory := time.Hour * 24 * 7

	slot := phase0.Slot(1)
	expiresAt := CalculateBlockExpiration(slot, defaultSecondsPerSlot, defaultSlotsPerEpoch, genesis, daysOfHistory)

	if expiresAt.Before(genesis.Add(time.Hour * 24)) {
		t.Errorf("Expected block to expire at %v, got %v", genesis.Add(time.Hour*24), expiresAt)
	}

	if expiresAt.After(genesis.Add(time.Hour * 24 * 8)) {
		t.Errorf("Expected block to expire at %v, got %v", genesis.Add(time.Hour*24*7), expiresAt)
	}
}

func TestExpireMultiple(t *testing.T) {
	t.Parallel()

	genesis, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	daysOfHistory := time.Hour * 24 * 7

	tests := []struct {
		slot   phase0.Slot
		expect string
	}{
		{0, "2020-01-08 00:00:00 +0000 UTC"},
		{1, "2020-01-08 00:00:12 +0000 UTC"},
		{2, "2020-01-08 00:00:24 +0000 UTC"},
		{3, "2020-01-08 00:00:36 +0000 UTC"},
		{4, "2020-01-08 00:00:48 +0000 UTC"},
		{5, "2020-01-08 00:01:00 +0000 UTC"},
		{15000, "2020-01-10 02:00:00 +0000 UTC"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.slot), func(t *testing.T) {
			t.Parallel()

			expiresAt := CalculateBlockExpiration(test.slot, defaultSecondsPerSlot, defaultSlotsPerEpoch, genesis, daysOfHistory)
			if expiresAt.String() != test.expect {
				t.Errorf("Expected %v, got %v", test.expect, expiresAt.String())
			}
		})
	}
}
