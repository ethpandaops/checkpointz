package ssz

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/state"

	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

type Encoder struct {
	customPreset bool
	dynssz       *dynssz.DynSsz
	spec         map[string]any
	specMtx      sync.Mutex
}

func NewEncoder(customPreset bool) *Encoder {
	return &Encoder{
		customPreset: customPreset,
	}
}

func (e *Encoder) getDynamicSSZ() *dynssz.DynSsz {
	e.specMtx.Lock()
	defer e.specMtx.Unlock()

	if e.dynssz == nil {
		e.dynssz = dynssz.NewDynSsz(e.spec)
	}

	return e.dynssz
}

func (e *Encoder) SetSpec(newSpec *state.Spec) {
	e.specMtx.Lock()
	defer e.specMtx.Unlock()

	e.spec = newSpec.FullSpec
	e.dynssz = nil
}

func (e *Encoder) GetBlockRoot(block *spec.VersionedSignedBeaconBlock) (root phase0.Root, err error) {
	var blockObj sszutils.FastsszHashRoot

	switch block.Version {
	case spec.DataVersionPhase0:
		blockObj = block.Phase0.Message
	case spec.DataVersionAltair:
		blockObj = block.Altair.Message
	case spec.DataVersionBellatrix:
		blockObj = block.Bellatrix.Message
	case spec.DataVersionCapella:
		blockObj = block.Capella.Message
	case spec.DataVersionDeneb:
		blockObj = block.Deneb.Message
	case spec.DataVersionElectra:
		blockObj = block.Electra.Message
	case spec.DataVersionFulu:
		blockObj = block.Fulu.Message
	default:
		return phase0.Root{}, errors.New("unknown block version")
	}

	if e.customPreset {
		root, err = e.getDynamicSSZ().HashTreeRoot(blockObj)
	} else {
		root, err = blockObj.HashTreeRoot()
	}

	if err != nil {
		return phase0.Root{}, err
	}

	return root, nil
}

func (e *Encoder) EncodeBlockSSZ(block *spec.VersionedSignedBeaconBlock) (ssz []byte, err error) {
	var blockObj sszutils.FastsszMarshaler

	switch block.Version {
	case spec.DataVersionPhase0:
		blockObj = block.Phase0
	case spec.DataVersionAltair:
		blockObj = block.Altair
	case spec.DataVersionBellatrix:
		blockObj = block.Bellatrix
	case spec.DataVersionCapella:
		blockObj = block.Capella
	case spec.DataVersionDeneb:
		blockObj = block.Deneb
	case spec.DataVersionElectra:
		blockObj = block.Electra
	case spec.DataVersionFulu:
		blockObj = block.Fulu
	default:
		return nil, errors.New("unknown block version")
	}

	if e.customPreset {
		ssz, err = e.getDynamicSSZ().MarshalSSZ(blockObj)
	} else {
		ssz, err = blockObj.MarshalSSZ()
	}

	if err != nil {
		return nil, err
	}

	return ssz, nil
}

func (e *Encoder) EncodeBlockJSON(block *spec.VersionedSignedBeaconBlock) ([]byte, error) {
	var blockObj json.Marshaler

	switch block.Version {
	case spec.DataVersionPhase0:
		blockObj = block.Phase0
	case spec.DataVersionAltair:
		blockObj = block.Altair
	case spec.DataVersionBellatrix:
		blockObj = block.Bellatrix
	case spec.DataVersionCapella:
		blockObj = block.Capella
	case spec.DataVersionDeneb:
		blockObj = block.Deneb
	case spec.DataVersionElectra:
		blockObj = block.Electra
	case spec.DataVersionFulu:
		blockObj = block.Fulu
	default:
		return nil, errors.New("unknown block version")
	}

	ssz, err := blockObj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return ssz, nil
}

func (e *Encoder) GetStateRoot(beaconState *spec.VersionedBeaconState) (root phase0.Root, err error) {
	var stateObj sszutils.FastsszHashRoot

	switch beaconState.Version {
	case spec.DataVersionPhase0:
		stateObj = beaconState.Phase0
	case spec.DataVersionAltair:
		stateObj = beaconState.Altair
	case spec.DataVersionBellatrix:
		stateObj = beaconState.Bellatrix
	case spec.DataVersionCapella:
		stateObj = beaconState.Capella
	case spec.DataVersionDeneb:
		stateObj = beaconState.Deneb
	case spec.DataVersionElectra:
		stateObj = beaconState.Electra
	case spec.DataVersionFulu:
		stateObj = beaconState.Fulu
	default:
		return phase0.Root{}, errors.New("unknown state version")
	}

	if e.customPreset {
		root, err = e.getDynamicSSZ().HashTreeRoot(stateObj)
	} else {
		root, err = stateObj.HashTreeRoot()
	}

	if err != nil {
		return phase0.Root{}, err
	}

	return root, nil
}
func (e *Encoder) EncodeStateSSZ(beaconState *spec.VersionedBeaconState) (ssz []byte, err error) {
	var stateObj sszutils.FastsszMarshaler

	switch beaconState.Version {
	case spec.DataVersionPhase0:
		stateObj = beaconState.Phase0
	case spec.DataVersionAltair:
		stateObj = beaconState.Altair
	case spec.DataVersionBellatrix:
		stateObj = beaconState.Bellatrix
	case spec.DataVersionCapella:
		stateObj = beaconState.Capella
	case spec.DataVersionDeneb:
		stateObj = beaconState.Deneb
	case spec.DataVersionElectra:
		stateObj = beaconState.Electra
	case spec.DataVersionFulu:
		stateObj = beaconState.Fulu
	default:
		return nil, errors.New("unknown state version")
	}

	if e.customPreset {
		ssz, err = e.getDynamicSSZ().MarshalSSZ(stateObj)
	} else {
		ssz, err = stateObj.MarshalSSZ()
	}

	if err != nil {
		return nil, err
	}

	return ssz, nil
}
