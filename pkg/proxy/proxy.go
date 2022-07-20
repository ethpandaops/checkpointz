package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skylenet/eth-proxy/pkg/jsonrpc"
)

type Proxy struct {
	Log *logrus.Logger
	Cfg Config
	Monitors
	beaconAPIPathMatcher       Matcher
	executionRPCMethodsMatcher Matcher
}

type Monitors struct {
	BeaconMonitor    *BeaconMonitor
	ExecutionMonitor *ExecutionMonitor
}

func NewProxy(log *logrus.Logger, conf *Config) *Proxy {
	if err := conf.Validate(); err != nil {
		log.Fatalf("can't start proxy: %s", err)
	}

	bm := NewBeaconMonitor(log, conf.BeaconUpstreams)
	em := NewExecutionMonitor(log, conf.ExecutionUpstreams)
	p := &Proxy{
		Cfg: *conf,
		Log: log,
		Monitors: Monitors{
			BeaconMonitor:    bm,
			ExecutionMonitor: em,
		},
	}

	p.beaconAPIPathMatcher = NewAllowMatcher(p.Cfg.BeaconConfig.APIAllowPaths)
	p.executionRPCMethodsMatcher = NewAllowMatcher(p.Cfg.ExecutionConfig.RPCAllowMethods)

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
		http.HandleFunc(endpoint, p.beaconProxyRequestHandler(rp, upstream.Name))
	}

	for _, upstream := range p.Cfg.ExecutionConfig.ExecutionUpstreams {
		rp, err := newHTTPReverseProxy(upstream.Address, p.Cfg.ExecutionConfig.ProxyTimeoutSeconds)
		if err != nil {
			p.Log.WithError(err).Fatal("can't add execution upstream server")
		}

		upstreamProxies[upstream.Name] = rp
		endpoint := fmt.Sprintf("/proxy/execution/%s/", upstream.Name)
		http.HandleFunc(endpoint, p.executionProxyRequestHandler(rp, upstream.Name))
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

func (p *Proxy) beaconProxyRequestHandler(proxy *httputil.ReverseProxy, upstreamName string) func(http.ResponseWriter, *http.Request) {
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
		if !p.beaconAPIPathMatcher.Matches(r.URL.Path) {
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

func (p *Proxy) executionProxyRequestHandler(proxy *httputil.ReverseProxy, upstreamName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		r.URL.Path = strings.ReplaceAll(r.URL.Path, fmt.Sprintf("/proxy/execution/%s", upstreamName), "")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			p.Log.WithError(err).Error("failed reading rpc body")
			w.WriteHeader(http.StatusInternalServerError)

			err = json.NewEncoder(w).Encode(jsonrpc.NewJSONRPCResponseError(json.RawMessage("1"), jsonrpc.ErrorInternal, "server error"))
			if err != nil {
				p.Log.WithError(err).Error("failed writing to http.ResponseWriter.1")
			}

			return
		}

		method, err := parseRPCPayload(body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			err = json.NewEncoder(w).Encode(jsonrpc.NewJSONRPCResponseError(json.RawMessage("1"), jsonrpc.ErrorInvalidParams, err.Error()))
			if err != nil {
				p.Log.WithError(err).Error("failed writing to http.ResponseWriter.2")
			}

			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if !p.executionRPCMethodsMatcher.Matches(method) {
			w.WriteHeader(http.StatusForbidden)

			err := json.NewEncoder(w).Encode(jsonrpc.NewJSONRPCResponseError(json.RawMessage("1"), jsonrpc.ErrorMethodNotFound, "method not allowed"))
			if err != nil {
				p.Log.WithError(err).Error("failed writing to http.ResponseWriter.3")
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
			Beacon    map[string]BeaconStatus    `json:"beacon_nodes"`
			Execution map[string]ExecutionStatus `json:"execution_nodes"`
		}{
			Beacon:    p.Monitors.BeaconMonitor.status,
			Execution: p.Monitors.ExecutionMonitor.status,
		}

		b, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			p.Log.WithError(err).Error("failed writing to status to http.ResponseWriter")
		}
	}
}

func parseRPCPayload(body []byte) (method string, err error) {
	rpcPayload := struct {
		ID     json.RawMessage   `json:"id"`
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	}{}

	err = json.Unmarshal(body, &rpcPayload)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse json RPC payload")
	}

	return rpcPayload.Method, nil
}
