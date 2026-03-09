package eth

import (
	"math"
	"math/big"
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
	slotOffset := time.Duration(0)

	if durationPerSlot > 0 {
		offset := new(big.Int).Mul(
			new(big.Int).SetUint64(uint64(slot)),
			big.NewInt(int64(durationPerSlot)),
		)

		if offset.IsInt64() {
			slotOffset = time.Duration(offset.Int64())
		} else {
			slotOffset = time.Duration(math.MaxInt64)
		}
	}

	slotStartTime := genesisTime.Add(slotOffset).UTC()

	return SlotTime{
		StartTime: slotStartTime,
		EndTime:   slotStartTime.Add(durationPerSlot),
	}
}
