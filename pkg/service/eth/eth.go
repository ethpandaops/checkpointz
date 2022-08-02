package eth

import (
	"context"
	"fmt"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/sirupsen/logrus"
)

// Handler is the Eth API handler. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Handler struct {
	log      logrus.FieldLogger
	provider beacon.FinalityProvider
}

// NewHandler returns a new Handler instance.
func NewHandler(log logrus.FieldLogger, beac beacon.FinalityProvider) *Handler {
	return &Handler{
		log:      log.WithField("module", "api/eth"),
		provider: beac,
	}
}

// BeaconBlock returns the beacon block for the given block ID.
func (h *Handler) BeaconBlock(ctx context.Context, blockID BlockIdentifier) (*spec.VersionedSignedBeaconBlock, error) {
	switch blockID.Type() {
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
	default:
		return nil, fmt.Errorf("invalid block id type: %v", blockID.Type())
	}
}

// BeaconBlock returns the beacon state for the given state id.
func (h *Handler) BeaconState(ctx context.Context, stateID StateIdentifier) (*spec.VersionedBeaconState, error) {
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

		if finality == nil {
			return nil, fmt.Errorf("no finality known")
		}

		return h.provider.GetBeaconStateByStateRoot(ctx, finality.Finalized.Root)
	default:
		return nil, fmt.Errorf("invalid state id type: %v", stateID.Type())
	}
}
