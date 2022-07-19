package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Proxy struct {
	Log *logrus.Logger
	Cfg Config
	Monitors
	beaconAllowedPathRegex string
}

type Monitors struct {
	BeaconMonitor *BeaconMonitor
}

func NewProxy(log *logrus.Logger, conf *Config) *Proxy {
	if err := conf.Validate(); err != nil {
		log.Fatalf("can't start proxy: %s", err)
	}

	bm := NewBeaconMonitor(log, conf.BeaconUpstreams)
	p := &Proxy{
		Cfg: *conf,
		Log: log,
		Monitors: Monitors{
			BeaconMonitor: bm,
		},
	}

	pathRegex := ""
	for i, path := range p.Cfg.BeaconConfig.APIAllowPath {
		pathRegex += fmt.Sprintf("(%s)", path)
		if i < len(p.Cfg.BeaconConfig.APIAllowPath)-1 {
			pathRegex += "|"
		}
	}

	p.beaconAllowedPathRegex = pathRegex

	return p
}

func (p *Proxy) Serve() error {
	upstreamProxies := make(map[string]*httputil.ReverseProxy)

	for _, upstream := range p.Cfg.BeaconConfig.BeaconUpstreams {
		rp, err := newHTTPReverseProxy(upstream.Address, p.Cfg.BeaconConfig.ProxyTimeoutSeconds)
		if err != nil {
			p.Log.WithError(err).Fatal("can't add beacon upstream server")
		}

		upstreamProxies[upstream.Name] = rp
		endpoint := fmt.Sprintf("/proxy/beacon/%s/", upstream.Name)
		http.HandleFunc(endpoint, p.proxyRequestHandler(rp, upstream.Name))
	}

	http.HandleFunc("/status", p.statusRequestHandler())
	p.Log.WithField("listenAddr", p.Cfg.ListenAddr).Info("started proxy server")

	err := http.ListenAndServe(p.Cfg.ListenAddr, nil)
	if err != nil {
		p.Log.WithError(err).Fatal("can't start HTTP server")
	}

	return err
}

func newHTTPReverseProxy(targetHost string, proxyTimeoutSeconds uint) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	dialer := &net.Dialer{
		Timeout: time.Duration(proxyTimeoutSeconds) * time.Second,
	}
	rp.Transport = &http.Transport{
		Dial: dialer.Dial,
	}

	return rp, nil
}

func (p *Proxy) proxyRequestHandler(proxy *httputil.ReverseProxy, upstreamName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ReplaceAll(r.URL.Path, fmt.Sprintf("/proxy/beacon/%s", upstreamName), "")
		// Only allow GET methods for now
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)

			_, err := fmt.Fprintf(w, "METHOD NOT ALLOWED\n")
			if err != nil {
				p.Log.WithError(err).Error("failed writing to http.ResponseWriter")
			}

			return
		}
		// Check if path is allowed
		match, _ := regexp.MatchString(p.beaconAllowedPathRegex, r.URL.Path)
		if !match {
			w.WriteHeader(http.StatusForbidden)

			_, err := fmt.Fprintf(w, "FORBIDDEN. Path is not allowed\n")
			if err != nil {
				p.Log.WithError(err).Error("failed writing to http.ResponseWriter")
			}

			return
		}

		proxy.ServeHTTP(w, r)
	}
}

func (p *Proxy) statusRequestHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		resp := struct {
			Beacon map[string]BeaconStatus `json:"beaconNodes"`
		}{
			Beacon: p.Monitors.BeaconMonitor.status,
		}

		bytes, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bytes)
		if err != nil {
			p.Log.WithError(err).Error("failed writing to status to http.ResponseWriter")
		}
	}
}
