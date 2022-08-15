package eth

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	callsCount      *prometheus.CounterVec
	errorCallsCount *prometheus.CounterVec
}

func NewMetrics(namespace string) *Metrics {
	labels := prometheus.Labels{
		"service": "eth",
	}

	namespace += "_service_eth"

	m := &Metrics{
		callsCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			ConstLabels: labels,
			Name:        "calls_count",
			Help:        "Number of calls",
		}, []string{"method", "identifier"}),
		errorCallsCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			ConstLabels: labels,
			Name:        "error_calls_count",
			Help:        "Number of calls with errors",
		}, []string{"method", "identifier"}),
	}

	prometheus.MustRegister(m.callsCount)

	return m
}

func (m *Metrics) ObserveCall(method, identifier string) {
	m.callsCount.WithLabelValues(method, identifier).Inc()
}

func (m *Metrics) ObserveErrorCall(method, identifier string) {
	m.errorCallsCount.WithLabelValues(method, identifier).Inc()
}
