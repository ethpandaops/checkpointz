package checkpointz

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/samcm/checkpointz/pkg/api"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/sirupsen/logrus"
)

type Server struct {
	log *logrus.Logger
	Cfg Config

	provider beacon.FinalityProvider

	http *api.Handler
}

func NewServer(log *logrus.Logger, conf *Config) *Server {
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	provider := beacon.NewMajorityProvider(log, conf.BeaconConfig.BeaconUpstreams)

	s := &Server{
		Cfg: *conf,
		log: log,

		http: api.NewHandler(log, provider),

		provider: provider,
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	s.provider.StartAsync(ctx)

	router := httprouter.New()

	s.http.Register(ctx, router)

	s.log.Fatal(http.ListenAndServe(s.Cfg.ListenAddr, router))

	return nil
}
