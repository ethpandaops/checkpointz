package beacon_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

var (
	defaultSecondsPerSlot = time.Second * 12
)

func CalculateSlotExpiration(slot phase0.Slot, slotsOfHistory int) phase0.Slot {
	return slot + phase0.Slot(slotsOfHistory)
}

func GetSlotTime(slot phase0.Slot, secondsPerSlot time.Duration, genesis time.Time) time.Time {
	return genesis.Add(time.Duration(slot) * secondsPerSlot)
}

func TestExpiresAtSlot(t *testing.T) {
	slotsOfHistory := int(50)

	slot := phase0.Slot(1)
	expiresAtSlot := CalculateSlotExpiration(slot, slotsOfHistory)

	if expiresAtSlot != phase0.Slot(51) {
		t.Errorf("CalculateSlotExpiration() = %v, want %v", expiresAtSlot, phase0.Slot(51))
	}
}

func TestGetSlotTimeGenesis(t *testing.T) {
	genesis, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	slot := phase0.Slot(0)
	slotTime := GetSlotTime(slot, defaultSecondsPerSlot, genesis)

	if slotTime != genesis {
		t.Errorf("GetSlotTime() = %v, want %v", slotTime, genesis)
	}
}

func TestGetSlotTimeNormal(t *testing.T) {
	genesis, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	slot := phase0.Slot(2)
	slotTime := GetSlotTime(slot, defaultSecondsPerSlot, genesis)

	if slotTime != genesis.Add(time.Duration(slot)*defaultSecondsPerSlot) {
		t.Errorf("GetSlotTime() = %v, want %v", slotTime, genesis)
	}
}

func TestExpireMultiple(t *testing.T) {
	t.Parallel()

	genesis, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	slotsOfHistory := int(5)

	tests := []struct {
		slot   phase0.Slot
		expect string
	}{
		{0, "2020-01-01 00:01:00 +0000 UTC"},
		{1, "2020-01-01 00:01:12 +0000 UTC"},
		{2, "2020-01-01 00:01:24 +0000 UTC"},
		{3, "2020-01-01 00:01:36 +0000 UTC"},
		{4, "2020-01-01 00:01:48 +0000 UTC"},
		{5, "2020-01-01 00:02:00 +0000 UTC"},
		{15000, "2020-01-03 02:01:00 +0000 UTC"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.slot), func(t *testing.T) {
			test := test

			t.Parallel()

			expirySlot := CalculateSlotExpiration(test.slot, slotsOfHistory)

			expiryTime := GetSlotTime(expirySlot, defaultSecondsPerSlot, genesis)

			if expiryTime.String() != test.expect {
				t.Errorf("Expected %v, got %v", test.expect, expiryTime.String())
			}
		})
	}
}
