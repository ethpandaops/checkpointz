package checkpointz

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samcm/checkpointz/pkg/api"
	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/sirupsen/logrus"
)

var (
	namespace = "checkpointz"
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

	provider := beacon.NewMajorityProvider(
		namespace,
		log,
		conf.BeaconConfig.BeaconUpstreams,
		conf.CheckpointzConfig.MaxBlockCacheSize,
		conf.CheckpointzConfig.MaxBlockCacheSize,
	)

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

	if err := s.http.Register(ctx, router); err != nil {
		return err
	}

	if err := s.ServeMetrics(ctx); err != nil {
		return err
	}

	s.log.Fatal(http.ListenAndServe(s.Cfg.ListenAddr, router))

	return nil
}

func (s *Server) ServeMetrics(ctx context.Context) error {
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(s.Cfg.GlobalConfig.MetricsAddr, nil); err != nil {
			s.log.Fatal(err)
		}
	}()

	return nil
}
