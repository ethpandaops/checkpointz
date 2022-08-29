package checkpointz

import (
	"context"

	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/samcm/checkpointz/pkg/eth"
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
func (h *Handler) V1BeaconSlots(ctx context.Context, req *BeaconSlotsRequest) (*BeaconSlotsResponse, error) {
	response := &BeaconSlotsResponse{}

	slots, err := h.provider.ListFinalizedSlots(ctx)
	if err != nil {
		return nil, err
	}

	response.Slots = []BeaconSlot{}

	for _, s := range slots {
		slot := BeaconSlot{
			Slot: s,
		}

		if block, err := h.provider.GetBlockBySlot(ctx, slot.Slot); err == nil {
			if blockRoot, err := block.Root(); err == nil {
				slot.BlockRoot = eth.RootAsString(blockRoot)
			}

			if stateRoot, err := block.StateRoot(); err == nil {
				slot.StateRoot = eth.RootAsString(stateRoot)
			}
		}

		if epoch, err := h.provider.GetEpochBySlot(ctx, slot.Slot); err == nil {
			slot.Epoch = epoch
		}

		response.Slots = append(response.Slots, slot)
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

	if epoch, err := h.provider.GetEpochBySlot(ctx, req.slot); err == nil {
		response.Epoch = epoch
	}

	return response, nil
}
