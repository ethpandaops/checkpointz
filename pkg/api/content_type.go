package api

import "fmt"

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
		return "unknown"
	}

	return ""
}

func DeriveContentType(accept string) ContentType {
	switch accept {
	case "application/json":
		return ContentTypeJSON
	// Default to JSON if they don't care what they get.
	case "*/*":
		return ContentTypeJSON
	case "application/yaml":
		return ContentTypeYAML
	case "application/octet-stream":
		return ContentTypeSSZ
	// TODO(sam.caldermason): HACK to support Nimbus - what should we do here?
	case "application/octet-stream,application/json;q=0.9":
		return ContentTypeSSZ
	}

	return ContentTypeUnknown
}

func ValidateContentType(contentType ContentType, accepting []ContentType) error {
	if !DoesAccept(accepting, contentType) {
		return fmt.Errorf("unsupported content-type: %s", contentType.String())
	}

	return nil
}
