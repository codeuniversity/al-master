package master

import (
	"github.com/codeuniversity/al-proto"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestNameOfLatestState(t *testing.T) {
	testStatesFolderName := "states-test"
	err := os.MkdirAll(filepath.Join(testStatesFolderName), 0755)
	_, err = os.Create(filepath.Join(testStatesFolderName, "STATE_20190222194317"))
	_, err = os.Create(filepath.Join(testStatesFolderName, "STATE_20190222194284"))
	_, err = os.Create(filepath.Join(testStatesFolderName, "STAT_201902A22194284"))

	files, err := ioutil.ReadDir(filepath.Join(testStatesFolderName))

	if err != nil {
		log.Fatal("creating or reading files for test failed")
	}
	assert.Equal(t, "STATE_20190222194317", nameOfLatestState(files))
}

func TestStateNameValid(t *testing.T) {
	t.Run("with valid state name", func(t *testing.T) {
		stateName := "STATE_20190222194317"
		assert.True(t, stateNameValid(stateName))
	})
	t.Run("with invalid state name", func(t *testing.T) {
		stateName := "STTE_2019022a2194241"
		assert.False(t, stateNameValid(stateName))
	})
}

func TestStateNameToInt(t *testing.T) {
	t.Run("with valid state name", func(t *testing.T) {
		stateName := "STATE_20190222194241"
		stateInt, _ := stateNameToInt(stateName)
		assert.Equal(t, int64(20190222194241), stateInt)
	})
	t.Run("with invalid state name", func(t *testing.T) {
		stateName := "STTE_2019022a2194241"
		stateInt, _ := stateNameToInt(stateName)
		assert.Equal(t, int64(0), stateInt)
	})
}

func TestMinMaxBucketCells(t *testing.T) {
	cell1 := &proto.Cell{Id: "1", Pos: &proto.Vector{}}
	cell2 := &proto.Cell{Id: "2", Pos: &proto.Vector{}}
	cell3 := &proto.Cell{Id: "3", Pos: &proto.Vector{}}

	t.Run("with buckets of length 0", func(t *testing.T) {
		buckets := Buckets{
			"key1": []*proto.Cell{cell1},
			"key2": []*proto.Cell{},
			"key3": []*proto.Cell{cell1, cell2, cell3},
		}
		min, max := minMaxBucketCells(&buckets)
		assert.Equal(t, float64(0), min)
		assert.Equal(t, float64(3), max)
	})
	t.Run("with no buckets", func(t *testing.T) {
		buckets := Buckets{}
		min, max := minMaxBucketCells(&buckets)
		assert.Equal(t, float64(0), min)
		assert.Equal(t, float64(0), max)
	})

}

func TestMedianCellsPerBucket(t *testing.T) {
	cell1 := &proto.Cell{Id: "1", Pos: &proto.Vector{}}
	cell2 := &proto.Cell{Id: "2", Pos: &proto.Vector{}}
	cell3 := &proto.Cell{Id: "3", Pos: &proto.Vector{}}
	cell4 := &proto.Cell{Id: "4", Pos: &proto.Vector{}}

	t.Run("even number of buckets", func(t *testing.T) {
		buckets := Buckets{
			"key1": []*proto.Cell{},
			"key2": []*proto.Cell{cell1, cell2},
			"key3": []*proto.Cell{cell1, cell2, cell3, cell4},
		}
		median := medianCellsPerBucket(&buckets)
		assert.Equal(t, float64(2), median)
	})
	t.Run("odd number of buckets", func(t *testing.T) {
		buckets := Buckets{
			"key1": []*proto.Cell{},
			"key2": []*proto.Cell{cell1},
			"key3": []*proto.Cell{cell1, cell2, cell3},
		}
		median := medianCellsPerBucket(&buckets)
		assert.Equal(t, float64(1), median)
	})
	t.Run("no buckets", func(t *testing.T) {
		buckets := Buckets{}
		median := medianCellsPerBucket(&buckets)
		assert.Equal(t, float64(0), median)
	})
}

func TestAverageCellsPerBucket(t *testing.T) {
	cell1 := &proto.Cell{Id: "1", Pos: &proto.Vector{}}
	cell2 := &proto.Cell{Id: "2", Pos: &proto.Vector{}}
	cell3 := &proto.Cell{Id: "3", Pos: &proto.Vector{}}

	t.Run("standard test", func(t *testing.T) {
		buckets := Buckets{
			"key1": []*proto.Cell{},
			"key2": []*proto.Cell{cell1, cell2, cell3},
			"key3": []*proto.Cell{cell1, cell2, cell3},
		}
		average := averageCellsPerBucket(&buckets)
		assert.Equal(t, float64(2), average)
	})
	t.Run("no buckets", func(t *testing.T) {
		buckets := Buckets{}
		average := averageCellsPerBucket(&buckets)
		assert.Equal(t, float64(0), average)
	})
}
