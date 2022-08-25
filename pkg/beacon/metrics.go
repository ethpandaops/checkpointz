package beacon

import (
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	servingEpoch prometheus.Gauge
	headEpoch    prometheus.Gauge
}

func NewMetrics(namespace string) *Metrics {
	m := &Metrics{
		servingEpoch: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "serving_epoch",
			Help:      "The current serving epoch",
		}),
		headEpoch: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "head_epoch",
			Help:      "The current head finalized epoch",
		}),
	}

	prometheus.MustRegister(m.servingEpoch)
	prometheus.MustRegister(m.headEpoch)

	return m
}

func (m *Metrics) ObserveServingEpoch(epoch phase0.Epoch) {
	m.servingEpoch.Set(float64(uint64(epoch)))
}

func (m *Metrics) ObserveHeadEpoch(epoch phase0.Epoch) {
	m.headEpoch.Set(float64(uint64(epoch)))
}
