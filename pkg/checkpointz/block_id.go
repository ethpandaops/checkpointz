package checkpointz

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type BlockIdentifier int

const (
	Invalid BlockIdentifier = iota
	Head
	Genesis
	Finalized
	Slot
	Root
)

func ParseBlockID(id string) (BlockIdentifier, error) {
	switch id {
	case "head":
		return Head, nil
	case "genesis":
		return Genesis, nil
	case "finalized":
		return Finalized, nil
	}

	if strings.HasPrefix(id, "0x") {
		return Root, nil
	}

	if _, err := strconv.ParseInt(id, 10, 64); err == nil {
		return Slot, nil
	}

	return Invalid, fmt.Errorf("invalid block ID: %s", id)
}

func NewSlotFromString(id string) (phase0.Slot, error) {
	slot, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	}

	return phase0.Slot(slot), nil
}

func NewRootFromString(id string) (phase0.Root, error) {
	b, err := hex.DecodeString(strings.TrimPrefix(id, "0x"))
	if err != nil {
		return phase0.Root{}, errors.Wrap(err, "invalid value for root")
	}

	root := phase0.Root{}

	if len(b) != len(root) {
		return phase0.Root{}, fmt.Errorf("incorrect length %d for root", len(b))
	}

	copy(root[:], b)

	return root, nil
}
