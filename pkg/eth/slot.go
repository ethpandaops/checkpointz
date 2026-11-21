package eth

import (
	"time"
)

type SlotTime struct {
	// The time at which the slot started.
	StartTime time.Time `json:"start_time"`
	// The time at which the slot ends.
	EndTime time.Time `json:"end_time"`
}
