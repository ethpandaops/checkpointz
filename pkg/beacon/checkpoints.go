package beacon

import (
	"errors"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type Checkpoints struct {
	Finalized         Checkpoint
	Justified         Checkpoint
	PreviousJustified Checkpoint
}

type Checkpoint map[phase0.Epoch]map[phase0.Root]int

func (c Checkpoint) Add(epoch phase0.Epoch, root phase0.Root) {
	if _, ok := c[epoch]; !ok {
		c[epoch] = make(map[phase0.Root]int)
	}

	if _, ok := c[epoch][root]; !ok {
		c[epoch][root] = 1
	}

	c[epoch][root]++
}

func NewCheckpoints(finalities []*v1.Finality) *Checkpoints {
	checkpoints := &Checkpoints{
		Finalized:         make(Checkpoint),
		Justified:         make(Checkpoint),
		PreviousJustified: make(Checkpoint),
	}

	for _, finality := range finalities {
		checkpoints.Finalized.Add(finality.Finalized.Epoch, finality.Finalized.Root)
		checkpoints.Justified.Add(finality.Justified.Epoch, finality.Justified.Root)
		checkpoints.PreviousJustified.Add(finality.PreviousJustified.Epoch, finality.PreviousJustified.Root)
	}

	return checkpoints
}

// Majority finds the majority of the checkpoints.
// TODO(sam.calder-mason): Is it safe to disconnect a bundle of checkpoints like this?
func (c Checkpoints) Majority() (*v1.Finality, error) {
	finality := &v1.Finality{}

	finalizedCount := -1

	for epoch, epochMap := range c.Finalized {
		for root, count := range epochMap {
			if count > finalizedCount {
				finality.Finalized = &phase0.Checkpoint{
					Epoch: epoch,
					Root:  root,
				}
				finalizedCount = count
			}
		}
	}

	justifiedCount := -1

	for epoch, epochMap := range c.Justified {
		for root, count := range epochMap {
			if count > justifiedCount {
				finality.Justified = &phase0.Checkpoint{
					Epoch: epoch,
					Root:  root,
				}
				justifiedCount = count
			}
		}
	}

	previousJustifiedCount := -1

	for epoch, epochMap := range c.Justified {
		for root, count := range epochMap {
			if count > previousJustifiedCount {
				finality.PreviousJustified = &phase0.Checkpoint{
					Epoch: epoch,
					Root:  root,
				}
				previousJustifiedCount = count
			}
		}
	}

	if finalizedCount == -1 {
		return finality, errors.New("no finality")
	}

	if finalizedCount <= len(c.Finalized)/2 {
		return finality, errors.New("unable to determine majority")
	}

	return finality, nil
}
