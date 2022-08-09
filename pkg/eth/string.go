package eth

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func RootAsString(root phase0.Root) string {
	return fmt.Sprintf("%#x", root)
}

func SlotAsString(slot phase0.Slot) string {
	return fmt.Sprintf("%d", slot)
}
