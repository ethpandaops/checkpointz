package beacon

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/chuckpreslar/emission"
	"github.com/go-co-op/gocron"
	"github.com/samcm/beacon/state"
	"github.com/samcm/checkpointz/pkg/beacon/checkpoints"
	"github.com/samcm/checkpointz/pkg/beacon/node"
	"github.com/samcm/checkpointz/pkg/beacon/store"
	"github.com/sirupsen/logrus"
)

type Default struct {
	log logrus.FieldLogger

	config      *Config
	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter

	head          *v1.Finality
	servingBundle *v1.Finality

	blocks *store.Block
	states *store.BeaconState

	spec    *state.Spec
	genesis *v1.Genesis

	historicalSlotFailures map[phase0.Slot]int

	metrics *Metrics
}

var _ FinalityProvider = (*Default)(nil)

var (
	topicFinalityHeadUpdated = "finality_head_updated"
)

func NewDefaultProvider(namespace string, log logrus.FieldLogger, nodes []node.Config, config *Config) FinalityProvider {
	return &Default{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/default"),
		nodes:       NewNodesFromConfig(log, nodes, namespace),
		config:      config,

		head:          &v1.Finality{},
		servingBundle: &v1.Finality{},

		historicalSlotFailures: make(map[phase0.Slot]int),

		broker: emission.NewEmitter(),
		blocks: store.NewBlock(log, config.Caches.Blocks, namespace),
		states: store.NewBeaconState(log, config.Caches.States, namespace),

		metrics: NewMetrics(namespace + "_beacon"),
	}
}

func (d *Default) Start(ctx context.Context) error {
	if err := d.nodes.StartAll(ctx); err != nil {
		return err
	}

	go func() {
		for {
			// Wait until we have a single healthy node.
			_, err := d.nodes.Healthy(ctx).NotSyncing(ctx).RandomNode(ctx)
			if err != nil {
				d.log.WithError(err).Error("Waiting for a healthy, non-syncing node before beginning..")
				time.Sleep(time.Second * 5)

				continue
			}

			if err := d.startCrons(ctx); err != nil {
				d.log.WithError(err).Fatal("Failed to start crons")
			}

			break
		}
	}()

	return nil
}

func (d *Default) startCrons(ctx context.Context) error {
	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("5s").Do(func() {
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

	go func() {
		if err := d.startGenesisLoop(ctx); err != nil {
			d.log.WithError(err).Fatal("Failed to start genesis loop")
		}
	}()

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
	for {
		select {
		case <-time.After(time.Second * 1):
			if err := d.checkForNewServingCheckpoint(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check for new serving checkpoint")

				time.Sleep(time.Second * 30)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) checkForNewServingCheckpoint(ctx context.Context) error {
	// Don't bother checking if we don't know the head yet.
	if d.head == nil {
		return nil
	}

	if d.head.Finalized == nil {
		return nil
	}

	// If head == serving, we're done.
	if d.servingBundle != nil && d.servingBundle.Finalized != nil && d.servingBundle.Finalized.Epoch == d.head.Finalized.Epoch {
		return nil
	}

	if err := d.downloadServingCheckpoint(ctx, d.head); err != nil {
		return err
	}

	return nil
}
func (d *Default) Healthy(ctx context.Context) (bool, error) {
	if len(d.nodes.Healthy(ctx)) == 0 {
		return false, nil
	}

	return true, nil
}

func (d *Default) Syncing(ctx context.Context) (bool, error) {
	if len(d.nodes.NotSyncing(ctx)) == 0 {
		return true, nil
	}

	return false, nil
}

func (d *Default) Finality(ctx context.Context) (*v1.Finality, error) {
	return d.servingBundle, nil
}

func (d *Default) checkFinality(ctx context.Context) error {
	aggFinality := []*v1.Finality{}
	readyNodes := d.nodes.Ready(ctx)

	for _, node := range readyNodes {
		finality, err := node.Beacon.GetFinality(ctx)
		if err != nil {
			d.log.Info("Failed to get finality from node", "node", node.Config.Name)

			continue
		}

		aggFinality = append(aggFinality, finality)
	}

	Default, err := checkpoints.NewMajorityDecider().Decide(aggFinality)
	if err != nil {
		return err
	}

	if d.head == nil || d.head.Finalized == nil || d.head.Finalized.Root != Default.Finalized.Root {
		d.head = Default

		d.publishFinalityCheckpointHeadUpdated(ctx, Default)

		d.log.WithField("epoch", Default.Finalized.Epoch).WithField("root", fmt.Sprintf("%#x", Default.Finalized.Root)).Info("New finalized head checkpoint")

		d.metrics.ObserveHeadEpoch(Default.Finalized.Epoch)
	}

	return nil
}

func (d *Default) checkBeaconSpec(ctx context.Context) error {
	// No-Op if we already have a beacon spec
	if d.spec != nil {
		return nil
	}

	d.log.Debug("Fetching beacon spec")

	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	s, err := upstream.Beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	// store the beacon state spec
	d.spec = s

	d.log.Info("Fetched beacon spec")

	return nil
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

	g, err := upstream.Beacon.GetGenesis(ctx)
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

func (d *Default) publishFinalityCheckpointHeadUpdated(ctx context.Context, checkpoint *v1.Finality) {
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

func (d *Default) GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*[]byte, error) {
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

func (d *Default) GetBeaconStateByStateRoot(ctx context.Context, stateRoot phase0.Root) (*[]byte, error) {
	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) GetBeaconStateByRoot(ctx context.Context, root phase0.Root) (*[]byte, error) {
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

func (d *Default) storeBlock(ctx context.Context, block *spec.VersionedSignedBeaconBlock) error {
	if d.spec == nil {
		return errors.New("beacon chain spec is unknown")
	}

	if d.genesis == nil {
		return errors.New("genesis time is unknown")
	}

	if block == nil {
		return errors.New("block is nil")
	}

	root, err := block.Root()
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

	expiresAtSlot := CalculateSlotExpiration(slot, d.config.HistoricalEpochCount*int(d.spec.SlotsPerEpoch))
	expiresAt := GetSlotTime(expiresAtSlot, d.spec.SecondsPerSlot, d.genesis.GenesisTime).
		Add(time.Minute * 15) // Store it for an extra 15 minutes to simplify the expiry logic.

	if slot == phase0.Slot(0) {
		expiresAt = time.Now().Add(999999 * time.Hour)
	}

	if err := d.blocks.Add(block, expiresAt); err != nil {
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

		if node.Beacon == nil {
			continue
		}

		rsp[node.Config.Name].Healthy = node.Beacon.GetStatus(ctx).Healthy()

		// nolint:gocritic // invalid
		if spec, err := node.Beacon.GetSpec(ctx); err == nil {
			rsp[node.Config.Name].NetworkName = spec.ConfigName
		}

		finality, err := node.Beacon.GetFinality(ctx)
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
	if d.spec == nil {
		return slots, errors.New("no beacon chain spec available")
	}

	finality, err := d.Finality(ctx)
	if err != nil {
		return slots, err
	}

	if finality.Finalized == nil {
		return slots, errors.New("no finalized checkpoint available")
	}

	latestSlot := phase0.Slot(uint64(finality.Finalized.Epoch) * uint64(d.spec.SlotsPerEpoch))

	for i, val := uint64(latestSlot), uint64(latestSlot)-uint64(d.spec.SlotsPerEpoch)*uint64(d.config.HistoricalEpochCount); i > val; i -= uint64(d.spec.SlotsPerEpoch) {
		slots = append(slots, phase0.Slot(i))
	}

	return slots, nil
}

func (d *Default) GetEpochBySlot(ctx context.Context, slot phase0.Slot) (phase0.Epoch, error) {
	if d.spec == nil {
		return phase0.Epoch(0), errors.New("no upstream beacon state spec available")
	}

	return phase0.Epoch(uint64(slot) / uint64(d.spec.SlotsPerEpoch)), nil
}
