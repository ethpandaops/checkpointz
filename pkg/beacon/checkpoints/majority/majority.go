package majority

import (
	"errors"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethpandaops/checkpointz/pkg/eth"
)

type Decider struct{}

var (
	ErrNoMajorityFound = errors.New("no majority finality found")
)

func New() *Decider {
	return &Decider{}
}

func (m *Decider) Decide(checkpoints []*v1.Finality) (*v1.Finality, error) {
	// Checkpoints with missing fields can't be keyed and don't count towards
	// the majority threshold, matching how upstreams that fail to return
	// finality at all are treated.
	valid := make([]*v1.Finality, 0, len(checkpoints))

	for _, checkpoint := range checkpoints {
		if checkpoint == nil ||
			checkpoint.Finalized == nil ||
			checkpoint.Justified == nil ||
			checkpoint.PreviousJustified == nil {
			continue
		}

		valid = append(valid, checkpoint)
	}

	common := make(map[string]struct {
		Finality *v1.Finality
		Count    int
	}, len(valid))

	for _, checkpoint := range valid {
		key := eth.RootAsString(checkpoint.Finalized.Root) + "-" +
			eth.RootAsString(checkpoint.Justified.Root) + "-" +
			eth.RootAsString(checkpoint.PreviousJustified.Root)

		if _, exists := common[key]; !exists {
			common[key] = struct {
				Finality *v1.Finality
				Count    int
			}{
				Finality: checkpoint,
				Count:    0,
			}
		}

		val, exists := common[key]
		if exists {
			val.Count++
			common[key] = val
		}
	}

	for _, v := range common {
		if v.Count > len(valid)/2 {
			return v.Finality, nil
		}
	}

	return nil, ErrNoMajorityFound
}
