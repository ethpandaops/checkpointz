package store

import (
	"errors"
	"time"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/checkpointz/pkg/cache"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/sirupsen/logrus"
)

type BeaconState struct {
	store *cache.TTLMap
	log   logrus.FieldLogger
}

func NewBeaconState(log logrus.FieldLogger, config Config, namespace string) *BeaconState {
	c := &BeaconState{
		log:   log.WithField("component", "beacon/store/beacon_state"),
		store: cache.NewTTLMap(config.MaxItems, "state", namespace),
	}

	c.store.OnItemDeleted(func(key string, value interface{}, expiredAt time.Time) {
		c.log.WithField("state_root", key).WithField("expired_at", expiredAt.String()).Debug("State was deleted from the cache")
	})

	c.store.EnableMetrics(namespace)

	return c
}

func (c *BeaconState) Add(stateRoot phase0.Root, state *spec.VersionedBeaconState, expiresAt time.Time, slot phase0.Slot) error {
	invincible := false
	if slot == 0 {
		invincible = true
	}

	c.store.Add(eth.RootAsString(stateRoot), state, expiresAt, invincible)

	c.log.WithFields(
		logrus.Fields{
			"state_root": eth.RootAsString(stateRoot),
			"expires_at": expiresAt.String(),
		},
	).Debug("Added state")

	return nil
}

func (c *BeaconState) GetByStateRoot(stateRoot phase0.Root) (*spec.VersionedBeaconState, error) {
	data, _, err := c.store.Get(eth.RootAsString(stateRoot))
	if err != nil {
		return nil, err
	}

	return c.parseState(data)
}

func (c *BeaconState) parseState(data interface{}) (*spec.VersionedBeaconState, error) {
	state, ok := data.(*spec.VersionedBeaconState)
	if !ok {
		return nil, errors.New("invalid state")
	}

	return state, nil
}
