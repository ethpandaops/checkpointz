package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBlobSidecarIndices(t *testing.T) {
	tests := []struct {
		name     string
		raw      []string
		expected []int
		wantErr  bool
	}{
		{"empty", []string{}, []int{}, false},
		{"single", []string{"0"}, []int{0}, false},
		{"multiple", []string{"0", "1", "2"}, []int{0, 1, 2}, false},
		{"duplicates collapsed", []string{"0", "0", "0", "1"}, []int{0, 1}, false},
		{"order preserved", []string{"2", "0", "1"}, []int{2, 0, 1}, false},
		{"negative rejected", []string{"-1"}, nil, true},
		{"non-numeric rejected", []string{"abc"}, nil, true},
		{"empty string rejected", []string{""}, nil, true},
		{"at cap", manyIndices(maxBlobSidecarIndices), nil, false},
		{"over cap rejected", manyIndices(maxBlobSidecarIndices + 1), nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			indices, err := parseBlobSidecarIndices(test.raw)

			if test.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			if test.expected != nil {
				assert.Equal(t, test.expected, indices)
			}
		})
	}
}

func TestParseBlobSidecarIndicesDuplicatesDoNotAmplify(t *testing.T) {
	raw := make([]string, 0, maxBlobSidecarIndices)
	for i := 0; i < maxBlobSidecarIndices; i++ {
		raw = append(raw, "0")
	}

	indices, err := parseBlobSidecarIndices(raw)
	require.NoError(t, err)
	assert.Equal(t, []int{0}, indices)
}

func manyIndices(n int) []string {
	raw := make([]string, 0, n)
	for i := 0; i < n; i++ {
		raw = append(raw, fmt.Sprintf("%d", i))
	}

	return raw
}
