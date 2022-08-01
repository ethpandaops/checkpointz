package checkpointz

import (
	"fmt"
	"net/http"
)

// WriteJSONResponse writes a JSON response to the given writer.
func WriteJSONResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf("{ \"data\": %s }", data)))
}

func WriteSSZResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf("%x\n", data)))
}
