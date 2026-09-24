package checkpoints

import (
	"github.com/ethpandaops/checkpointz/pkg/beacon/checkpoints/majority"
	v1 "github.com/ethpandaops/go-eth2-client/api/v1"
)

type Decider interface {
	Decide(checkpoints []*v1.Finality) (*v1.Finality, error)
}

var _ Decider = (*majority.Decider)(nil)

func NewMajorityDecider() *majority.Decider {
	return &majority.Decider{}
}
