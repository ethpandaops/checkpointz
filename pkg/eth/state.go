package eth

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func GetSlotFromState(state *spec.VersionedBeaconState) (phase0.Slot, error) {
	if state == nil {
		return 0, fmt.Errorf("state is nil")
	}

	if state.Version == spec.DataVersionPhase0 {
		return phase0.Slot(state.Phase0.Slot), nil
	}

	if state.Version == spec.DataVersionAltair {
		return phase0.Slot(state.Altair.Slot), nil
	}

	if state.Version == spec.DataVersionBellatrix {
		return phase0.Slot(state.Bellatrix.Slot), nil
	}

	return 0, nil
}

// func GetStateHashTreeRoot(state *spec.VersionedBeaconState) (phase0.Root, error) {
// 	if state == nil {
// 		return 0, fmt.Errorf("state is nil")
// 	}

// 	if state.Version == spec.DataVersionPhase0 {
// 		return phase0.Slot(state.Phase0.Slot), nil
// 	}

// 	if state.Version == spec.DataVersionAltair {
// 		return phase0.Slot(state.Altair.Slot), nil
// 	}

// 	if state.Version == spec.DataVersionBellatrix {
// 		return phase0.Slot(state.Bellatrix.Slot), nil
// 	}

// 	return 0, nil
// }
