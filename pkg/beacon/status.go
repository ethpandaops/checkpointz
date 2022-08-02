package beacon

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
)

type UpstreamStatus struct {
	Name     string       `json:"name"`
	Healthy  bool         `json:"healthy"`
	Finality *v1.Finality `json:"finality"`
}
