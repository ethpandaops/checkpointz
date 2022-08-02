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
func (h *Handler) BeaconBlock(ctx context.Context, req *BeaconBlockRequest) (*spec.VersionedSignedBeaconBlock, error) {
	switch req.BlockID.Type() {
	case Slot:
		slot, err := NewSlotFromString(req.BlockID.Value())
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockBySlot(ctx, slot)
	case Root:
		root, err := req.BlockID.AsRoot()
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockByRoot(ctx, root)
	default:
		return nil, fmt.Errorf("invalid block id type: %v", req.BlockID.Type())
	}
}
