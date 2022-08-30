package checkpointz

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/beacon"
)

type StatusResponse struct {
	Upstreams map[string]*beacon.UpstreamStatus `json:"upstreams"`
	Finality  *v1.Finality                      `json:"finality"`
	PublicURL string                            `json:"public_url,omitempty"`
}

type BeaconSlot struct {
	Slot      phase0.Slot  `json:"slot"`
	BlockRoot string       `json:"block_root,omitempty"`
	StateRoot string       `json:"state_root,omitempty"`
	Epoch     phase0.Epoch `json:"epoch"`
}

type BeaconSlotsResponse struct {
	Slots []BeaconSlot `json:"slots"`
}

type BeaconSlotResponse struct {
	Block *spec.VersionedSignedBeaconBlock `json:"block"`
	Epoch phase0.Epoch                     `json:"epoch"`
}
