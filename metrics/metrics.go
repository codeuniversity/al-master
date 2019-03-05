package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//AmountOfBuckets, the amount of buckets cells are currently distributed in
	AmountOfBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "buckets_count",
		Help: "the amount of buckets cells are currently distributed in",
	})
	// AverageCellsPerBucket, the average number of cells throughout all buckets
	AverageCellsPerBucket = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "buckets_average_cell_count",
		Help: "the average number of cells throughout all buckets",
	})
	//MedianCellsPerBucket, the median number of cells throughout all buckets
	MedianCellsPerBucket = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "buckets_median_cell_count",
		Help: "the median number of cells throughout all buckets",
	})
	//MinCellsInBuckets, the amount of cells the bucket with the least cells contains
	MinCellsInBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "buckets_min_cell_count",
		Help: "the amount of cells the bucket with the least cells contains",
	})
	//MaxCellsInBuckets, the amount of cells the bucket with the most cells contains
	MaxCellsInBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bucket_max_cell_count",
		Help: "the amount of cells the bucket with the most cells contains",
	})

	//CISCallCounter, the number of times a CIS instance got called
	CISCallCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cis_call_count",
		Help: "the number of times a CIS instance got called",
	})
	//CisCallDurationSeconds, the amount of time it takes a CIS to respond to a call in seconds
	CisCallDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "cis_call_duration_seconds",
		Help:    "the amount of time it takes a CIS to respond to a call in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 5},
	})
	//CISClientCount, the number of used CIS clients
	CISClientCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cis_client_count",
		Help: "the number of used CIS clients",
	})

	//WebSocketConnectionsCount, the number of currently active websocket connections
	WebSocketConnectionsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_connections_count",
		Help: "the number of currently active websocket connections",
	})
)
