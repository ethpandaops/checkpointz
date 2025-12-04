package ssz

import (
	"errors"
	"sync"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/state"

	dynssz "github.com/pk910/dynamic-ssz"
)

type Encoder struct {
	dynssz  *dynssz.DynSsz
	spec    map[string]any
	specMtx sync.Mutex
}

func NewEncoder() *Encoder {
	return &Encoder{}
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

func (e *Encoder) GetBlockRoot(block *spec.VersionedSignedBeaconBlock) (phase0.Root, error) {
	ds := e.getDynamicSSZ()

	var blockObj any

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

	root, err := ds.HashTreeRoot(blockObj)
	if err != nil {
		return phase0.Root{}, err
	}

	return root, nil
}

func (e *Encoder) EncodeBlockSSZ(block *spec.VersionedSignedBeaconBlock) ([]byte, error) {
	ds := e.getDynamicSSZ()

	var blockObj any

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

	ssz, err := ds.MarshalSSZ(blockObj)
	if err != nil {
		return nil, err
	}

	return ssz, nil
}

type blockJsonWriter interface {
	MarshalJSON() ([]byte, error)
}

func (e *Encoder) EncodeBlockJSON(block *spec.VersionedSignedBeaconBlock) ([]byte, error) {
	var blockObj blockJsonWriter

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

func (e *Encoder) GetStateRoot(beaconState *spec.VersionedBeaconState) (phase0.Root, error) {
	ds := e.getDynamicSSZ()

	var stateObj any

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

	root, err := ds.HashTreeRoot(stateObj)
	if err != nil {
		return phase0.Root{}, err
	}

	return root, nil
}

func (e *Encoder) EncodeStateSSZ(beaconState *spec.VersionedBeaconState) ([]byte, error) {
	ds := e.getDynamicSSZ()

	var stateObj any

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

	ssz, err := ds.MarshalSSZ(stateObj)
	if err != nil {
		return nil, err
	}

	return ssz, nil
}
