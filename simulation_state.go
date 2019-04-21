package master

import (
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

//SimulationState ...
type SimulationState struct {
	CellBuckets Buckets
	TimeStep    uint64
}

func (s *SimulationState) saveState() error {
	saveTime := time.Now()
	err := os.MkdirAll(statesFolderName, 0755)
	if err != nil {
		return err
	}
	temporaryPath := buildTemporaryStateFilePath(saveTime)
	file, err := os.Create(temporaryPath)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, buildStateFilePath(saveTime))
}

func (s *SimulationState) loadState(statePath string) error {
	file, err := os.Open(statePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&s)
	if err != nil {
		return err
	}
	return file.Close()
}

func (s *SimulationState) loadLatestState() error {
	files, err := ioutil.ReadDir(filepath.Join(statesFolderName))
	if err != nil {
		return err
	}
	latestStateName := nameOfLatestState(files)
	return s.loadState(filepath.Join(statesFolderName, latestStateName))
}

func nameOfLatestState(files []os.FileInfo) (latestStateName string) {
	var latestStateInt int64

	for _, f := range files {
		stateName := f.Name()
		stateInt, _ := stateNameToInt(stateName)

		if stateNameValid(stateName) && stateInt > latestStateInt {
			latestStateName = stateName
			latestStateInt = stateInt
		}
	}
	return
}

func stateNameValid(stateName string) bool {
	var validStateName = regexp.MustCompile(`STATE_\d+`)
	return validStateName.MatchString(stateName)
}

func stateNameToInt(stateName string) (int64, error) {
	return strconv.ParseInt(stateName[6:], 10, 64)
}

func buildStateFilePath(saveTime time.Time) string {
	return filepath.Join(statesFolderName, "STATE_"+string(saveTime.Format("20060102150405")))
}

func buildTemporaryStateFilePath(saveTime time.Time) string {
	return filepath.Join(statesFolderName, "SAVING_"+string(saveTime.Format("20060102150405")))
}
