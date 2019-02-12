package batchDivision

import (
	"github.com/codeuniversity/al-proto"
	"math"
	"math/rand"
	"strconv"
	"time"
)

func createBatches(cells []*proto.Cell, batchSize uint) map[string][]*proto.Cell {
	dict := make(map[string][]*proto.Cell)

	for i := 0; i < len(cells); i++ {
		key := createBatchKey(cells[i], batchSize)

		if val, ok := dict[key]; ok {
			dict[key] = append(val, cells[i])
		} else {
			dict[key] = []*proto.Cell{cells[i]}
		}
	}
	return dict
}

func createBatchKey(cell *proto.Cell, batchSize uint) string {
	var batchXPosition float64
	var batchYPosition float64
	var batchZPosition float64

	if cell.Pos.X >= 0 {
		batchXPosition = math.Ceil(float64(cell.Pos.X/float32(batchSize))) * float64(batchSize)
	} else {
		batchXPosition = math.Floor(float64(cell.Pos.X/float32(batchSize))) * float64(batchSize)
	}

	if cell.Pos.Y >= 0 {
		batchYPosition = math.Ceil(float64(cell.Pos.Y/float32(batchSize))) * float64(batchSize)
	} else {
		batchYPosition = math.Floor(float64(cell.Pos.Y/float32(batchSize))) * float64(batchSize)
	}

	if cell.Pos.Z >= 0 {
		batchZPosition = math.Ceil(float64(cell.Pos.Z/float32(batchSize))) * float64(batchSize)
	} else {
		batchZPosition = math.Floor(float64(cell.Pos.Z/float32(batchSize))) * float64(batchSize)
	}

	return strconv.Itoa(int(batchXPosition)) + "/" + strconv.Itoa(int(batchYPosition)) + "/" + strconv.Itoa(int(batchZPosition))
}

func randomFloatBetweenTwoFloats(float1 float32, float2 float32) float32 {
	if float1 == float2 {
		return float1
	}
	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)

	if float1 < float2 {
		return float1 + rng.Float32()*(float2-float1)
	} else {
		return float2 + rng.Float32()*(float1-float2)
	}
}

func createRandomCells(quantity uint, minX float32, maxX float32, minY float32, maxY float32, minZ float32, maxZ float32) (cells []*proto.Cell) {
	var cell proto.Cell
	for i := uint(0); i < quantity; i++ {
		cell = proto.Cell{
			Id: uint64(i),
			Pos: &proto.Vector{
				X: randomFloatBetweenTwoFloats(minX, maxX),
				Y: randomFloatBetweenTwoFloats(minY, maxY),
				Z: randomFloatBetweenTwoFloats(minZ, maxZ),
			},
		}
		cells = append(cells, &cell)
	}
	return
}
