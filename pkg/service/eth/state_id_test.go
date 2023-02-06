package eth

import "testing"

func TestStateIDMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want StateIDType
	}{
		{"aaa", StateIDInvalid},
		{"head", StateIDHead},
		{"genesis", StateIDGenesis},
		{"finalized", StateIDFinalized},
		{"100", StateIDSlot},
		{"0x4a74943698817939e32aa6b2c688ccf1336bbff9190e400cc1360013d635da59", StateIDRoot},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			test := test

			t.Parallel()

			got, err := NewStateIdentifier(test.id)
			if err != nil {
				t.Fatal(err)
			}

			if got.Type() != test.want {
				t.Errorf("StateIDFromString() = %v, want %v", got, test.want)
			}
		})
	}
}
