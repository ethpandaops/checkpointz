package store

import (
	"errors"
	"time"

	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/checkpointz/pkg/cache"
	"github.com/ethpandaops/checkpointz/pkg/eth"
	"github.com/sirupsen/logrus"
)

type BlobSidecar struct {
	store *cache.TTLMap
	log   logrus.FieldLogger
}

func NewBlobSidecar(log logrus.FieldLogger, config Config, namespace string) *BlobSidecar {
	d := &BlobSidecar{
		log:   log.WithField("component", "beacon/store/blob_sidecar"),
		store: cache.NewTTLMap(config.MaxItems, "blob_sidecar", namespace),
	}

	d.store.OnItemDeleted(func(key string, value interface{}, expiredAt time.Time) {
		d.log.WithField("key", key).WithField("expired_at", expiredAt.String()).Debug("Blob sidecar was deleted from the cache")
	})

	d.store.EnableMetrics(namespace)

	return d
}

func (d *BlobSidecar) Add(slot phase0.Slot, sidecars []*deneb.BlobSidecar, expiresAt time.Time) error {
	d.store.Add(eth.SlotAsString(slot), sidecars, expiresAt, false)

	d.log.WithFields(
		logrus.Fields{
			"slot":       eth.SlotAsString(slot),
			"expires_at": expiresAt.String(),
		},
	).Debug("Added blob sidecar")

	return nil
}

func (d *BlobSidecar) GetBySlot(slot phase0.Slot) ([]*deneb.BlobSidecar, error) {
	data, _, err := d.store.Get(eth.SlotAsString(slot))
	if err != nil {
		return nil, err
	}

	return d.parseSidecar(data)
}

func (d *BlobSidecar) parseSidecar(data interface{}) ([]*deneb.BlobSidecar, error) {
	sidecar, ok := data.([]*deneb.BlobSidecar)
	if !ok {
		return nil, errors.New("invalid blob sidecar type")
	}

	return sidecar, nil
}
