package cache

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Operations *prometheus.CounterVec
	Hits       prometheus.Counter
	Misses     prometheus.Counter
	Len        prometheus.Gauge
}

var (
	OperationADD   = "add"
	OperationGET   = "get"
	OperationDEL   = "del"
	OperationEVICT = "evict"
)

func NewMetrics(name, namespace string) Metrics {
	labels := prometheus.Labels{
		"cache": name,
	}

	m := Metrics{
		Operations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "operations",
			ConstLabels: labels,
			Help:        "Number of operations",
		}, []string{"type"}),
		Hits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			ConstLabels: labels,
			Name:        "hits",
			Help:        "Number of hits",
		}),
		Misses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			ConstLabels: labels,
			Name:        "misses",
			Help:        "Number of misses",
		}),
		Len: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			ConstLabels: labels,
			Name:        "len",
			Help:        "Count of items in the cache",
		}),
	}

	return m
}

func (m Metrics) Register() {
	prometheus.MustRegister(m.Operations)
	prometheus.MustRegister(m.Hits)
	prometheus.MustRegister(m.Misses)
	prometheus.MustRegister(m.Len)
}

func (m Metrics) ObserveOperations(opType string, n int) {
	m.Operations.WithLabelValues(opType).Add(float64(n))
}

func (m Metrics) ObserveHit() {
	m.Hits.Add(1)
}

func (m Metrics) ObserveMiss() {
	m.Misses.Add(1)
}

func (m Metrics) ObserveLen(n int) {
	m.Len.Set(float64(n))
}
