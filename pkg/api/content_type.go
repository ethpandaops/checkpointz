package api

import (
	"fmt"
	"strings"
)

type ContentType int

const (
	ContentTypeUnknown ContentType = iota
	ContentTypeJSON
	ContentTypeYAML
	ContentTypeSSZ
)

func (c ContentType) String() string {
	switch c {
	case ContentTypeJSON:
		return "application/json"
	case ContentTypeYAML:
		return "application/yaml"
	case ContentTypeSSZ:
		return "application/octet-stream"
	case ContentTypeUnknown:
		return "application/unknown"
	}

	return ""
}

func DeriveContentType(accept string) ContentType {
	// Split the accept header by commas to handle multiple content types
	acceptTypes := strings.Split(accept, ",")
	for _, acceptType := range acceptTypes {
		// Split each type by semicolon to handle q-values
		parts := strings.Split(acceptType, ";")
		contentType := strings.TrimSpace(parts[0])

		switch contentType {
		case "application/json":
			return ContentTypeJSON
		case "*/*":
			return ContentTypeJSON
		case "application/yaml":
			return ContentTypeYAML
		case "application/octet-stream":
			return ContentTypeSSZ
		}
	}

	// Default to JSON if they don't care what they get.
	if accept == "" {
		return ContentTypeJSON
	}

	return ContentTypeUnknown
}

func ValidateContentType(contentType ContentType, accepting []ContentType) error {
	if !DoesAccept(accepting, contentType) {
		return fmt.Errorf("unsupported content-type: %s", contentType.String())
	}

	return nil
}
