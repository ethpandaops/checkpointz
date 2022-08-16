package eth

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type StateIDType int

const (
	StateIDInvalid StateIDType = iota
	StateIDHead
	StateIDGenesis
	StateIDFinalized
	StateIDSlot
	StateIDRoot
)

type StateIdentifier struct {
	t StateIDType
	v string
}

func (id StateIdentifier) String() string {
	return id.v
}

func (id StateIdentifier) Type() StateIDType {
	return id.t
}

func (id StateIdentifier) Value() string {
	return id.v
}

func (id StateIdentifier) AsRoot() (phase0.Root, error) {
	if id.t != StateIDRoot {
		return phase0.Root{}, fmt.Errorf("invalid block ID type %d", id.t)
	}

	return NewRootFromString(id.v)
}

func (id StateIdentifier) AsSlot() (phase0.Slot, error) {
	if id.t != StateIDSlot {
		return phase0.Slot(0), fmt.Errorf("invalid block ID type %d", id.t)
	}

	return NewSlotFromString(id.v)
}

func NewStateIdentifier(id string) (StateIdentifier, error) {
	switch id {
	case "head":
		return newStateIdentifier(StateIDHead, id), nil
	case "genesis":
		return newStateIdentifier(StateIDGenesis, id), nil
	case "finalized":
		return newStateIdentifier(StateIDFinalized, id), nil
	}

	if strings.HasPrefix(id, "0x") {
		return newStateIdentifier(StateIDRoot, id), nil
	}

	if _, err := strconv.ParseInt(id, 10, 64); err == nil {
		return newStateIdentifier(StateIDSlot, id), nil
	}

	return newStateIdentifier(StateIDInvalid, id), fmt.Errorf("invalid state ID: %s", id)
}

func newStateIdentifier(id StateIDType, value string) StateIdentifier {
	return StateIdentifier{
		t: id,
		v: value,
	}
}

func (t StateIDType) String() string {
	switch t {
	case StateIDHead:
		return string(IDHead)
	case StateIDGenesis:
		return string(IDGenesis)
	case StateIDFinalized:
		return string(IDFinalized)
	case StateIDSlot:
		return string(IDSlot)
	case StateIDRoot:
		return string(IDRoot)
	}

	return string(IDInvalid)
}
