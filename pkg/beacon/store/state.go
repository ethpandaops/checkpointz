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

func NewBeaconState(log logrus.FieldLogger, maxTTL time.Duration, maxItems int) *BeaconState {
	c := &BeaconState{
		log:   log.WithField("component", "beacon/store/beacon_state"),
		store: cache.NewTTLMap(maxItems, maxTTL),
	}

	c.store.OnItemEvicted(func(key string, value interface{}) {
		c.log.WithField("state_root", key).Debug("State was evicted from the cache")
	})

	return c
}

func (c *BeaconState) Add(stateRoot phase0.Root, state *[]byte) error {
	c.store.Add(eth.RootAsString(stateRoot), state)
	c.log.WithFields(
		logrus.Fields{
			"state_root": eth.RootAsString(stateRoot),
		},
	).Debug("Added state")

	return nil
}

func (c *BeaconState) GetByStateRoot(stateRoot phase0.Root) (*[]byte, error) {
	data, err := c.store.Get(eth.RootAsString(stateRoot))
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
