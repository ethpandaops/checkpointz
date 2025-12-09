package ssz

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/state"
	"golang.org/x/sync/semaphore"

	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/pk910/dynamic-ssz/sszutils"
)

// sszBufferPool provides reusable buffers for SSZ encoding to reduce allocations.
// Beacon states can be 100s of MB, so reusing buffers significantly reduces GC pressure.
var sszBufferPool = sync.Pool{
	New: func() any {
		// Start with 1MB buffer, will grow as needed
		b := make([]byte, 0, 1024*1024)
		return &b
	},
}

type Encoder struct {
	customPreset bool
	dynssz       *dynssz.DynSsz
	spec         map[string]any
	specMtx      sync.Mutex

	// memorySem limits total memory used for concurrent SSZ encoding.
	// If nil, no limit is applied.
	memorySem *semaphore.Weighted
}

// NewEncoder creates a new SSZ encoder.
// If memoryBudget is > 0, limits concurrent SSZ encoding to that many bytes.
// If memoryBudget is <= 0, no limit is applied.
func NewEncoder(customPreset bool, memoryBudget int64) *Encoder {
	var sem *semaphore.Weighted
	if memoryBudget > 0 {
		sem = semaphore.NewWeighted(memoryBudget)
	}

	return &Encoder{
		customPreset: customPreset,
		memorySem:    sem,
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

// WriteStateSSZ encodes the beacon state as SSZ and writes directly to w.
// This uses a pooled buffer to reduce allocations for large states (~100s of MB).
// If the encoder was created with a memory budget, this method will block until
// sufficient memory is available and respects context cancellation.
// The memory budget is released immediately after encoding completes, allowing
// slow client connections to stream without holding the budget hostage.
// Returns the number of bytes written.
func (e *Encoder) WriteStateSSZ(ctx context.Context, w io.Writer, beaconState *spec.VersionedBeaconState) (int64, error) {
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
		return 0, errors.New("unknown state version")
	}

	// Get the size upfront - needed for both memory budgeting and buffer allocation
	size := stateObj.SizeSSZ()

	// If memory budget is configured, acquire memory from the semaphore.
	// Note: We release the semaphore after encoding, not after streaming.
	// This allows slow clients to receive data without holding the budget.
	if e.memorySem != nil {
		if err := e.memorySem.Acquire(ctx, int64(size)); err != nil {
			return 0, err
		}
	}

	// For custom presets, fall back to regular encoding (dynamic-ssz doesn't support MarshalSSZTo)
	if e.customPreset {
		data, err := e.getDynamicSSZ().MarshalSSZ(stateObj)

		// Release memory budget immediately after encoding (before streaming to client)
		if e.memorySem != nil {
			e.memorySem.Release(int64(size))
		}

		if err != nil {
			return 0, err
		}

		n, err := w.Write(data)

		return int64(n), err
	}

	// Acquire a pooled buffer
	bufPtr, ok := sszBufferPool.Get().(*[]byte)
	if !ok || bufPtr == nil {
		// Pool returned unexpected type, allocate fresh buffer
		b := make([]byte, 0, size)
		bufPtr = &b
	}

	buf := *bufPtr

	// Ensure buffer has enough capacity
	if cap(buf) < size {
		buf = make([]byte, 0, size)
	} else {
		buf = buf[:0]
	}

	// Marshal into the buffer
	data, err := stateObj.MarshalSSZTo(buf)

	// Release memory budget immediately after encoding (before streaming to client)
	// This allows slow clients to receive data without holding the budget hostage.
	if e.memorySem != nil {
		e.memorySem.Release(int64(size))
	}

	if err != nil {
		// Return buffer to pool even on error
		*bufPtr = buf
		sszBufferPool.Put(bufPtr)

		return 0, err
	}

	// Write to the output (this can take a long time for slow clients, but we've
	// already released the memory budget so other requests can proceed)
	n, err := w.Write(data)

	// Return buffer to pool
	*bufPtr = buf
	sszBufferPool.Put(bufPtr)

	return int64(n), err
}

// StateSizeSSZ returns the SSZ encoded size of the beacon state without encoding it.
// Useful for setting Content-Length headers before streaming.
func (e *Encoder) StateSizeSSZ(beaconState *spec.VersionedBeaconState) (int, error) {
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
		return 0, errors.New("unknown state version")
	}

	return stateObj.SizeSSZ(), nil
}
