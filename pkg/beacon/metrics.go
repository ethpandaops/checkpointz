package beacon

import (
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	servingEpoch prometheus.Gauge
}

func NewMetrics(namespace string) *Metrics {
	m := &Metrics{
		servingEpoch: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "serving_epoch",
			Help:      "The current serving epoch",
		}),
	}

	prometheus.MustRegister(m.servingEpoch)

	return m
}

func (m *Metrics) ObserveServingEpoch(epoch phase0.Epoch) {
	m.servingEpoch.Set(float64(uint64(epoch)))
}
