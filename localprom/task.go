package localprom

import (
	"github.com/YLonely/sysdig-monitor/log"
	"github.com/prometheus/client_golang/prometheus"
)

// PromTask contains logic to record the prometheus metrics
type PromTask interface {
	// Run() should be nonblocking
	Run()
}

type taskHandler func() error
type promTask struct {
	th taskHandler
}

func NewCounterTask(counter *prometheus.Counter, h func(*prometheus.Counter) error) PromTask {
	th := func() error { return h(counter) }
	return &promTask{th: th}
}

func NewCounterVecTask(counterVec *prometheus.CounterVec, h func(*prometheus.CounterVec) error) PromTask {
	th := func() error { return h(counterVec) }
	return &promTask{th: th}
}

func NewGaugeTask(gauge *prometheus.Gauge, h func(*prometheus.Gauge) error) PromTask {
	th := func() error { return h(gauge) }
	return &promTask{th: th}
}

func NewGaugeVecTask(gaugeVec *prometheus.GaugeVec, h func(*prometheus.GaugeVec) error) PromTask {
	th := func() error { return h(gaugeVec) }
	return &promTask{th: th}
}
func NewHistogramTask(histogram *prometheus.Histogram, h func(*prometheus.Histogram) error) PromTask {
	th := func() error { return h(histogram) }
	return &promTask{th: th}
}

func NewHistogramVecTask(histogramVec *prometheus.HistogramVec, h func(*prometheus.HistogramVec) error) PromTask {
	th := func() error { return h(histogramVec) }
	return &promTask{th: th}
}

func NewSummaryTask(summary *prometheus.Summary, h func(*prometheus.Summary) error) PromTask {
	th := func() error { return h(summary) }
	return &promTask{th: th}
}

func NewSummaryVecTask(summaryVec *prometheus.SummaryVec, h func(*prometheus.SummaryVec) error) PromTask {
	th := func() error { return h(summaryVec) }
	return &promTask{th: th}
}

func (pt *promTask) Run() {
	err := pt.th()
	if err != nil {
		log.L.WithError(err).Error()
	}
}
