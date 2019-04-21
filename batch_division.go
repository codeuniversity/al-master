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

func bucketKeyFor(pos *proto.Vector, batchSize uint) BucketKey {
	batchXPosition := axisBatchPositionFor(pos.X, batchSize)
	batchYPosition := axisBatchPositionFor(pos.Y, batchSize)
	batchZPosition := axisBatchPositionFor(pos.Z, batchSize)
	return NewBucketKey(batchXPosition, batchYPosition, batchZPosition)
}

//SurroundingKeys of the key, including diagonals
func (k BucketKey) SurroundingKeys() []BucketKey {
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
	for otherX := x - 1; otherX <= x+1; otherX++ {
		for otherY := y - 1; otherY <= y+1; otherY++ {
			for otherZ := z - 1; otherZ <= z+1; otherZ++ {
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
