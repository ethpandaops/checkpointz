package eth

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type SlotTime struct {
	// The time at which the slot started.
	StartTime time.Time `json:"start_time"`
	// The time at which the slot ends.
	EndTime time.Time `json:"end_time"`
}

func CalculateSlotTime(slot phase0.Slot, genesisTime time.Time, durationPerSlot time.Duration) SlotTime {
	slotStartTime := genesisTime.Add(time.Duration(slot) * durationPerSlot).UTC()

	return SlotTime{
		StartTime: slotStartTime,
		EndTime:   slotStartTime.Add(durationPerSlot),
	}
}
