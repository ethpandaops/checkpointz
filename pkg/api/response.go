package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ContentTypeResolver func() ([]byte, error)
type ContentTypeResolvers map[ContentType]ContentTypeResolver

type HTTPResponse struct {
	resolvers  ContentTypeResolvers
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	ExtraData  map[string]interface{}
}
type jsonResponse struct {
	Data json.RawMessage `json:"data"`

	ExecutionOptimisitc string `json:"execution_optimistic,omitempty"`
	Version             string `json:"version,omitempty"`
}

func (r HTTPResponse) MarshalAs(contentType ContentType) ([]byte, error) {
	if _, exists := r.resolvers[contentType]; !exists {
		return nil, fmt.Errorf("unsupported content-type: %s", contentType.String())
	}

	if contentType != ContentTypeJSON {
		return r.resolvers[contentType]()
	}

	return r.buildWrappedJSONResponse()
}

func (r HTTPResponse) SetEtag(etag string) {
	r.Headers["ETag"] = etag
}

func (r HTTPResponse) SetCacheControl(v string) {
	r.Headers["Cache-Control"] = v
}

func (r HTTPResponse) SetEthConsensusVersion(version string) {
	r.Headers["Eth-Consensus-Version"] = version
}

func NewSuccessResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		ExtraData:  make(map[string]interface{}),
	}
}

func NewInternalServerErrorResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusInternalServerError,
		Headers:    make(map[string]string),
		ExtraData:  make(map[string]interface{}),
	}
}

func NewBadRequestResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusBadRequest,
		Headers:    make(map[string]string),
		ExtraData:  make(map[string]interface{}),
	}
}

func NewUnsupportedMediaTypeResponse(resolvers ContentTypeResolvers) *HTTPResponse {
	return &HTTPResponse{
		resolvers:  resolvers,
		StatusCode: http.StatusUnsupportedMediaType,
		Headers:    make(map[string]string),
		ExtraData:  make(map[string]interface{}),
	}
}

func (r *HTTPResponse) AddExtraData(key string, value interface{}) {
	r.ExtraData[key] = value
}

func (r *HTTPResponse) buildWrappedJSONResponse() ([]byte, error) {
	data, err := r.resolvers[ContentTypeJSON]()
	if err != nil {
		return nil, err
	}

	rsp := jsonResponse{
		Data: data,
	}

	if v, exists := r.ExtraData["execution_optimistic"]; exists {
		if st, valid := v.(string); valid {
			rsp.ExecutionOptimisitc = st
		}
	}

	if v, exists := r.ExtraData["version"]; exists {
		if st, valid := v.(string); valid {
			rsp.Version = st
		}
	}

	return json.Marshal(rsp)
}
