package master

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testStatesFolderName = "states-test"

func cleanup() {
	err := os.RemoveAll(testStatesFolderName)
	if err != nil {
		panic(err)
	}
}

func TestNameOfLatestState(t *testing.T) {
	defer cleanup()

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
