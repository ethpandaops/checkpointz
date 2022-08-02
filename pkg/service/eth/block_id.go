package eth

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type BlockIDType int

const (
	Invalid BlockIDType = iota
	Head
	Genesis
	Finalized
	Slot
	Root
)

type BlockIdentifier struct {
	t BlockIDType
	v string
}

func (id BlockIdentifier) String() string {
	return id.v
}

func (id BlockIdentifier) Type() BlockIDType {
	return id.t
}

func (id BlockIdentifier) Value() string {
	return id.v
}

func (id BlockIdentifier) AsRoot() (phase0.Root, error) {
	if id.t != Root {
		return phase0.Root{}, fmt.Errorf("invalid block ID type %d", id.t)
	}

	return NewRootFromString(id.v)
}

func (id BlockIdentifier) AsSlot() (phase0.Slot, error) {
	if id.t != Slot {
		return phase0.Slot(0), fmt.Errorf("invalid block ID type %d", id.t)
	}

	return NewSlotFromString(id.v)
}

func NewBlockIdentifier(id string) (BlockIdentifier, error) {
	switch id {
	case "head":
		return newBlockIdentifier(Head, id), nil
	case "genesis":
		return newBlockIdentifier(Genesis, id), nil
	case "finalized":
		return newBlockIdentifier(Finalized, id), nil
	}

	if strings.HasPrefix(id, "0x") {
		return newBlockIdentifier(Root, id), nil
	}

	if _, err := strconv.ParseInt(id, 10, 64); err == nil {
		return newBlockIdentifier(Slot, id), nil
	}

	return newBlockIdentifier(Invalid, id), fmt.Errorf("invalid block ID: %s", id)
}

func newBlockIdentifier(id BlockIDType, value string) BlockIdentifier {
	return BlockIdentifier{
		t: id,
		v: value,
	}
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
