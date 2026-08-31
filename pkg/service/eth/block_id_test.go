package eth

import "testing"

func TestBlockIDMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id     string
		expect BlockIDType
	}{
		{string(IDHead), BlockIDHead},
		{string(IDGenesis), BlockIDGenesis},
		{string(IDFinalized), BlockIDFinalized},
		{"10", BlockIDSlot},
		{"0x4a74943698817939e32aa6b2c688ccf1336bbff9190e400cc1360013d635da59", BlockIDRoot},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			test := test

			t.Parallel()

			if id, err := NewBlockIdentifier(test.id); err != nil {
				t.Fatal(err)
			} else if id.Type() != test.expect {
				t.Errorf("Expected %d, got %d", test.expect, id.Type())
			}
		})
	}
}

func TestBlockIDRejectsNegativeSlot(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"-1", "-100"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			parsed, err := NewBlockIdentifier(id)
			if err == nil {
				t.Errorf("Expected error for %q, got type %d", id, parsed.Type())
			}

			if parsed.Type() != BlockIDInvalid {
				t.Errorf("Expected BlockIDInvalid for %q, got %d", id, parsed.Type())
			}
		})
	}
}

func TestNewSlotFromStringRejectsNegative(t *testing.T) {
	t.Parallel()

	if _, err := NewSlotFromString("-1"); err == nil {
		t.Error("Expected error for negative slot, got nil")
	}
}
