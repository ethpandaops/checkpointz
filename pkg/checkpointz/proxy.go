package checkpointz

import (
	"context"

	"github.com/samcm/checkpointz/pkg/checkpointz/beacon"
	"github.com/sirupsen/logrus"
)

type Provider struct {
	log *logrus.Logger
	Cfg Config

	upstream beacon.FinalityProvider
}

func NewProvider(log *logrus.Logger, conf *Config) *Provider {
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	p := &Provider{
		Cfg: *conf,
		log: log,

		upstream: beacon.NewMajorityProvider(log, conf.BeaconConfig.BeaconUpstreams),
	}

	return p
}

func (p *Provider) Start(ctx context.Context) error {
	if err := p.upstream.Start(ctx); err != nil {
		return err
	}

	return nil
}

// func (p *Proxy) Serve() error {
// 	upstreamProxies := make(map[string]*httputil.ReverseProxy)

// 	for _, upstream := range p.Cfg.BeaconConfig.BeaconUpstreams {
// 		rp, err := newHTTPReverseProxy(upstream.Address, upstream.ProxyTimeoutSeconds)
// 		if err != nil {
// 			p.log.WithError(err).Fatal("can't add beacon upstream server")
// 		}

// 		upstreamProxies[upstream.Name] = rp
// 		endpoint := fmt.Sprintf("/proxy/beacon/%s/", upstream.Name)
// 		http.HandleFunc(endpoint, p.beaconProxyRequestHandler(rp, upstream.Name))
// 	}

// 	http.HandleFunc("/status", p.statusRequestHandler())
// 	p.log.WithField("listenAddr", p.Cfg.ListenAddr).Info("started proxy server")

// 	err := http.ListenAndServe(p.Cfg.ListenAddr, nil)
// 	if err != nil {
// 		p.log.WithError(err).Fatal("can't start HTTP server")
// 	}

// 	return err
// }

// func newHTTPReverseProxy(targetHost string, proxyTimeoutSeconds uint) (*httputil.ReverseProxy, error) {
// 	u, err := url.Parse(targetHost)
// 	if err != nil {
// 		return nil, err
// 	}

// 	rp := httputil.NewSingleHostReverseProxy(u)
// 	dialer := &net.Dialer{
// 		Timeout: time.Duration(proxyTimeoutSeconds) * time.Second,
// 	}
// 	rp.Transport = &http.Transport{
// 		Dial: dialer.Dial,
// 	}

// 	return rp, nil
// }

// func (p *Proxy) beaconProxyRequestHandler(proxy *httputil.ReverseProxy, upstreamName string) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.URL.Path = strings.ReplaceAll(r.URL.Path, fmt.Sprintf("/proxy/beacon/%s", upstreamName), "")
// 		// Only allow GET methods for now
// 		if r.Method != http.MethodGet {
// 			w.WriteHeader(http.StatusMethodNotAllowed)

// 			_, err := fmt.Fprintf(w, "METHOD NOT ALLOWED\n")
// 			if err != nil {
// 				p.log.WithError(err).Error("failed writing to http.ResponseWriter")
// 			}

// 			return
// 		}
// 		// Check if path is allowed
// 		if !p.beaconAPIPathMatcher.Matches(r.URL.Path) {
// 			w.WriteHeader(http.StatusForbidden)

// 			_, err := fmt.Fprintf(w, "FORBIDDEN. Path is not allowed\n")
// 			if err != nil {
// 				p.log.WithError(err).Error("failed writing to http.ResponseWriter")
// 			}

// 			return
// 		}

// 		proxy.ServeHTTP(w, r)
// 	}
// }

// func (p *Proxy) statusRequestHandler() func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Add("Content-Type", "application/json")

// 		resp := struct {
// 			Beacon map[string]BeaconStatus `json:"beacon_nodes"`
// 		}{
// 			Beacon: p.Monitors.BeaconMonitor.status,
// 		}

// 		b, err := json.Marshal(resp)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		_, err = w.Write(b)
// 		if err != nil {
// 			p.log.WithError(err).Error("failed writing to status to http.ResponseWriter")
// 		}
// 	}
// }

// func parseRPCPayload(body []byte) (method string, err error) {
// 	rpcPayload := struct {
// 		ID     json.RawMessage   `json:"id"`
// 		Method string            `json:"method"`
// 		Params []json.RawMessage `json:"params"`
// 	}{}

// 	err = json.Unmarshal(body, &rpcPayload)
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to parse json RPC payload")
// 	}

// 	return rpcPayload.Method, nil
// }
