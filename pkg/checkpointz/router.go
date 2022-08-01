package checkpointz

import (
	"context"
	"net/http"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/julienschmidt/httprouter"
	"github.com/samcm/checkpointz/pkg/checkpointz/beacon"
	"github.com/sirupsen/logrus"
)

type Router struct {
	log *logrus.Logger
	Cfg Config

	provider beacon.FinalityProvider
}

func NewRouter(log *logrus.Logger, conf *Config) *Router {
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	r := &Router{
		Cfg: *conf,
		log: log,

		provider: beacon.NewMajorityProvider(log, conf.BeaconConfig.BeaconUpstreams),
	}

	return r
}

func (r *Router) Start(ctx context.Context) error {
	r.provider.StartAsync(ctx)

	router := httprouter.New()

	// TODO(sam.calder-mason): Break these routes in to their own modules i.e. eth & checkpointz
	router.GET("/eth/v2/beacon/blocks/:block_id", r.handleEthV2BeaconBlocks)

	r.log.Fatal(http.ListenAndServe(r.Cfg.ListenAddr, router))

	return nil
}

func (r *Router) handleEthV2BeaconBlocks(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	r.log.WithField("path", req.URL.Path).Info("handling request")

	ctx := req.Context()

	blockID := p.ByName("block_id")
	if blockID == "" {
		w.Write([]byte("block_id is required"))

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	id, err := ParseBlockID(blockID)
	if err != nil {
		w.Write([]byte(err.Error()))
		// TODO(sam.calder-mason): Write out a beacon api compliant error.
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	block := &spec.VersionedSignedBeaconBlock{}

	switch id {
	case Slot:
		slot, err := NewSlotFromString(blockID)
		if err != nil {
			w.Write([]byte(err.Error()))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		b, err := r.provider.GetBlockBySlot(ctx, slot)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		block = b
	case Root:
		root, err := NewRootFromString(blockID)
		if err != nil {
			w.Write([]byte(err.Error()))

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		block, err = r.provider.GetBlockByRoot(ctx, root)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

	default:
		w.Write([]byte("invalid block id"))
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if block == nil {
		w.Write([]byte("block not found"))
		w.WriteHeader(http.StatusNotFound)

		return
	}

	switch req.Header.Get("Accept") {
	case "application/octet-stream":
		b, err := block.Bellatrix.MarshalSSZ()
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		WriteSSZResponse(w, b)
	default:
		data, err := block.Bellatrix.MarshalJSON()
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		WriteJSONResponse(w, data)
	}

}
