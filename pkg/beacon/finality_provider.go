package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/samcm/beacon/api/types"
	"github.com/samcm/beacon/state"
)

// FinalityProvider is a provider of finality information.
type FinalityProvider interface {
	// Start starts the provider.
	Start(ctx context.Context) error
	// StartAsync starts the provider in a goroutine.
	StartAsync(ctx context.Context)
	// Healthy returns true if the provider is healthy.
	Healthy(ctx context.Context) (bool, error)
	// Peers returns the peers the provider is connected to).
	Peers(ctx context.Context) (types.Peers, error)
	// PeerCount returns the amount of peers the provider is connected to (the amount of healthy upstreams).
	PeerCount(ctx context.Context) (uint64, error)
	// Syncing returns the sync state of the provider.
	Syncing(ctx context.Context) (*v1.SyncState, error)
	// Head returns the head finality.
	Head(ctx context.Context) (*v1.Finality, error)
	// Finalized returns the finalized finality.
	Finalized(ctx context.Context) (*v1.Finality, error)
	// Genesis returns the chain genesis.
	Genesis(ctx context.Context) (*v1.Genesis, error)
	// Spec returns the chain spec.
	Spec(ctx context.Context) (*state.Spec, error)
	// UpstreamsStatus returns the status of all the upstreams.
	UpstreamsStatus(ctx context.Context) (map[string]*UpstreamStatus, error)
	// GetBlockBySlot returns the block at the given slot.
	GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error)
	// GetBlockByRoot returns the block with the given root.
	GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error)
	// GetBlockByStateRoot returns the block with the given root.
	GetBlockByStateRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error)
	// GetBeaconStateBySlot returns the beacon sate with the given slot.
	GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*[]byte, error)
	// GetBeaconStateByStateRoot returns the beacon sate with the given state root.
	GetBeaconStateByStateRoot(ctx context.Context, root phase0.Root) (*[]byte, error)
	// GetBeaconStateByRoot returns the beacon sate with the given root.
	GetBeaconStateByRoot(ctx context.Context, root phase0.Root) (*[]byte, error)
	// ListFinalizedSlots returns a slice of finalized slots.
	ListFinalizedSlots(ctx context.Context) ([]phase0.Slot, error)
	// GetEpochBySlot returns the epoch for the given slot.
	GetEpochBySlot(ctx context.Context, slot phase0.Slot) (phase0.Epoch, error)
	// OperatingMode returns the mode of operation for the instance.
	OperatingMode() OperatingMode
	// GetSlotTime returns the wall clock for the given slot.
	GetSlotTime(ctx context.Context, slot phase0.Slot) (eth.SlotTime, error)
}
