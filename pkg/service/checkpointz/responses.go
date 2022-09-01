package checkpointz

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/samcm/checkpointz/pkg/eth"
)

type StatusResponse struct {
	Upstreams     map[string]*beacon.UpstreamStatus `json:"upstreams"`
	Finality      *v1.Finality                      `json:"finality"`
	PublicURL     string                            `json:"public_url,omitempty"`
	BrandName     string                            `json:"brand_name,omitempty"`
	BrandImageURL string                            `json:"brand_image_url,omitempty"`
	Version       Version                           `json:"version"`
}

type Version struct {
	Full      string `json:"full"`
	Short     string `json:"short"`
	Release   string `json:"release"`
	GitCommit string `json:"git_commit"`
}

type BeaconSlot struct {
	Slot      phase0.Slot  `json:"slot"`
	BlockRoot string       `json:"block_root,omitempty"`
	StateRoot string       `json:"state_root,omitempty"`
	Epoch     phase0.Epoch `json:"epoch"`
	SlotTime  eth.SlotTime `json:"time"`
}

type BeaconSlotsResponse struct {
	Slots []BeaconSlot `json:"slots"`
}

type BeaconSlotResponse struct {
	Block    *spec.VersionedSignedBeaconBlock `json:"block"`
	Epoch    phase0.Epoch                     `json:"epoch"`
	SlotTime eth.SlotTime                     `json:"time"`
}
