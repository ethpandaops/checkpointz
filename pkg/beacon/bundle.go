package beacon

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec"
)

type CheckpointBundle struct {
	block *spec.VersionedSignedBeaconBlock
	state *[]byte
}

func NewCheckpointBundle(block *spec.VersionedSignedBeaconBlock, state *[]byte) *CheckpointBundle {
	return &CheckpointBundle{
		block: block,
		state: state,
	}
}

func (c *CheckpointBundle) ExpiresAt() time.Time {
	// TODO(sam.calder-mason): Actually this by calculating the weak subjectivity period.
	// Never expire for now.
	return time.Now().Add(24 * time.Hour)
}

func (c *CheckpointBundle) Block() *spec.VersionedSignedBeaconBlock {
	return c.block
}

func (c *CheckpointBundle) State() *[]byte {
	return c.state
}
