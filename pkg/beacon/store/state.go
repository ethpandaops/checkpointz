package store

import (
	"errors"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/cache"
	"github.com/samcm/checkpointz/pkg/eth"
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

func (c *BeaconState) Add(stateRoot phase0.Root, state *[]byte, expiresAt time.Time) error {
	c.store.Add(eth.RootAsString(stateRoot), state, expiresAt)
	c.log.WithFields(
		logrus.Fields{
			"state_root": eth.RootAsString(stateRoot),
			"expires_at": expiresAt.String(),
		},
	).Debug("Added state")

	return nil
}

func (c *BeaconState) GetByStateRoot(stateRoot phase0.Root) (*[]byte, error) {
	data, _, err := c.store.Get(eth.RootAsString(stateRoot))
	if err != nil {
		return nil, err
	}

	return c.parseState(data)
}

func (c *BeaconState) parseState(data interface{}) (*[]byte, error) {
	state, ok := data.(*[]byte)
	if !ok {
		return nil, errors.New("invalid state")
	}

	return state, nil
}
