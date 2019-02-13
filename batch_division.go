package master

import (
	"fmt"
	"github.com/codeuniversity/al-proto"
	"math"
)

type Buckets map[string][]*proto.Cell

func CreateBuckets(cells []*proto.Cell, batchSize uint) Buckets {
	dict := make(map[string][]*proto.Cell)

	for _, cell := range cells {
		key := keyFor(cell.Pos, batchSize)

		if val, ok := dict[key]; ok {
			dict[key] = append(val, cell)
		} else {
			dict[key] = []*proto.Cell{cell}
		}
	}
	return dict
}

func keyFor(pos *proto.Vector, batchSize uint) string {
	batchXPosition := axisBatchPositionFor(pos.X, batchSize)
	batchYPosition := axisBatchPositionFor(pos.Y, batchSize)
	batchZPosition := axisBatchPositionFor(pos.Z, batchSize)
	return fmt.Sprintf("%d/%d/%d", batchXPosition, batchYPosition, batchZPosition)
}

func axisBatchPositionFor(cellAxisPosition float32, batchSize uint) int {
	if cellAxisPosition >= 0 {
		return int(math.Ceil(float64(cellAxisPosition/float32(batchSize))) * float64(batchSize))
	} else {
		return int(math.Floor(float64(cellAxisPosition/float32(batchSize))) * float64(batchSize))
	}
}
