package api

import "net/http"

func DoesAccept(accepts []ContentType, input ContentType) bool {
	for _, a := range accepts {
		if a == input {
			return true
		}
	}

	return false
}

func NewContentTypeFromRequest(r *http.Request) ContentType {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return ContentTypeJSON
	}

	content := DeriveContentType(accept)

	if content == ContentTypeUnknown {
		return ContentTypeJSON
	}

	return content
}
