package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// FinalityProvider is a provider of finality information.
type FinalityProvider interface {
	// Start starts the provider.
	Start(ctx context.Context) error
	// StartAsync starts the provider in a goroutine.
	StartAsync(ctx context.Context)
	// Healthy returns true if the provider is healthy.
	Healthy(ctx context.Context) (bool, error)
	// Syncing returns true if the provider is syncing.
	Syncing(ctx context.Context) (bool, error)
	// Finality returns the finality.
	Finality(ctx context.Context) (*v1.Finality, error)
	// UpstreamsStatus returns the status of all the upstreams.
	UpstreamsStatus(ctx context.Context) (map[string]*UpstreamStatus, error)
	// GetBlockBySlot returns the block at the given slot.
	GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error)
	// GetBlockByRoot returns the block with the given root.
	GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error)
	// GetBlockByStateRoot returns the block with the given state root.
	GetBlockByStateRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error)
}
