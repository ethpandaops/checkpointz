package api

import (
	"fmt"
	"net/http"
)

// WriteJSONResponse writes a JSON response to the given writer.
func WriteJSONResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write([]byte(fmt.Sprintf("{ \"data\": %s }", data))); err != nil {
		return err
	}

	return nil
}

func WriteSSZResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func WriteContentAwareResponse(w http.ResponseWriter, data []byte, contentType ContentType) error {
	switch contentType {
	case ContentTypeJSON:
		return WriteJSONResponse(w, data)
	case ContentTypeSSZ:
		return WriteSSZResponse(w, data)
	default:
		return WriteJSONResponse(w, data)
	}
}
