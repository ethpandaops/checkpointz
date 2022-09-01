package eth

import (
	"reflect"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func TestCalculateSlotTime(t *testing.T) {
	genesisTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	durationPerSlot := time.Second * 12

	tests := []struct {
		name string
		slot phase0.Slot
		want SlotTime
	}{
		{
			name: "Test 1",
			slot: phase0.Slot(0),
			want: SlotTime{
				StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 1, 1, 0, 0, 12, 0, time.UTC),
			},
		},
		{
			name: "Test 2",
			slot: phase0.Slot(1),
			want: SlotTime{
				StartTime: time.Date(2020, 1, 1, 0, 0, 12, 0, time.UTC),
				EndTime:   time.Date(2020, 1, 1, 0, 0, 24, 0, time.UTC),
			},
		},
		{
			name: "Test 3",
			slot: phase0.Slot(100),
			want: SlotTime{
				StartTime: time.Date(2020, 1, 1, 0, 20, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 1, 1, 0, 20, 12, 0, time.UTC),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CalculateSlotTime(test.slot, genesisTime, durationPerSlot)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("CalculateSlotTime() = %v, want %v", got, test.want)
			}
		})
	}
}
