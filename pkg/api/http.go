package api

import (
	"encoding/json"
	"net/http"
)

// WriteJSONResponse writes a JSON response to the given writer.
func WriteJSONResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", ContentTypeJSON.String())

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func WriteSSZResponse(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", ContentTypeSSZ.String())

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

func WriteErrorResponse(w http.ResponseWriter, msg string, statusCode int) error {
	w.Header().Set("Content-Type", ContentTypeJSON.String())

	w.WriteHeader(statusCode)

	bytes, err := json.Marshal(
		BeaconError{
			Message: msg,
			Code:    statusCode,
		})
	if err != nil {
		return err
	}

	if _, err := w.Write(bytes); err != nil {
		return err
	}

	return nil
}
