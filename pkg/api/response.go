package api

import (
	"fmt"
	"net/http"
)

type ContentTypeResolver func() ([]byte, error)
type ContentTypeResolvers map[ContentType]ContentTypeResolver

type HTTPResponse struct {
	resolvers  ContentTypeResolvers
	StatusCode int `json:"status_code"`
}

func (r HTTPResponse) MarshalAs(contentType ContentType) ([]byte, error) {
	if _, exists := r.resolvers[contentType]; !exists {
		return nil, fmt.Errorf("unsupported content-type: %s", contentType.String())
	}

	return r.resolvers[contentType]()
}

func NewSuccessResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusOK,
	}
}

func NewInternalServerErrorResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusInternalServerError,
	}
}

func NewBadRequestResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusBadRequest,
	}
}

func NewUnsupportedMediaTypeResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusUnsupportedMediaType,
	}
}
