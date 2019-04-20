package master

import (
	"github.com/codeuniversity/al-master/metrics"
	"sort"
)

//UpdateBucketsMetrics updates prometheus bucket metrics
func UpdateBucketsMetrics(buckets *Buckets) {
	minBucketCells, maxBucketCells := minMaxBucketCells(buckets)

	metrics.AmountOfBuckets.Set(float64(len(*buckets)))
	metrics.MinCellsInBuckets.Set(minBucketCells)
	metrics.MaxCellsInBuckets.Set(maxBucketCells)
	metrics.AverageCellsPerBucket.Set(averageCellsPerBucket(buckets))
	metrics.MedianCellsPerBucket.Set(medianCellsPerBucket(buckets))
}

func averageCellsPerBucket(buckets *Buckets) (average float64) {
	if len(*buckets) == 0 {
		return
	}
	var cellsInBuckets []int
	var totalAmountOfCells int

	for _, bucket := range *buckets {
		cellsInBuckets = append(cellsInBuckets, len(bucket))
		totalAmountOfCells += len(bucket)
	}
	average = float64(totalAmountOfCells) / float64(len(cellsInBuckets))
	return
}

func medianCellsPerBucket(buckets *Buckets) (median float64) {
	bucketsLength := len(*buckets)
	if bucketsLength == 0 {
		return
	}
	var cellsInBuckets []int
	for _, bucket := range *buckets {
		cellsInBuckets = append(cellsInBuckets, len(bucket))
	}
	sort.Ints(cellsInBuckets)

	if len(*buckets)%2 != 0 {
		median = float64(cellsInBuckets[(bucketsLength-1)/2])
	} else {
		median = float64((cellsInBuckets[bucketsLength/2] + cellsInBuckets[(bucketsLength/2)-1]) / 2)
	}
	return
}

//minMaxBucketCells returns the amount of cells of the bucket with the most and the least amount of cells
func minMaxBucketCells(buckets *Buckets) (minCellsInBucket float64, maxCellsInBucket float64) {
	if len(*buckets) == 0 {
		return
	}
	var cellsInBucket int
	minCellsInBucket = -1

	for _, bucket := range *buckets {
		cellsInBucket = len(bucket)
		if float64(cellsInBucket) > maxCellsInBucket {
			maxCellsInBucket = float64(cellsInBucket)
		}
		if float64(cellsInBucket) < minCellsInBucket || minCellsInBucket == -1 {
			minCellsInBucket = float64(cellsInBucket)
		}
	}
	return
}
