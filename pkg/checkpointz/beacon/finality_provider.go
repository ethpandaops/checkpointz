package beacon

import (
	"context"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
)

// FinalityProvider is a provider of finality information.
type FinalityProvider interface {
	// Start starts the provider.
	Start(ctx context.Context) error
	// Healthy returns true if the provider is healthy.
	Healthy(ctx context.Context) (bool, error)
	// Syncing returns true if the provider is syncing.
	Syncing(ctx context.Context) (bool, error)
	// Finality returns the finality.
	Finality(ctx context.Context) (*v1.Finality, error)
}
