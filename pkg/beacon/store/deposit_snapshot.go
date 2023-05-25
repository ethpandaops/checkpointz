package store

import (
	"errors"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/api/types"
	"github.com/ethpandaops/checkpointz/pkg/cache"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/sirupsen/logrus"
)

type DepositSnapshot struct {
	store *cache.TTLMap
	log   logrus.FieldLogger
}

func NewDepositSnapshot(log logrus.FieldLogger, config Config, namespace string) *DepositSnapshot {
	d := &DepositSnapshot{
		log:   log.WithField("component", "beacon/store/deposit_snapshot"),
		store: cache.NewTTLMap(config.MaxItems, "deposit_snapshot", namespace),
	}

	d.store.OnItemDeleted(func(key string, value interface{}, expiredAt time.Time) {
		d.log.WithField("key", key).WithField("expired_at", expiredAt.String()).Debug("Deposit snapshot was deleted from the cache")
	})

	d.store.EnableMetrics(namespace)

	return d
}

func (d *DepositSnapshot) Add(epoch phase0.Epoch, snapshot *types.DepositSnapshot, expiresAt time.Time) error {
	d.store.Add(eth.EpochAsString(epoch), snapshot, expiresAt)

	d.log.WithFields(
		logrus.Fields{
			"epoch":      eth.EpochAsString(epoch),
			"expires_at": expiresAt.String(),
		},
	).Debug("Added deposit snapshot")

	return nil
}

func (d *DepositSnapshot) GetByEpoch(epoch phase0.Epoch) (*types.DepositSnapshot, error) {
	data, _, err := d.store.Get(eth.EpochAsString(epoch))
	if err != nil {
		return nil, err
	}

	return d.parseSnapshot(data)
}

func (d *DepositSnapshot) parseSnapshot(data interface{}) (*types.DepositSnapshot, error) {
	snapshot, ok := data.(*types.DepositSnapshot)
	if !ok {
		return nil, errors.New("invalid deposit snapshot type")
	}

	return snapshot, nil
}
