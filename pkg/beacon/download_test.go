package beacon

import (
	"context"
	"strings"
	"testing"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/checkpointz/pkg/beacon/ssz"
	"github.com/ethpandaops/checkpointz/pkg/beacon/store"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type mockBeaconStateFetcher struct {
	requestedStateIDs []string
	state             *spec.VersionedBeaconState
	err               error
}

func (m *mockBeaconStateFetcher) FetchBeaconState(_ context.Context, stateID string) (*spec.VersionedBeaconState, error) {
	m.requestedStateIDs = append(m.requestedStateIDs, stateID)

	return m.state, m.err
}

func TestDownloadAndStoreBeaconStateFetchesByRootAndVerifiesState(t *testing.T) {
	provider := newTestProvider(t)
	state := newTestPhase0BeaconState(phase0.Slot(1))
	expectedStateRoot, err := provider.sszEncoder.GetStateRoot(state)
	require.NoError(t, err)

	fetcher := &mockBeaconStateFetcher{
		state: state,
	}

	err = provider.downloadAndStoreBeaconState(context.Background(), expectedStateRoot, phase0.Slot(1), fetcher)
	require.NoError(t, err)
	require.Equal(t, []string{eth.RootAsString(expectedStateRoot)}, fetcher.requestedStateIDs)

	storedState, err := provider.states.GetByStateRoot(expectedStateRoot)
	require.NoError(t, err)
	require.Same(t, state, storedState)
}

func TestDownloadAndStoreBeaconStateRejectsMismatchedStateRoot(t *testing.T) {
	provider := newTestProvider(t)
	state := newTestPhase0BeaconState(phase0.Slot(1))
	unexpectedStateRoot, err := provider.sszEncoder.GetStateRoot(newTestPhase0BeaconState(phase0.Slot(2)))
	require.NoError(t, err)

	fetcher := &mockBeaconStateFetcher{
		state: state,
	}

	err = provider.downloadAndStoreBeaconState(context.Background(), unexpectedStateRoot, phase0.Slot(1), fetcher)
	require.ErrorContains(t, err, "beacon state root does not match")
	require.Equal(t, []string{eth.RootAsString(unexpectedStateRoot)}, fetcher.requestedStateIDs)

	_, err = provider.states.GetByStateRoot(unexpectedStateRoot)
	require.Error(t, err)
}

func newTestProvider(t *testing.T) *Default {
	t.Helper()

	logger, _ := test.NewNullLogger()
	namespace := strings.ReplaceAll(t.Name(), "/", "_")

	return &Default{
		log:        logger,
		sszEncoder: ssz.NewEncoder(false),
		states:     store.NewBeaconState(logger, store.Config{MaxItems: 5}, namespace),
	}
}

func newTestPhase0BeaconState(slot phase0.Slot) *spec.VersionedBeaconState {
	return &spec.VersionedBeaconState{
		Version: spec.DataVersionPhase0,
		Phase0: &phase0.BeaconState{
			Slot:                        slot,
			Fork:                        &phase0.Fork{},
			LatestBlockHeader:           &phase0.BeaconBlockHeader{},
			BlockRoots:                  make([]phase0.Root, 8192),
			StateRoots:                  make([]phase0.Root, 8192),
			HistoricalRoots:             []phase0.Root{},
			ETH1Data:                    &phase0.ETH1Data{BlockHash: make([]byte, 32)},
			ETH1DataVotes:               []*phase0.ETH1Data{},
			Validators:                  []*phase0.Validator{},
			Balances:                    []phase0.Gwei{},
			RANDAOMixes:                 make([]phase0.Root, 65536),
			Slashings:                   make([]phase0.Gwei, 8192),
			PreviousEpochAttestations:   []*phase0.PendingAttestation{},
			CurrentEpochAttestations:    []*phase0.PendingAttestation{},
			JustificationBits:           bitfield.NewBitvector4(),
			PreviousJustifiedCheckpoint: &phase0.Checkpoint{},
			CurrentJustifiedCheckpoint:  &phase0.Checkpoint{},
			FinalizedCheckpoint:         &phase0.Checkpoint{},
		},
	}
}
