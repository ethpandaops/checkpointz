package beacon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/chuckpreslar/emission"
	"github.com/ethpandaops/beacon/pkg/beacon"
	"github.com/ethpandaops/beacon/pkg/beacon/api/types"
	"github.com/ethpandaops/beacon/pkg/beacon/state"
	"github.com/ethpandaops/checkpointz/pkg/beacon/checkpoints"
	"github.com/ethpandaops/checkpointz/pkg/beacon/node"
	"github.com/ethpandaops/checkpointz/pkg/beacon/ssz"
	"github.com/ethpandaops/checkpointz/pkg/beacon/store"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/ethpandaops/ethwallclock"
	"github.com/go-co-op/gocron"
	perrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Default struct {
	log logrus.FieldLogger

	config      *Config
	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter
	sszEncoder  *ssz.Encoder

	head          *v1.Finality
	servingBundle *v1.Finality

	blocks           *store.Block
	states           *store.BeaconState
	depositSnapshots *store.DepositSnapshot
	blobSidecars     *store.BlobSidecar

	specMutex sync.Mutex
	spec      *state.Spec
	genesis   *v1.Genesis

	historicalSlotFailures map[phase0.Slot]int

	servingMutex    sync.Mutex
	historicalMutex sync.Mutex
	majorityMutex   sync.Mutex

	metrics *Metrics
}

var _ FinalityProvider = (*Default)(nil)

var (
	topicFinalityHeadUpdated = "finality_head_updated"
)

const (
	// FinalityHaltedServingPeriod defines how long we will happily serve finality data for after the chain has stopped finality.
	// TODO(sam.calder-mason): Derive from weak subjectivity period.
	FinalityHaltedServingPeriod = 14 * 24 * time.Hour
)

func NewDefaultProvider(namespace string, log logrus.FieldLogger, nodes []node.Config, config *Config) FinalityProvider {
	return &Default{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/default"),
		nodes:       NewNodesFromConfig(log, nodes, namespace, config.CustomPreset),
		config:      config,

		head:          &v1.Finality{},
		servingBundle: &v1.Finality{},

		historicalSlotFailures: make(map[phase0.Slot]int),

		broker:           emission.NewEmitter(),
		sszEncoder:       ssz.NewEncoder(config.CustomPreset),
		blocks:           store.NewBlock(log, config.Caches.Blocks, namespace),
		states:           store.NewBeaconState(log, config.Caches.States, namespace),
		depositSnapshots: store.NewDepositSnapshot(log, config.Caches.DepositSnapshots, namespace),
		blobSidecars:     store.NewBlobSidecar(log, config.Caches.BlobSidecars, namespace),

		servingMutex:    sync.Mutex{},
		historicalMutex: sync.Mutex{},
		majorityMutex:   sync.Mutex{},
		specMutex:       sync.Mutex{},

		metrics: NewMetrics(namespace + "_beacon"),
	}
}

func (d *Default) Start(ctx context.Context) error {
	d.log.Infof("Starting Finality provider in %s mode", d.OperatingMode())

	d.metrics.ObserveOperatingMode(d.OperatingMode())

	if err := d.nodes.StartAll(ctx); err != nil {
		return err
	}

	go func() {
		for {
			// Wait until we have a single healthy node.
			nd, err := d.nodes.Healthy(ctx).NotSyncing(ctx).RandomNode(ctx)
			if err != nil {
				d.log.WithError(err).Error("Waiting for a healthy, non-syncing node before beginning..")
				time.Sleep(time.Second * 5)

				continue
			}

			nd.Beacon.Wallclock().OnEpochChanged(func(epoch ethwallclock.Epoch) {
				// Refresh the spec on epoch change.
				// This will intentionally use any node (not the one that triggered the event) to fetch the spec.
				if err := d.refreshSpec(ctx); err != nil {
					d.log.WithError(err).Error("Failed to refresh spec")
				}
			})

			if err := d.startCrons(ctx); err != nil {
				d.log.WithError(err).Fatal("Failed to start crons")
			}

			go func() {
				if err := d.startGenesisLoop(ctx); err != nil {
					d.log.WithError(err).Fatal("Failed to start genesis loop")
				}
			}()

			if err := d.fetchUpstreamRequirements(ctx); err != nil {
				d.log.WithError(err).Error("Failed to fetch upstream requirements")
			}

			break
		}
	}()

	// Subscribe to the nodes' finality updates.
	for _, node := range d.nodes {
		n := node

		logCtx := d.log.WithFields(logrus.Fields{
			"node":   n.Config.Name,
			"reason": "serving_updater",
		})

		n.Beacon.OnFinalityCheckpointUpdated(ctx, func(ctx context.Context, event *beacon.FinalityCheckpointUpdated) error {
			logCtx.WithFields(logrus.Fields{
				"epoch": event.Finality.Finalized.Epoch,
				"root":  fmt.Sprintf("%#x", event.Finality.Finalized.Root),
			}).Info("Node has a new finalized checkpoint")

			// Check if we have a new majority finality.
			if err := d.checkFinality(ctx); err != nil {
				logCtx.WithError(err).Error("Failed to check finality")

				return err
			}

			if err := d.checkForNewServingCheckpoint(ctx); err != nil {
				logCtx.WithError(err).Error("Failed to check for new serving checkpoint after finality checkpoint updated")

				return err
			}

			return nil
		})

		n.Beacon.OnReady(ctx, func(ctx context.Context, _ *beacon.ReadyEvent) error {
			n.Beacon.Wallclock().OnEpochChanged(func(epoch ethwallclock.Epoch) {
				time.Sleep(time.Second * 5)

				if _, err := node.Beacon.FetchFinality(ctx, "head"); err != nil {
					logCtx.WithError(err).Error("Failed to fetch finality after epoch transition")
				}

				if err := d.checkFinality(ctx); err != nil {
					logCtx.WithError(err).Error("Failed to check finality")
				}

				if err := d.checkForNewServingCheckpoint(ctx); err != nil {
					logCtx.WithError(err).Error("Failed to check for new serving checkpoint after epoch change")
				}
			})

			return nil
		})
	}

	return nil
}

func (d *Default) startCrons(ctx context.Context) error {
	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("30s").Do(func() {
		if err := d.checkFinality(ctx); err != nil {
			d.log.WithError(err).Error("Failed to check finality")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("10s").Do(func() {
		if err := d.checkBeaconSpec(ctx); err != nil {
			d.log.WithError(err).Error("Failed to check beacon chain spec")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("3m").Do(func() {
		for _, node := range d.nodes.Healthy(ctx) {
			if _, err := node.Beacon.FetchFinality(ctx, "head"); err != nil {
				d.log.WithError(err).Error("Failed to fetch finality when polling")
			}
		}
	}); err != nil {
		return err
	}

	go func() {
		if err := d.startServingLoop(ctx); err != nil {
			d.log.WithError(err).Fatal("Failed to start serving loop")
		}
	}()

	go func() {
		if err := d.startHistoricalLoop(ctx); err != nil {
			d.log.WithError(err).Fatal("Failed to start historical loop")
		}
	}()

	s.StartAsync()

	return nil
}

func (d *Default) fetchUpstreamRequirements(ctx context.Context) error {
	if err := d.refreshSpec(ctx); err != nil {
		return err
	}

	return nil
}

func (d *Default) StartAsync(ctx context.Context) {
	go func() {
		if err := d.Start(ctx); err != nil {
			d.log.WithError(err).Error("Failed to start")
		}
	}()
}

func (d *Default) startGenesisLoop(ctx context.Context) error {
	if err := d.checkGenesis(ctx); err != nil {
		d.log.WithError(err).Error("Failed to check for genesis bundle")
	}

	if err := d.checkGenesisTime(ctx); err != nil {
		d.log.WithError(err).Error("Failed to check genesis time")
	}

	for {
		select {
		case <-time.After(time.Second * 15):
			if err := d.checkGenesisTime(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check genesis time")
			}

			if err := d.checkGenesis(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check for genesis")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) startHistoricalLoop(ctx context.Context) error {
	for {
		select {
		case <-time.After(time.Second * 15):
			if d.head == nil || d.head.Finalized == nil {
				continue
			}

			if err := d.fetchHistoricalCheckpoints(ctx, d.head); err != nil {
				d.log.WithError(err).Error("Failed to fetch historical checkpoints")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) startServingLoop(ctx context.Context) error {
	if err := d.checkForNewServingCheckpoint(ctx); err != nil {
		d.log.WithError(err).Error("Failed to check for serving checkpoint")
	}

	for {
		select {
		case <-time.After(time.Second * 5):
			if err := d.checkForNewServingCheckpoint(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check for new serving checkpoint")

				time.Sleep(time.Second * 15)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) checkForNewServingCheckpoint(ctx context.Context) error {
	d.servingMutex.Lock()
	defer d.servingMutex.Unlock()

	// Don't bother checking if we don't know the head yet.
	if d.head == nil {
		return errors.New("head finality is unknown")
	}

	if d.head.Finalized == nil {
		return errors.New("head finalized checkpoint is unknown")
	}

	logCtx := d.log.WithFields(logrus.Fields{
		"head_epoch": d.head.Finalized.Epoch,
		"head_root":  fmt.Sprintf("%#x", d.head.Finalized.Root),
	})

	// If we don't have a serving bundle already, download one.
	if d.servingBundle == nil {
		logCtx.Info("No serving bundle available, downloading")

		return d.downloadServingCheckpoint(ctx, d.head)
	}

	if d.servingBundle.Finalized == nil {
		logCtx.Info("Serving bundle is unknown, downloading")

		return d.downloadServingCheckpoint(ctx, d.head)
	}

	// If the head has moved on, download a new serving bundle.
	if d.servingBundle.Finalized.Epoch != d.head.Finalized.Epoch {
		logCtx.
			WithField("serving_epoch", d.servingBundle.Finalized.Epoch).
			WithField("serving_root", fmt.Sprintf("%#x", d.servingBundle.Finalized.Root)).
			Info("Head finality has advanced, downloading new serving bundle")

		return d.downloadServingCheckpoint(ctx, d.head)
	}

	return nil
}

func (d *Default) Healthy(ctx context.Context) (bool, error) {
	if len(d.nodes.Healthy(ctx)) == 0 {
		return false, nil
	}

	return true, nil
}

func (d *Default) Peers(ctx context.Context) (types.Peers, error) {
	peers := make(types.Peers, 0, len(d.nodes))

	for _, node := range d.nodes {
		status := "connected"

		if node.Beacon.Status().Syncing() || !node.Beacon.Status().Healthy() {
			status = "disconnected"
		}

		peers = append(peers, types.Peer{
			PeerID:    node.Config.Name,
			State:     status,
			Direction: "outbound",
		})
	}

	return peers, nil
}

func (d *Default) Syncing(ctx context.Context) (*v1.SyncState, error) {
	syncing := len(d.nodes.Healthy(ctx).Syncing(ctx)) == len(d.nodes.Healthy(ctx))

	syncState := &v1.SyncState{
		IsSyncing:    syncing,
		HeadSlot:     0,
		SyncDistance: 0,
	}

	sp, err := d.Spec()
	if err != nil {
		return syncState, err
	}

	if sp == nil {
		return syncState, errors.New("spec unknown")
	}

	if d.head != nil && d.head.Finalized != nil {
		syncState.HeadSlot = phase0.Slot(d.head.Finalized.Epoch) * sp.SlotsPerEpoch
	}

	if d.servingBundle != nil && d.servingBundle.Finalized != nil {
		syncState.SyncDistance = syncState.HeadSlot - phase0.Slot(d.servingBundle.Finalized.Epoch)*sp.SlotsPerEpoch
	}

	return syncState, nil
}

func (d *Default) Finalized(ctx context.Context) (*v1.Finality, error) {
	return d.servingBundle, nil
}

func (d *Default) Head(ctx context.Context) (*v1.Finality, error) {
	return d.head, nil
}

func (d *Default) Genesis(ctx context.Context) (*v1.Genesis, error) {
	if d.genesis == nil {
		return nil, errors.New("genesis bundle not yet available")
	}

	return d.genesis, nil
}

func (d *Default) setSpec(s *state.Spec) {
	d.specMutex.Lock()
	defer d.specMutex.Unlock()

	d.spec = s
	d.sszEncoder.SetSpec(s)
}

func (d *Default) Spec() (*state.Spec, error) {
	d.specMutex.Lock()
	defer d.specMutex.Unlock()

	if d.spec == nil {
		return nil, errors.New("config spec not yet available")
	}

	copied := *d.spec

	return &copied, nil
}

func (d *Default) SSZEncoder() *ssz.Encoder {
	return d.sszEncoder
}

func (d *Default) OperatingMode() OperatingMode {
	return d.config.Mode
}

func (d *Default) shouldDownloadStates() bool {
	return d.OperatingMode() == OperatingModeFull
}

func (d *Default) checkFinality(ctx context.Context) error {
	d.majorityMutex.Lock()
	defer d.majorityMutex.Unlock()

	aggFinality := []*v1.Finality{}
	readyNodes := d.nodes.Ready(ctx)

	for _, node := range readyNodes {
		finality, err := node.Beacon.Finality()
		if err != nil {
			d.log.Infof("Failed to get finality from node %s", node.Config.Name)

			continue
		}

		aggFinality = append(aggFinality, finality)
	}

	majority, err := checkpoints.NewMajorityDecider().Decide(aggFinality)
	if err != nil {
		return perrors.Wrap(err, "failed to decide majority finality")
	}

	if d.head == nil || d.head.Finalized == nil || d.head.Finalized.Root != majority.Finalized.Root {
		d.head = majority

		d.publishFinalityCheckpointHeadUpdated(ctx, majority)

		d.log.
			WithField("epoch", majority.Finalized.Epoch).
			WithField("root", fmt.Sprintf("%#x", majority.Finalized.Root)).
			Info("New finalized head checkpoint")

		d.metrics.ObserveHeadEpoch(majority.Finalized.Epoch)
	}

	return nil
}

func (d *Default) refreshSpec(ctx context.Context) error {
	d.log.Debug("Fetching beacon spec")

	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	s, err := upstream.Beacon.Spec()
	if err != nil {
		return err
	}

	// store the beacon state spec
	d.setSpec(s)

	d.log.Debug("Fetched beacon spec")

	return nil
}

func (d *Default) checkBeaconSpec(ctx context.Context) error {
	// No-Op if we already have a beacon spec
	_, err := d.Spec()
	if err == nil {
		return nil
	}

	return d.refreshSpec(ctx)
}

func (d *Default) checkGenesisTime(ctx context.Context) error {
	// No-Op if we already have a genesis time
	if d.genesis != nil {
		return nil
	}

	d.log.Debug("Fetching genesis time")

	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	g, err := upstream.Beacon.Genesis()
	if err != nil {
		return err
	}

	// store the genesis time
	d.genesis = g

	d.log.Info("Fetched genesis time")

	return nil
}

func (d *Default) OnFinalityCheckpointHeadUpdated(ctx context.Context, cb func(ctx context.Context, checkpoint *v1.Finality) error) {
	d.broker.On(topicFinalityHeadUpdated, func(checkpoint *v1.Finality) {
		if err := cb(ctx, checkpoint); err != nil {
			d.log.WithError(err).Error("Failed to handle finality updated")
		}
	})
}

func (d *Default) publishFinalityCheckpointHeadUpdated(_ context.Context, checkpoint *v1.Finality) {
	d.broker.Emit(topicFinalityHeadUpdated, checkpoint)
}

func (d *Default) GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetBySlot(slot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetByRoot(root)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBlockByStateRoot(ctx context.Context, stateRoot phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetByStateRoot(stateRoot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBlobSidecarsBySlot(ctx context.Context, slot phase0.Slot) ([]*deneb.BlobSidecar, error) {
	return d.blobSidecars.GetBySlot(slot)
}

func (d *Default) GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedBeaconState, error) {
	block, err := d.GetBlockBySlot(ctx, slot)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) GetBeaconStateByStateRoot(ctx context.Context, stateRoot phase0.Root) (*spec.VersionedBeaconState, error) {
	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) GetBeaconStateByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedBeaconState, error) {
	block, err := d.GetBlockByRoot(ctx, root)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) storeBlock(_ context.Context, block *spec.VersionedSignedBeaconBlock) error {
	_, err := d.Spec()
	if err != nil {
		return err
	}

	if d.genesis == nil {
		return errors.New("genesis time is unknown")
	}

	if block == nil {
		return errors.New("block is nil")
	}

	root, err := d.sszEncoder.GetBlockRoot(block)
	if err != nil {
		return err
	}

	exists, err := d.blocks.GetByRoot(root)
	if err == nil && exists != nil {
		return nil
	}

	slot, err := block.Slot()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(FinalityHaltedServingPeriod)

	if slot == phase0.Slot(0) {
		expiresAt = time.Now().Add(999999 * time.Hour)
	}

	if err := d.blocks.Add(root, block, expiresAt); err != nil {
		return err
	}

	return nil
}

func (d *Default) UpstreamsStatus(ctx context.Context) (map[string]*UpstreamStatus, error) {
	rsp := make(map[string]*UpstreamStatus)

	for _, node := range d.nodes {
		rsp[node.Config.Name] = &UpstreamStatus{
			Name:    node.Config.Name,
			Healthy: false,
		}

		rsp[node.Config.Name].Healthy = node.Beacon.Status().Healthy()

		if nodeSpec, err := node.Beacon.Spec(); err == nil {
			network := nodeSpec.ConfigName
			if network == "" {
				// Fall back to our static map.
				network = eth.GetNetworkName(nodeSpec.DepositChainID)
			}

			rsp[node.Config.Name].NetworkName = network
		}

		finality, err := node.Beacon.Finality()
		if err != nil {
			continue
		}

		if finality == nil {
			continue
		}

		rsp[node.Config.Name].Finality = finality
	}

	return rsp, nil
}

func (d *Default) ListFinalizedSlots(ctx context.Context) ([]phase0.Slot, error) {
	slots := []phase0.Slot{}

	sp, err := d.Spec()
	if err != nil {
		return slots, errors.New("no beacon chain spec available")
	}

	finality, err := d.Head(ctx)
	if err != nil {
		return slots, err
	}

	if finality.Finalized == nil {
		return slots, errors.New("no finalized checkpoint available")
	}

	latestSlot := phase0.Slot(uint64(finality.Finalized.Epoch) * uint64(sp.SlotsPerEpoch))

	//nolint:gosec // HistoricalEpochCount is validated as positive in config.
	for i, val := uint64(latestSlot), uint64(latestSlot)-uint64(sp.SlotsPerEpoch)*uint64(d.config.HistoricalEpochCount); i > val; i -= uint64(sp.SlotsPerEpoch) {
		slots = append(slots, phase0.Slot(i))
	}

	return slots, nil
}

func (d *Default) GetEpochBySlot(ctx context.Context, slot phase0.Slot) (phase0.Epoch, error) {
	_, err := d.Spec()
	if err != nil {
		return phase0.Epoch(0), errors.New("no upstream beacon state spec available")
	}

	return phase0.Epoch(uint64(slot) / uint64(d.spec.SlotsPerEpoch)), nil
}

func (d *Default) PeerCount(ctx context.Context) (uint64, error) {
	return uint64(len(d.nodes.Healthy(ctx).NotSyncing(ctx))), nil
}

func (d *Default) GetSlotTime(ctx context.Context, slot phase0.Slot) (eth.SlotTime, error) {
	SlotTime := eth.SlotTime{}

	_, err := d.Spec()
	if err != nil {
		return SlotTime, errors.New("no upstream beacon state spec available")
	}

	if d.genesis == nil {
		return SlotTime, errors.New("genesis time is unknown")
	}

	return eth.CalculateSlotTime(slot, d.genesis.GenesisTime, d.spec.SecondsPerSlot.AsDuration()), nil
}

func (d *Default) GetDepositSnapshot(ctx context.Context, epoch phase0.Epoch) (*types.DepositSnapshot, error) {
	return d.depositSnapshots.GetByEpoch(epoch)
}

// shouldFallbackToFinalized checks if the "checkpoint" endpoint should fall back
// to serving the finalized state/block. This happens when the network is finalizing
// normally (unfinality gap < UnfinalizedCheckpointMinEpochGap, default 5 epochs).
// We should only serve unfinalized checkpoints when the network has been non-finalizing
// for a significant period.
func (d *Default) shouldFallbackToFinalized(ctx context.Context) bool {
	minGap := d.config.UnfinalizedCheckpointMinEpochGap
	if minGap <= 0 {
		minGap = 5
	}

	sp, err := d.Spec()
	if err != nil {
		d.log.WithError(err).Debug("Cannot determine unfinality gap: spec not available")
		return true // Safe default: serve finalized
	}

	if d.head == nil || d.head.Finalized == nil {
		d.log.Debug("Cannot determine unfinality gap: head finality unknown")
		return true // Safe default: serve finalized
	}

	// Get current epoch from wallclock
	upstream, err := d.nodes.DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		d.log.WithError(err).Debug("Cannot determine unfinality gap: no upstream nodes")
		return true // Safe default: serve finalized
	}

	currentSlot := upstream.Beacon.Wallclock().Slots().Current()
	currentEpoch := phase0.Epoch(uint64(currentSlot.Number()) / uint64(sp.SlotsPerEpoch))
	finalizedEpoch := d.head.Finalized.Epoch

	gap := int(currentEpoch) - int(finalizedEpoch)

	d.log.WithFields(logrus.Fields{
		"currentEpoch":   currentEpoch,
		"finalizedEpoch": finalizedEpoch,
		"gap":            gap,
		"minGap":         minGap,
	}).Debug("Checking unfinality gap for checkpoint fallback")

	return gap < minGap
}

// getCheckpointEpochStartSlot returns the epoch-start slot for checkpoint sync.
// By default it returns the start of (current_epoch - 1) to avoid serving
// blocks from an incomplete epoch. The offset is configurable via
// CheckpointEpochOffset (default 1).
func (d *Default) getCheckpointEpochStartSlot(ctx context.Context) (phase0.Slot, error) {
	sp, err := d.Spec()
	if err != nil {
		return 0, errors.New("chain spec not known")
	}

	// Use DataProviders directly without Healthy/Ready filters because on
	// non-finalizing networks nodes may report is_syncing=true or even
	// temporarily unhealthy (e.g., when EL is offline), but they can still
	// serve state data from their CL.
	upstream, err := d.nodes.DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return 0, fmt.Errorf("no upstream nodes available: %w", err)
	}

	currentSlot := upstream.Beacon.Wallclock().Slots().Current()
	headSlot := currentSlot.Number()
	spe := uint64(sp.SlotsPerEpoch)
	currentEpoch := headSlot / spe

	// Default offset is 1 (previous epoch), configurable via CheckpointEpochOffset.
	// Minimum offset is 1 to avoid serving blocks from the current incomplete epoch.
	offset := uint64(d.config.CheckpointEpochOffset)
	if offset < 1 {
		offset = 1
	}
	if offset >= currentEpoch {
		return 0, fmt.Errorf("checkpoint epoch offset %d is too large for current epoch %d", offset, currentEpoch)
	}

	targetEpoch := currentEpoch - offset
	epochStartSlot := phase0.Slot(targetEpoch * spe)

	d.log.WithField("currentEpoch", currentEpoch).
		WithField("targetEpoch", targetEpoch).
		WithField("epochStartSlot", epochStartSlot).
		Info("Computed checkpoint epoch-start slot")

	return epochStartSlot, nil
}

// getCheckpointUpstream returns a random upstream data provider.
func (d *Default) getCheckpointUpstream(ctx context.Context) (*Node, error) {
	upstream, err := d.nodes.DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("no upstream nodes available: %w", err)
	}
	return upstream, nil
}

// GetBeaconStateByCheckpoint returns the beacon state for the checkpoint.
// If the network unfinality gap is less than UnfinalizedCheckpointMinEpochGap (default 5),
// falls back to serving the finalized state instead.
// Otherwise fetches the state at the epoch-start slot (always available even for empty slots).
func (d *Default) GetBeaconStateByCheckpoint(ctx context.Context) (*spec.VersionedBeaconState, error) {
	// Check if we should fall back to finalized state
	if d.shouldFallbackToFinalized(ctx) {
		d.log.Info("Checkpoint request falling back to finalized state (unfinality gap below threshold)")

		finality, err := d.Finalized(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get finalized checkpoint for fallback: %w", err)
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finalized checkpoint available for fallback")
		}

		return d.GetBeaconStateByRoot(ctx, finality.Finalized.Root)
	}

	if d.config.FixedCheckpointRoot != "" {
		upstream, err := d.getCheckpointUpstream(ctx)
		if err != nil {
			return nil, err
		}
		d.log.WithField("root", d.config.FixedCheckpointRoot).Info("Fetching checkpoint state by fixed root")
		return upstream.Beacon.FetchBeaconState(ctx, d.config.FixedCheckpointRoot)
	}

	epochStartSlot, err := d.getCheckpointEpochStartSlot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compute checkpoint slot: %w", err)
	}

	upstream, err := d.getCheckpointUpstream(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch state at the epoch-start slot. The state always exists even if
	// no block was proposed at that slot (the CL processes empty slots).
	slotStr := eth.SlotAsString(epochStartSlot)
	d.log.WithField("slot", slotStr).Info("Fetching checkpoint state from upstream")

	beaconState, err := upstream.Beacon.FetchBeaconState(ctx, slotStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch checkpoint state at slot %s: %w", slotStr, err)
	}

	return beaconState, nil
}

// GetBlockByCheckpoint returns the block for the checkpoint.
// If the network unfinality gap is less than UnfinalizedCheckpointMinEpochGap (default 5),
// falls back to serving the finalized block instead.
// Otherwise finds the last block at or before the epoch-start slot.
func (d *Default) GetBlockByCheckpoint(ctx context.Context) (*spec.VersionedSignedBeaconBlock, error) {
	// Check if we should fall back to finalized block
	if d.shouldFallbackToFinalized(ctx) {
		d.log.Info("Checkpoint request falling back to finalized block (unfinality gap below threshold)")

		finality, err := d.Finalized(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get finalized checkpoint for fallback: %w", err)
		}

		if finality == nil || finality.Finalized == nil {
			return nil, fmt.Errorf("no finalized checkpoint available for fallback")
		}

		return d.GetBlockByRoot(ctx, finality.Finalized.Root)
	}

	if d.config.FixedCheckpointRoot != "" {
		upstream, err := d.getCheckpointUpstream(ctx)
		if err != nil {
			return nil, err
		}
		d.log.WithField("root", d.config.FixedCheckpointRoot).Info("Fetching checkpoint block by fixed root")
		return upstream.Beacon.FetchBlock(ctx, d.config.FixedCheckpointRoot)
	}

	epochStartSlot, err := d.getCheckpointEpochStartSlot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compute checkpoint slot: %w", err)
	}

	upstream, err := d.getCheckpointUpstream(ctx)
	if err != nil {
		return nil, err
	}

	// Try the epoch-start slot first, then walk backwards to find the last
	// block at or before it (in case the epoch-start slot was empty).
	for slot := epochStartSlot; slot > 0; slot-- {
		slotStr := eth.SlotAsString(slot)
		block, err := upstream.Beacon.FetchBlock(ctx, slotStr)
		if err != nil {
			d.log.WithField("slot", slotStr).Debug("No block at slot, trying previous")
			continue
		}
		if block == nil {
			continue
		}

		d.log.WithField("slot", slotStr).WithField("epochStartSlot", epochStartSlot).
			Info("Found checkpoint block")
		return block, nil
	}

	return nil, errors.New("no block found for checkpoint")
}
