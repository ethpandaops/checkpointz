package checkpointz

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/samcm/checkpointz/pkg/beacon"
)

type StatusResponse struct {
	Upstreams map[string]*beacon.UpstreamStatus `json:"upstreams"`
	Finality  *v1.Finality                      `json:"finality"`
}

type BeaconSlotResponse struct {
	Block *spec.VersionedSignedBeaconBlock `json:"block"`
}
