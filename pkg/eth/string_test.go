package eth

import (
	"testing"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func TestRootToString(t *testing.T) {
	tests := []struct {
		name string
		root phase0.Root
		want string
	}{
		{
			name: "Test 1",
			root: phase0.Root{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
			want: "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
		},
		{
			name: "Test 2",
			root: phase0.Root{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RootAsString(test.root)
			if got != test.want {
				t.Errorf("RootAsString() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSlotToString(t *testing.T) {
	tests := []struct {
		name string
		slot phase0.Slot
		want string
	}{
		{
			name: "Test 1",
			slot: phase0.Slot(0),
			want: "0",
		},
		{
			name: "Test 2",
			slot: phase0.Slot(1000),
			want: "1000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SlotAsString(test.slot)
			if got != test.want {
				t.Errorf("SlotAsString() = %v, want %v", got, test.want)
			}
		})
	}
}
