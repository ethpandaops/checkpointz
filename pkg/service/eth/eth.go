package eth

import (
	"context"
	"fmt"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/api/types"
	"github.com/ethpandaops/beacon/pkg/beacon/state"
	"github.com/ethpandaops/checkpointz/pkg/beacon"
	"github.com/ethpandaops/checkpointz/pkg/version"
	"github.com/sirupsen/logrus"
)

// Handler is the Eth Service handler. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Handler struct {
	log      logrus.FieldLogger
	provider beacon.FinalityProvider

	metrics *Metrics
}

// NewHandler returns a new Handler instance.
func NewHandler(log logrus.FieldLogger, beac beacon.FinalityProvider, namespace string) *Handler {
	return &Handler{
		log:      log.WithField("module", "service/eth"),
		provider: beac,

		metrics: NewMetrics(namespace),
	}
}

// BeaconBlock returns the beacon block for the given block ID.
func (h *Handler) BeaconBlock(ctx context.Context, blockID BlockIdentifier) (*spec.VersionedSignedBeaconBlock, error) {
	var err error

	const call = "beacon_block"

	h.metrics.ObserveCall(call, blockID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, blockID.Type().String())
		}
	}()

	switch blockID.Type() {
	case BlockIDGenesis:
		return h.provider.GetBlockBySlot(ctx, phase0.Slot(0))
	case BlockIDSlot:
		slot, err := NewSlotFromString(blockID.Value())
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockBySlot(ctx, slot)
	case BlockIDRoot:
		root, err := blockID.AsRoot()
		if err != nil {
			return nil, err
		}

		return h.provider.GetBlockByRoot(ctx, root)
	case BlockIDFinalized:
		finality, err := h.provider.Finalized(ctx)
		if err != nil {
			return nil, err
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finality")
		}

		return h.provider.GetBlockByRoot(ctx, finality.Finalized.Root)
	default:
		return nil, fmt.Errorf("invalid block id: %v", blockID.String())
	}
}

// BeaconGenesis returns the details of the chain's genesis.
func (h *Handler) BeaconGenesis(ctx context.Context) (*v1.Genesis, error) {
	var err error

	const call = "beacon_genesis"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return h.provider.Genesis(ctx)
}

// ConfigSpec gets the spec configuration.
func (h *Handler) ConfigSpec(ctx context.Context) (*state.Spec, error) {
	var err error

	const call = "config_spec"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return h.provider.Spec()
}

// ForkSchedule returns the upcoming forks.
func (h *Handler) ForkSchedule(ctx context.Context) ([]*state.ScheduledFork, error) {
	var err error

	const call = "fork_schedule"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	sp, err := h.provider.Spec()
	if err != nil {
		return nil, err
	}

	schedule, err := sp.ForkEpochs.AsScheduledForks()
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

// DepositContract gets the Eth1 deposit address and chain ID
func (h *Handler) DepositContract(ctx context.Context) (*DepositContract, error) {
	var err error

	const call = "config_deposit_contract"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	sp, err := h.provider.Spec()
	if err != nil {
		return nil, err
	}

	return &DepositContract{
		Address: sp.DepositContractAddress,
		ChainID: fmt.Sprintf("%d", sp.DepositChainID),
	}, nil
}

// DepositContract gets the deposit snapshot at the finalized checkpoint.
func (h *Handler) DepositSnapshot(ctx context.Context) (*types.DepositSnapshot, error) {
	var err error

	const call = "beacon_deposit_snapshot"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	finality, err := h.provider.Finalized(ctx)
	if err != nil {
		return nil, err
	}

	if finality == nil || finality.Finalized == nil {
		return nil, fmt.Errorf("no finality known")
	}

	snapshot, err := h.provider.GetDepositSnapshot(ctx, finality.Finalized.Epoch)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// NodeSyncing returns the sync state of the beacon node.
func (h *Handler) NodeSyncing(ctx context.Context) (*v1.SyncState, error) {
	var err error

	const call = "node_syncing"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return h.provider.Syncing(ctx)
}

// NodeVersion returns the version of the beacon node.
func (h *Handler) NodeVersion(ctx context.Context) (string, error) {
	var err error

	const call = "node_version"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return version.FullVWithGOOS(), nil
}

// Peers returns the peers connected to the beacon node.
func (h *Handler) Peers(ctx context.Context) (types.Peers, error) {
	var err error

	const call = "node_peers"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return h.provider.Peers(ctx)
}

// PeerCount returns the amount of peers connected to the beacon node.
func (h *Handler) PeerCount(ctx context.Context) (uint64, error) {
	var err error

	const call = "node_peer_count"

	h.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, "")
		}
	}()

	return h.provider.PeerCount(ctx)
}

// BeaconState returns the beacon state for the given state id.
func (h *Handler) BeaconState(ctx context.Context, stateID StateIdentifier) (*spec.VersionedBeaconState, error) {
	var err error

	const call = "beacon_state"

	h.metrics.ObserveCall(call, stateID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, stateID.Type().String())
		}
	}()

	switch stateID.Type() {
	case StateIDSlot:
		slot, err := NewSlotFromString(stateID.Value())
		if err != nil {
			return nil, err
		}

		return h.provider.GetBeaconStateBySlot(ctx, slot)
	case StateIDRoot:
		root, err := stateID.AsRoot()
		if err != nil {
			return nil, err
		}

		return h.provider.GetBeaconStateByStateRoot(ctx, root)
	case StateIDFinalized:
		finality, err := h.provider.Finalized(ctx)
		if err != nil {
			return nil, err
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finality known")
		}

		return h.provider.GetBeaconStateByRoot(ctx, finality.Finalized.Root)
	case StateIDGenesis:
		return h.provider.GetBeaconStateBySlot(ctx, phase0.Slot(0))
	default:
		return nil, fmt.Errorf("invalid state id: %v", stateID.String())
	}
}

// FinalityCheckpoints returns the finality checkpoints for the given state id.
func (h *Handler) FinalityCheckpoints(ctx context.Context, stateID StateIdentifier) (*v1.Finality, error) {
	var err error

	const call = "finality_checkpoints"

	h.metrics.ObserveCall(call, stateID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, stateID.Type().String())
		}
	}()

	switch stateID.Type() {
	case StateIDHead:
		finality, err := h.provider.Head(ctx)
		if err != nil {
			return nil, err
		}

		if finality.Finalized == nil {
			return nil, fmt.Errorf("no finalized state known")
		}

		return finality, nil
	case StateIDFinalized:
		finality, err := h.provider.Finalized(ctx)
		if err != nil {
			return nil, err
		}

		if finality.Finalized == nil {
			return nil, fmt.Errorf("no finalized state known")
		}

		return finality, nil
	default:
		return nil, fmt.Errorf("invalid state id: %v", stateID.String())
	}
}

// BlockRoot returns the beacon block root for the given block ID.
func (h *Handler) BlockRoot(ctx context.Context, blockID BlockIdentifier) (phase0.Root, error) {
	var err error

	const call = "block_root"

	h.metrics.ObserveCall(call, blockID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, blockID.Type().String())
		}
	}()

	switch blockID.Type() {
	case BlockIDGenesis:
		block, err := h.provider.GetBlockBySlot(ctx, phase0.Slot(0))
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no genesis block")
		}

		return block.Root()
	case BlockIDSlot:
		slot, err := NewSlotFromString(blockID.Value())
		if err != nil {
			return phase0.Root{}, err
		}

		block, err := h.provider.GetBlockBySlot(ctx, slot)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for slot %v", slot)
		}

		return block.Root()
	case BlockIDRoot:
		root, err := blockID.AsRoot()
		if err != nil {
			return phase0.Root{}, err
		}

		block, err := h.provider.GetBlockByRoot(ctx, root)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for root %v", root)
		}

		return block.Root()
	case BlockIDFinalized:
		finality, err := h.provider.Finalized(ctx)
		if err != nil {
			return phase0.Root{}, err
		}

		if finality == nil || finality.Finalized == nil {
			return phase0.Root{}, fmt.Errorf("no finality")
		}

		block, err := h.provider.GetBlockByRoot(ctx, finality.Finalized.Root)
		if err != nil {
			return phase0.Root{}, err
		}

		if block == nil {
			return phase0.Root{}, fmt.Errorf("no block for finalized root %v", finality.Finalized.Root)
		}

		return block.Root()
	default:
		return phase0.Root{}, fmt.Errorf("invalid block id: %v", blockID.String())
	}
}

// BlobSidecars returns the blob sidecars for the given block ID.
func (h *Handler) BlobSidecars(ctx context.Context, blockID BlockIdentifier, indices []int) ([]*deneb.BlobSidecar, error) {
	var err error

	const call = "blob_sidecars"

	h.metrics.ObserveCall(call, blockID.Type().String())

	defer func() {
		if err != nil {
			h.metrics.ObserveErrorCall(call, blockID.Type().String())
		}
	}()

	slot := phase0.Slot(0)

	switch blockID.Type() {
	case BlockIDGenesis:
		//nolint:govet // False positive
		block, err := h.provider.GetBlockBySlot(ctx, phase0.Slot(0))
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("no genesis block")
		}

		sl, err := block.Slot()
		if err != nil {
			return nil, err
		}

		slot = sl
	case BlockIDSlot:
		//nolint:govet // False positive
		sslot, err := NewSlotFromString(blockID.Value())
		if err != nil {
			return nil, err
		}

		block, err := h.provider.GetBlockBySlot(ctx, sslot)
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("no block for slot %v", sslot)
		}

		sl, err := block.Slot()
		if err != nil {
			return nil, err
		}

		slot = sl
	case BlockIDRoot:
		//nolint:govet // False positive
		root, err := blockID.AsRoot()
		if err != nil {
			return nil, err
		}

		block, err := h.provider.GetBlockByRoot(ctx, root)
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("no block for root %v", root)
		}

		sl, err := block.Slot()
		if err != nil {
			return nil, err
		}

		slot = sl
	case BlockIDFinalized:
		//nolint:govet // False positive
		finality, err := h.provider.Finalized(ctx)
		if err != nil {
			return nil, err
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finality")
		}

		block, err := h.provider.GetBlockByRoot(ctx, finality.Finalized.Root)
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, fmt.Errorf("no block for finalized root %v", finality.Finalized.Root)
		}

		sl, err := block.Slot()
		if err != nil {
			return nil, err
		}

		slot = sl
	default:
		return nil, fmt.Errorf("invalid block id: %v", blockID.String())
	}

	sidecars, err := h.provider.GetBlobSidecarsBySlot(ctx, slot)
	if err != nil {
		return nil, err
	}

	if len(indices) == 0 {
		return sidecars, nil
	}

	filtered := make([]*deneb.BlobSidecar, 0, len(indices))

	for _, index := range indices {
		if index < 0 {
			return nil, fmt.Errorf("invalid index %v", index)
		}

		// Find the sidecar with the given index
		for i, sidecar := range sidecars {
			if index == int(sidecar.Index) {
				filtered = append(filtered, sidecars[i])

				break
			}
		}
	}

	return filtered, nil
}
