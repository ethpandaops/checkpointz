package eth

import (
	"context"
	"fmt"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/sirupsen/logrus"
)

// Handler is the Eth Service handler. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Handler struct {
	log      logrus.FieldLogger
	provider beacon.FinalityProvider

	metrics *Metrics
}

// NewHandler returns a new Handler instance.
func NewHandler(log logrus.FieldLogger, beac beacon.FinalityProvider, namespace string) *Handler {
	return &Handler{
		log:      log.WithField("module", "service/eth"),
		provider: beac,

		metrics: NewMetrics(namespace),
	}
}

// BeaconBlock returns the beacon block for the given block ID.
func (h *Handler) BeaconBlock(ctx context.Context, blockID BlockIdentifier) (*spec.VersionedSignedBeaconBlock, error) {
	var err error

	const call = "beacon_block"

	h.metrics.ObserveCall(call, blockID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, blockID.Type().String())
		}
	}()

	switch blockID.Type() {
	case BlockIDGenesis:
		return h.provider.GetBlockBySlot(ctx, phase0.Slot(0))
	case BlockIDSlot:
		slot, err := NewSlotFromString(blockID.Value())
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockBySlot(ctx, slot)
	case BlockIDRoot:
		root, err := blockID.AsRoot()
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockByRoot(ctx, root)
	case BlockIDFinalized:
		finality, err := h.provider.Finality(ctx)
		if err != nil {
			return nil, err
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finality")
		}

		return h.provider.GetBlockByRoot(ctx, finality.Finalized.Root)
	default:
		return nil, fmt.Errorf("invalid block id: %v", blockID.String())
	}
}

// BeaconBlock returns the beacon state for the given state id.
func (h *Handler) BeaconState(ctx context.Context, stateID StateIdentifier) (*[]byte, error) {
	var err error

	const call = "beacon_state"

	h.metrics.ObserveCall(call, stateID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, stateID.Type().String())
		}
	}()

	switch stateID.Type() {
	case StateIDSlot:
		slot, err := NewSlotFromString(stateID.Value())
		if err != nil {
			return nil, err
		}

		return h.provider.GetBeaconStateBySlot(ctx, slot)
	case StateIDRoot:
		root, err := stateID.AsRoot()
		if err != nil {
			return nil, err
		}

		return h.provider.GetBeaconStateByStateRoot(ctx, root)
	case StateIDFinalized:
		finality, err := h.provider.Finality(ctx)
		if err != nil {
			return nil, err
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finality known")
		}

		return h.provider.GetBeaconStateByRoot(ctx, finality.Finalized.Root)
	case StateIDGenesis:
		return h.provider.GetBeaconStateBySlot(ctx, phase0.Slot(0))
	default:
		return nil, fmt.Errorf("invalid state id: %v", stateID.String())
	}
}

// FinalityCheckpoints returns the finality checkpoints for the given state id.
func (h *Handler) FinalityCheckpoints(ctx context.Context, stateID StateIdentifier) (*v1.Finality, error) {
	var err error

	const call = "finality_checkpoints"

	h.metrics.ObserveCall(call, stateID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, stateID.Type().String())
		}
	}()

	switch stateID.Type() {
	case StateIDHead:
		finality, err := h.provider.Finality(ctx)
		if err != nil {
			return nil, err
		}

		if finality.Finalized == nil {
			return nil, fmt.Errorf("no finalized state known")
		}

		return finality, nil
	default:
		return nil, fmt.Errorf("invalid state id: %v", stateID.String())
	}
}

// BlockRoot returns the beacon block root for the given block ID.
func (h *Handler) BlockRoot(ctx context.Context, blockID BlockIdentifier) (phase0.Root, error) {
	var err error

	const call = "block_root"

	h.metrics.ObserveCall(call, blockID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, blockID.Type().String())
		}
	}()

	switch blockID.Type() {
	case BlockIDGenesis:
		block, err := h.provider.GetBlockBySlot(ctx, phase0.Slot(0))
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no genesis block")
		}

		return block.Root()
	case BlockIDSlot:
		slot, err := NewSlotFromString(blockID.Value())
		if err != nil {
			return phase0.Root{}, err
		}

		block, err := h.provider.GetBlockBySlot(ctx, slot)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for slot %v", slot)
		}

		return block.Root()
	case BlockIDRoot:
		root, err := blockID.AsRoot()
		if err != nil {
			return phase0.Root{}, err
		}

		block, err := h.provider.GetBlockByRoot(ctx, root)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for root %v", root)
		}

		return block.Root()
	case BlockIDFinalized:
		finality, err := h.provider.Finality(ctx)
		if err != nil {
			return phase0.Root{}, err
		}

		if finality == nil || finality.Finalized == nil {
			return phase0.Root{}, fmt.Errorf("no finality")
		}

		block, err := h.provider.GetBlockByRoot(ctx, finality.Finalized.Root)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for finalized root %v", finality.Finalized.Root)
		}

		return block.Root()
	default:
		return phase0.Root{}, fmt.Errorf("invalid block id: %v", blockID.String())
	}
}
