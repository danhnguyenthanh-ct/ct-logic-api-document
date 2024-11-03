package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	LabelStatus = "status"
	LabelMethod = "method"
)

var (
	cusMetrics *customMetrics
)

type Metrics interface {
	RecordCustomMetrics(method string, value float64, begin time.Time, err error)
}

type customMetrics struct {
	gaugeMetric   *prometheus.GaugeVec
	counterMetric *prometheus.CounterVec
	latencyMetric *prometheus.HistogramVec
}

func (m *customMetrics) RecordCustomMetrics(method string, value float64, begin time.Time, err error) {
	var labels = make(map[string]string)
	labels[LabelStatus] = "succeed"
	labels[LabelMethod] = method
	if err != nil {
		labels[LabelStatus] = "error"
	}
	m.latencyMetric.With(labels).Observe(time.Since(begin).Seconds())
}

func NewCustomMetrics() Metrics {
	cusMetrics = &customMetrics{
		gaugeMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "request_status",
			},
			[]string{LabelStatus, LabelMethod}),
		counterMetric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "request_count",
			},
			[]string{LabelStatus, LabelMethod}),
		latencyMetric: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_latency",
				Buckets: []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5},
			},
			[]string{LabelStatus, LabelMethod}),
	}
	prometheus.MustRegister(cusMetrics.gaugeMetric)
	prometheus.MustRegister(cusMetrics.counterMetric)
	prometheus.MustRegister(cusMetrics.latencyMetric)
	return cusMetrics
}
