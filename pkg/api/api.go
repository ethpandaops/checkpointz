package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/julienschmidt/httprouter"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/samcm/checkpointz/pkg/service/checkpointz"
	"github.com/samcm/checkpointz/pkg/service/eth"
	"github.com/sirupsen/logrus"
)

// Handler is an API handler that is responsible for negotiating with a HTTP api.
// All http-level concerns should be handled in this package, with the "namespaces" (eth/checkpointz)
// handling all business logic and dealing with concrete types.
type Handler struct {
	log logrus.FieldLogger

	eth         *eth.Handler
	checkpointz *checkpointz.Handler
}

func NewHandler(log logrus.FieldLogger, beac beacon.FinalityProvider) *Handler {
	return &Handler{
		log: log.WithField("module", "api"),

		eth:         eth.NewHandler(log, beac),
		checkpointz: checkpointz.NewHandler(log, beac),
	}
}

func (h *Handler) Register(ctx context.Context, router *httprouter.Router) error {
	router.GET("/eth/v2/beacon/blocks/:block_id", h.wrappedHandler(h.handleEthV2BeaconBlocks))

	router.GET("/checkpointz/v1/status", h.wrappedHandler(h.handleCheckpointzStatus))

	return nil
}

func (h *Handler) wrappedHandler(handler func(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		contentType := NewContentTypeFromRequest(r)
		ctx := r.Context()

		h.log.WithFields(logrus.Fields{
			"method":       r.Method,
			"path":         r.URL.Path,
			"content_type": contentType,
		}).Debug("Handling request")

		response, err := handler(ctx, r, p, contentType)
		if err != nil {
			// TODO(sam.calder-mason): Maybe log here?
			http.Error(w, err.Error(), response.StatusCode)

			return
		}

		data, err := response.MarshalAs(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if err := WriteContentAwareResponse(w, data, contentType); err != nil {
			h.log.WithError(err).Error("Failed to write response")
		}
	}
}

func (h *Handler) handleEthV2BeaconBlocks(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error) {
	if err := ValidateContentType(contentType, []ContentType{ContentTypeJSON, ContentTypeSSZ}); err != nil {
		return NewUnsupportedMediaTypeResponse(nil), err
	}

	req, err := eth.NewValidatedBeaconBlockRequestFromRequest(r, p)
	if err != nil {
		return NewBadRequestResponse(nil), err
	}

	block, err := h.eth.BeaconBlock(ctx, req)
	if err != nil {
		return NewInternalServerErrorResponse(nil), err
	}

	switch block.Version {
	case spec.DataVersionPhase0:
		return NewSuccessResponse(ContentTypeResolvers{
			ContentTypeJSON: block.Phase0.MarshalJSON,
			ContentTypeSSZ:  block.Phase0.MarshalSSZ,
		}), nil
	case spec.DataVersionAltair:
		return NewSuccessResponse(ContentTypeResolvers{
			ContentTypeJSON: block.Altair.MarshalJSON,
			ContentTypeSSZ:  block.Altair.MarshalSSZ,
		}), nil
	case spec.DataVersionBellatrix:
		return NewSuccessResponse(ContentTypeResolvers{
			ContentTypeJSON: block.Bellatrix.MarshalJSON,
			ContentTypeSSZ:  block.Bellatrix.MarshalSSZ,
		}), nil
	default:
		return NewInternalServerErrorResponse(nil), errors.New("unknown block version")
	}
}

func (h *Handler) handleCheckpointzStatus(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error) {
	if err := ValidateContentType(contentType, []ContentType{ContentTypeJSON}); err != nil {
		return NewUnsupportedMediaTypeResponse(nil), err
	}

	status, err := h.checkpointz.V1Status(ctx, checkpointz.NewStatusRequest())
	if err != nil {
		return NewInternalServerErrorResponse(nil), err
	}

	return NewSuccessResponse(ContentTypeResolvers{
		ContentTypeJSON: func() ([]byte, error) {
			return json.Marshal(status)
		},
	}), nil
}
