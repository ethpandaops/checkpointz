package checkpointz

import (
	"context"
	"net/http"
	"time"

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

	provider := beacon.NewDefaultProvider(
		namespace,
		log,
		conf.BeaconConfig.BeaconUpstreams,
		conf.Checkpointz,
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

	server := &http.Server{
		Addr:              s.Cfg.GlobalConfig.ListenAddr,
		ReadHeaderTimeout: 3 * time.Minute,
	}

	server.Handler = router

	if err := server.ListenAndServe(); err != nil {
		s.log.Fatal(err)
	}

	return nil
}

func (s *Server) ServeMetrics(ctx context.Context) error {
	go func() {
		server := &http.Server{
			Addr:              s.Cfg.GlobalConfig.MetricsAddr,
			ReadHeaderTimeout: 15 * time.Second,
		}

		server.Handler = promhttp.Handler()

		if err := server.ListenAndServe(); err != nil {
			s.log.Fatal(err)
		}
	}()

	return nil
}
