package checkpointz

import (
	"context"

	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/sirupsen/logrus"
)

// Handler is the Checkpointz API handler. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Handler struct {
	log      logrus.FieldLogger
	provider beacon.FinalityProvider
}

// NewHandler returns a new Handler instance.
func NewHandler(log logrus.FieldLogger, beac beacon.FinalityProvider) *Handler {
	return &Handler{
		log:      log.WithField("module", "api/checkpointz"),
		provider: beac,
	}
}

// Status returns the status for checkpointz.
func (h *Handler) V1Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	response := &StatusResponse{}

	upstreams, err := h.provider.UpstreamsStatus(ctx)
	if err != nil {
		return nil, err
	}

	response.Upstreams = upstreams

	finality, err := h.provider.Finality(ctx)
	if err != nil {
		return nil, err
	}

	if finality != nil {
		response.Finality = finality
	}

	return response, nil
}

// Slot returns the beacon slot for checkpointz.
func (h *Handler) V1BeaconSlot(ctx context.Context, req *BeaconSlotRequest) (*BeaconSlotResponse, error) {
	response := &BeaconSlotResponse{}

	block, err := h.provider.GetBlockBySlot(ctx, req.slot)
	if err != nil {
		return nil, err
	}

	response.Block = block

	return response, nil
}
