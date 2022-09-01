package beacon

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func CalculateSlotExpiration(slot phase0.Slot, slotsOfHistory int) phase0.Slot {
	return slot + phase0.Slot(slotsOfHistory)
}

func GetSlotTime(slot phase0.Slot, secondsPerSlot time.Duration, genesis time.Time) time.Time {
	return genesis.Add(time.Duration(slot) * secondsPerSlot)
}
