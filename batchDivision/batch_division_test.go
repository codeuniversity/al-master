package batchDivision

import (
	"fmt"
	"github.com/codeuniversity/al-proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomFloatBetweenTwoFloats(t *testing.T) {
	t.Run("min value as first parameter and max value as second parameter", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(1.2, 1.5)
		assert.True(t, randomFloat >= 1.2 && randomFloat <= 1.5)
	})
	t.Run("max value as first parameter and min value as second parameter", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(1.5, 1.2)
		assert.True(t, randomFloat >= 1.2 && randomFloat <= 1.5)
	})
	t.Run("when min value equals max value", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(1.2, 1.2)
		assert.True(t, randomFloat == 1.2)
	})
	t.Run("first parameter minus value", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(-1.5, 1.2)
		assert.True(t, randomFloat >= -1.5 && randomFloat <= 1.2)
	})
	t.Run("second parameter minus value", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(1.5, -1.2)
		assert.True(t, randomFloat >= -1.2 && randomFloat <= 1.5)
	})
	t.Run("bot parameter minus values", func(t *testing.T) {
		randomFloat := randomFloatBetweenTwoFloats(-1.5, -1.2)
		assert.True(t, randomFloat >= -1.5 && randomFloat <= -1.2)
	})
}

func TestCreateBatchKey(t *testing.T) {
	t.Run("with positive values", func(t *testing.T) {
		cell := &proto.Cell{
			Pos: &proto.Vector{
				X: 1.4,
				Y: 0.5,
				Z: 2.6,
			},
		}
		batchSize := uint(2)
		key := createBatchKey(cell, batchSize)
		expectedKey := "2/2/4"
		assert.Equal(t, expectedKey, key)
	})
	t.Run("positive & negative values", func(t *testing.T) {
		cell := &proto.Cell{
			Pos: &proto.Vector{
				X: -1.4,
				Y: -0.5,
				Z: 2.6,
			},
		}
		batchSize := uint(2)
		key := createBatchKey(cell, batchSize)
		expectedKey := "-2/-2/4"
		assert.Equal(t, expectedKey, key)
	})
}

func TestCreateBatches(t *testing.T) {
	cell1 := &proto.Cell{Pos: &proto.Vector{X: 1.1, Y: 1.2, Z: 1.3}}
	cell2 := &proto.Cell{Pos: &proto.Vector{X: 1.2, Y: 1.3, Z: 1.4}}
	cell3 := &proto.Cell{Pos: &proto.Vector{X: 1.7, Y: 2.1, Z: 8.4}}
	cell4 := &proto.Cell{Pos: &proto.Vector{X: -0.4, Y: -9.8, Z: 5.4}}
	cells := []*proto.Cell{cell1, cell2, cell3, cell4}

	t.Run("batch size 1", func(t *testing.T) {
		batchSize := uint(1)

		dict := createBatches(cells, batchSize)

		assert.Equal(t, []*proto.Cell{cell1, cell2}, dict["2/2/2"])
		assert.Equal(t, []*proto.Cell{cell3}, dict["2/3/9"])
		assert.Equal(t, []*proto.Cell{cell4}, dict["-1/-10/6"])
	})

	t.Run("batch size 4", func(t *testing.T) {
		batchSize := uint(4)

		dict := createBatches(cells, batchSize)

		assert.Equal(t, []*proto.Cell{cell1, cell2}, dict["4/4/4"])
		assert.Equal(t, []*proto.Cell{cell3}, dict["4/4/12"])
		assert.Equal(t, []*proto.Cell{cell4}, dict["-4/-12/8"])
	})
}

func BenchmarkCreateBatches(b *testing.B) {
	cells := createRandomCells(uint(4096000), -1000, 1000, -1000, 1000, -1000, 1000)

	for n := 1000; n <= 4096000; n = n * 2 {
		b.Run(fmt.Sprintf("benchmark with n=%d", n), func(b *testing.B) {
			createBatches(cells[:n], 10)
		})
	}
}
