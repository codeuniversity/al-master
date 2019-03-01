package metrics

import "github.com/prometheus/client_golang/prometheus"

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
)
