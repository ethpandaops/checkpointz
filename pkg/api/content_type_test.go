package api_test

import (
	"net/http"
	"testing"

	"github.com/ethpandaops/checkpointz/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestDeriveContentType(t *testing.T) {
	tests := []struct {
		name     string
		accept   string
		expected api.ContentType
	}{
		{"JSON", "application/json", api.ContentTypeJSON},
		{"Wildcard", "*/*", api.ContentTypeJSON},
		{"YAML", "application/yaml", api.ContentTypeYAML},
		{"SSZ", "application/octet-stream", api.ContentTypeSSZ},
		{"Unknown", "application/unknown", api.ContentTypeUnknown},
		{"Empty", "", api.ContentTypeJSON},
		{"QValue JSON", "application/json;q=0.8", api.ContentTypeJSON},
		{"QValue YAML", "application/yaml;q=0.5", api.ContentTypeYAML},
		{"QValue Multiple", "application/json;q=0.8, application/yaml;q=0.5", api.ContentTypeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := api.DeriveContentType(tt.accept)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType api.ContentType
		accepting   []api.ContentType
		expectError bool
	}{
		{"Valid JSON", api.ContentTypeJSON, []api.ContentType{api.ContentTypeJSON, api.ContentTypeYAML}, false},
		{"Invalid JSON", api.ContentTypeJSON, []api.ContentType{api.ContentTypeYAML}, true},
		{"Valid YAML", api.ContentTypeYAML, []api.ContentType{api.ContentTypeJSON, api.ContentTypeYAML}, false},
		{"Invalid YAML", api.ContentTypeYAML, []api.ContentType{api.ContentTypeJSON}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.ValidateContentType(tt.contentType, tt.accepting)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewContentTypeFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		accept   string
		expected api.ContentType
	}{
		{"JSON", "application/json", api.ContentTypeJSON},
		{"Wildcard", "*/*", api.ContentTypeJSON},
		{"YAML", "application/yaml", api.ContentTypeYAML},
		{"SSZ", "application/octet-stream", api.ContentTypeSSZ},
		{"Unknown", "application/unknown", api.ContentTypeJSON},
		{"Empty", "", api.ContentTypeJSON},
		{"QValue JSON", "application/json;q=0.8", api.ContentTypeJSON},
		{"QValue YAML", "application/yaml;q=0.5", api.ContentTypeYAML},
		{"QValue Multiple", "application/json;q=0.8, application/yaml;q=0.5", api.ContentTypeJSON},
		{"Nimbus example", "application/octet-stream,application/json;q=0.9", api.ContentTypeSSZ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
			assert.NoError(t, err)
			req.Header.Set("Accept", tt.accept)

			result := api.NewContentTypeFromRequest(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}
