package beacon

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func CalculateBlockExpiration(slot phase0.Slot, secondsPerSlot time.Duration, slotsPerEpoch uint64, genesis time.Time, historyDuration time.Duration) time.Time {
	// Calculate the wall clock for when the block was created
	createdAt := genesis.Add(time.Duration(uint64(slot)) * secondsPerSlot)

	// Add our configured block history days to the createdAt time
	return createdAt.Add(historyDuration)
}
