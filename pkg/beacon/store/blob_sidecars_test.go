package store

import (
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestBlobSidecarAddAndGet(t *testing.T) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 10}
	namespace := "test_a"
	blobSidecarStore := NewBlobSidecar(logger, config, namespace)

	slot := phase0.Slot(100)
	expiresAt := time.Now().Add(10 * time.Minute)
	sidecars := []*deneb.BlobSidecar{
		{
			Blob: deneb.Blob{},
		},
	}

	err := blobSidecarStore.Add(slot, sidecars, expiresAt)
	assert.NoError(t, err)

	retrievedSidecars, err := blobSidecarStore.GetBySlot(slot)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSidecars)
	assert.Equal(t, sidecars, retrievedSidecars)
}

func TestBlobSidecarGetBySlotNotFound(t *testing.T) {
	logger, _ := test.NewNullLogger()
	config := Config{MaxItems: 10}
	namespace := "test_b"
	blobSidecarStore := NewBlobSidecar(logger, config, namespace)

	slot := phase0.Slot(200)

	retrievedSidecars, err := blobSidecarStore.GetBySlot(slot)
	assert.Error(t, err)
	assert.Nil(t, retrievedSidecars)
}
