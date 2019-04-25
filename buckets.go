package master

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/codeuniversity/al-proto"
)

//BucketKey is of the form "<x>/<y>/<z>"-index of a cell pos
type BucketKey string

//NewBucketKey generates a BucketKey in the form of "<x>/<y>/<z>"
func NewBucketKey(x, y, z int) BucketKey {
	return BucketKey(fmt.Sprintf("%d/%d/%d", x, y, z))
}

//Buckets is a map from "x/y/z"-bucket-index of the cells
type Buckets map[BucketKey][]*proto.Cell

//CreateBuckets from the cells
func CreateBuckets(cells []*proto.Cell, batchSize uint) Buckets {
	dict := make(map[BucketKey][]*proto.Cell)

	for _, cell := range cells {
		key := bucketKeyFor(cell.Pos, batchSize)

		if val, ok := dict[key]; ok {
			dict[key] = append(val, cell)
		} else {
			dict[key] = []*proto.Cell{cell}
		}
	}
	return dict
}

//Merge otherBuckets into the Buckets Merge is called on
func (b Buckets) Merge(otherBuckets Buckets) {
	for key, cells := range otherBuckets {
		b[key] = append(b[key], cells...)
	}
}

//AllCells stored in the different Buckets
func (b Buckets) AllCells() []*proto.Cell {
	allCells := []*proto.Cell{}
	for _, cells := range b {
		allCells = append(allCells, cells...)
	}
	return allCells
}

func bucketKeyFor(pos *proto.Vector, batchSize uint) BucketKey {
	batchXPosition := axisBatchPositionFor(pos.X, batchSize)
	batchYPosition := axisBatchPositionFor(pos.Y, batchSize)
	batchZPosition := axisBatchPositionFor(pos.Z, batchSize)
	return NewBucketKey(batchXPosition, batchYPosition, batchZPosition)
}

//SurroundingKeys of the key, including diagonals
func (k BucketKey) SurroundingKeys(width int) []BucketKey {
	width64 := int64(width)
	components := strings.Split(string(k), "/")
	if len(components) != 3 {
		return nil
	}
	x, err := strconv.ParseInt(components[0], 10, 32)
	if err != nil {
		return nil
	}
	y, err := strconv.ParseInt(components[1], 10, 32)
	if err != nil {
		return nil
	}
	z, err := strconv.ParseInt(components[2], 10, 32)
	if err != nil {
		return nil
	}
	keys := []BucketKey{}
	for otherX := x - width64; otherX <= x+width64; otherX += width64 {
		for otherY := y - width64; otherY <= y+width64; otherY += width64 {
			for otherZ := z - width64; otherZ <= z+width64; otherZ += width64 {
				if otherX == x && otherY == y && otherZ == z {
					continue
				}
				keys = append(keys, NewBucketKey(int(otherX), int(otherY), int(otherZ)))
			}
		}
	}
	return keys
}

func axisBatchPositionFor(cellAxisPosition float32, batchSize uint) int {
	if cellAxisPosition >= 0 {
		return int(math.Ceil(float64(cellAxisPosition/float32(batchSize))) * float64(batchSize))
	}
	return int(math.Floor(float64(cellAxisPosition/float32(batchSize))) * float64(batchSize))
}

/*//AllocateBucketsByXAxis returns one slice with all buckets whose X axis position is positive and one slice with all buckets whose X axis is negative
func (b Buckets) AllocateBucketsByXAxis(batchSize uint) (bucketsXPosPositive, bucketsXPosNegative Buckets) {
	cells := b.AllCells()
	for _, cell := range cells {
		key := bucketKeyFor(cell.Pos, batchSize)

		if cell.Pos.X >= 0 {
			if val, ok := bucketsXPosPositive[key]; ok {
				bucketsXPosPositive[key] = append(val, cell)
			} else {
				bucketsXPosPositive[key] = []*proto.Cell{cell}
			}
		} else {
			if val, ok := bucketsXPosNegative[key]; ok {
				bucketsXPosNegative[key] = append(val, cell)
			} else {
				bucketsXPosNegative[key] = []*proto.Cell{cell}
			}
		}
	}
	return
}*/