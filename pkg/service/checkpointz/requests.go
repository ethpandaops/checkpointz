package checkpointz

import "github.com/attestantio/go-eth2-client/spec/phase0"

type StatusRequest struct {
}

func (r *StatusRequest) Validate() error {
	return nil
}

func NewStatusRequest() *StatusRequest {
	return &StatusRequest{}
}

type BeaconSlotsRequest struct {
}

func (r *BeaconSlotsRequest) Validate() error {
	return nil
}

func NewBeaconSlotsRequest() *BeaconSlotsRequest {
	return &BeaconSlotsRequest{}
}

type BeaconSlotRequest struct {
	slot phase0.Slot
}

func (r *BeaconSlotRequest) Validate() error {
	return nil
}

func NewBeaconSlotRequest(slot phase0.Slot) *BeaconSlotRequest {
	return &BeaconSlotRequest{
		slot: slot,
	}
}
