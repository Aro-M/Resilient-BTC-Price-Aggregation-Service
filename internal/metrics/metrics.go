package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FetchSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fetch_success_total",
		Help: "Total successful BTC price fetches per source.",
	}, []string{"source"})

	FetchFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fetch_failure_total",
		Help: "Total failed BTC price fetches per source.",
	}, []string{"source"})

	CurrentPrice = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "current_price",
		Help: "Current aggregated BTC/USD price.",
	})

	SourceStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "source_status",
		Help: "Source health: 1=healthy, 0=failing.",
	}, []string{"source"})
)