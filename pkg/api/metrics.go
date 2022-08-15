package api

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	requests        *prometheus.CounterVec
	responses       *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewMetrics(namespace string) Metrics {
	m := Metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "request_count",
			Help:      "Number of requests",
		}, []string{"method", "path"}),
		responses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "response_count",
			Help:      "Number of responses",
		}, []string{"method", "path", "code", "encoding"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Request duration (in seconds.)",
		}, []string{"method", "path", "encoding"}),
	}

	prometheus.MustRegister(m.requests)
	prometheus.MustRegister(m.responses)
	prometheus.MustRegister(m.requestDuration)

	return m
}

func (m Metrics) ObserveRequest(method, path string) {
	m.requests.WithLabelValues(method, path).Inc()
}

func (m Metrics) ObserveResponse(method, path, code, encoding string, duration time.Duration) {
	m.responses.WithLabelValues(method, path, code, encoding).Inc()
	m.requestDuration.WithLabelValues(method, path, encoding).Observe(duration.Seconds())
}
