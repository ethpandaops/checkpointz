package api

import (
	"fmt"
	"net/http"
)

type ContentTypeResolver func() ([]byte, error)
type ContentTypeResolvers map[ContentType]ContentTypeResolver

type HTTPResponse struct {
	resolvers  ContentTypeResolvers
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
}

func (r HTTPResponse) MarshalAs(contentType ContentType) ([]byte, error) {
	if _, exists := r.resolvers[contentType]; !exists {
		return nil, fmt.Errorf("unsupported content-type: %s", contentType.String())
	}

	return r.resolvers[contentType]()
}

func (r HTTPResponse) SetEtag(etag string) {
	r.Headers["ETag"] = etag
}

func (r HTTPResponse) SetCacheControl(v string) {
	r.Headers["Cache-Control"] = v
}

func NewSuccessResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
	}
}

func NewInternalServerErrorResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusInternalServerError,
		Headers:    make(map[string]string),
	}
}

func NewBadRequestResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusBadRequest,
		Headers:    make(map[string]string),
	}
}

func NewUnsupportedMediaTypeResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusUnsupportedMediaType,
		Headers:    make(map[string]string),
	}
}
