package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//AmountOfBuckets, the amount of buckets cells are currently distributed in
	AmountOfBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "amount_of_buckets",
		Help: "the amount of buckets cells are currently distributed in",
	})
	// AverageCellsPerBucket, the average number of cells throughout all buckets
	AverageCellsPerBucket = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "average_cells_per_bucket",
		Help: "the average number of cells throughout all buckets",
	})
	//MedianCellsPerBucket, the median number of cells throughout all buckets
	MedianCellsPerBucket = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "median_cells_per_bucket",
		Help: "the median number of cells throughout all buckets",
	})
	//MinCellsInBuckets, the amount of cells the bucket with the least cells contains
	MinCellsInBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "min_amount_of_cells",
		Help: "the amount of cells the bucket with the least cells contains",
	})
	//MaxCellsInBuckets, the amount of cells the bucket with the most cells contains
	MaxCellsInBuckets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "max_amount_of_cells",
		Help: "the amount of cells the bucket with the most cells contains",
	})

	//CallCISCounter, the number of times a CIS instance got called
	CallCISCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "number_of_cis_calls",
		Help: "the number of times a CIS instance got called",
	})
	//CisCallDurationMilliseconds, the amount of time it takes a CIS to respond to a call in milliseconds
	CisCallDurationMilliseconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "cis_call_duration_milliseconds",
		Help:    "the amount of time it takes a CIS to respond to a call in milliseconds",
		Buckets: prometheus.LinearBuckets(0, 10, 10),
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
