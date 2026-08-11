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

// Decide picks the checkpoint that a majority of the configured upstreams agree on.
//
// totalUpstreams is the number of configured upstreams eligible to vote, not the
// number of entries in checkpoints. Only nodes that are currently ready respond,
// so checkpoints can be a strict subset of the configured set; thresholding against
// len(checkpoints) instead of totalUpstreams would let a single responder (or any
// minority that happens to be the only one currently ready) declare itself a
// majority. Requiring count > totalUpstreams/2 means a majority can only ever be
// reached once more than half of the *configured* upstreams agree, and naturally
// requires at least two agreeing responses whenever more than one upstream is
// configured.
func (m *Decider) Decide(checkpoints []*v1.Finality, totalUpstreams int) (*v1.Finality, error) {
	common := make(map[string]struct {
		Finality *v1.Finality
		Count    int
	})

	for _, checkpoint := range checkpoints {
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
		if v.Count > totalUpstreams/2 {
			return v.Finality, nil
		}
	}

	return nil, ErrNoMajorityFound
}
