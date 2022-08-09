package eth

import "testing"

func TestBlockIDMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id     string
		expect BlockIDType
	}{
		{"aaa", BlockIDInvalid},
		{"head", BlockIDHead},
		{"genesis", BlockIDGenesis},
		{"finalized", BlockIDFinalized},
		{"10", BlockIDSlot},
		{"0x4a74943698817939e32aa6b2c688ccf1336bbff9190e400cc1360013d635da59", BlockIDRoot},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			t.Parallel()

			if id, err := NewBlockIdentifier(test.id); err != nil {
				t.Fatal(err)
			} else if id.Type() != test.expect {
				t.Errorf("Expected %d, got %d", test.expect, id.Type())
			}
		})
	}
}
