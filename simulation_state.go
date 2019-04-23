package master

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	proto "github.com/codeuniversity/al-proto"
)

//SimulationState ...
type SimulationState struct {
	CellBuckets Buckets
	TimeStep    uint64

	currentBucketRequestsInflight map[BucketKey]bool
	nextBucketRequestsInflight    map[BucketKey]bool

	currentReturnedBatchChan chan *proto.CellComputeBatch
	nextReturnedBatchChan    chan *proto.CellComputeBatch

	currentWaitGroup *sync.WaitGroup
	nextWaitGroup    *sync.WaitGroup
}

//NewSimulationState with internal fields initialized
func NewSimulationState(buckets Buckets) *SimulationState {
	state := &SimulationState{
		CellBuckets: buckets,
		TimeStep:    0,
	}
	state.intializeComplexFields()
	return state
}

//LoadSimulationState from file
func LoadSimulationState(statePath string) (*SimulationState, error) {
	file, err := os.Open(statePath)
	if err != nil {
		return nil, err
	}
	s := &SimulationState{}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(s)
	if err != nil {
		return nil, err
	}

	s.intializeComplexFields()

	return s, file.Close()
}

//LoadLatestSimulationState from file
func LoadLatestSimulationState() (*SimulationState, error) {
	files, err := ioutil.ReadDir(filepath.Join(statesFolderName))
	if err != nil {
		return nil, err
	}
	latestStateName := nameOfLatestState(files)
	return LoadSimulationState(filepath.Join(statesFolderName, latestStateName))
}

//MarkRequestInflight will panic if the key was set to inflight before
func (s *SimulationState) MarkRequestInflight(key BucketKey) {
	alreadyMarked := s.nextBucketRequestsInflight[key]
	if alreadyMarked {
		panic(fmt.Sprintf("bucket request state was request to be marked as inflight but was already marked before"))
	}

	s.nextBucketRequestsInflight[key] = true
}

//RequestInflightFromLastStep returns wether or not a request for a bucket key was already made in the last step for this point in time
func (s *SimulationState) RequestInflightFromLastStep(key BucketKey) bool {
	return s.currentBucketRequestsInflight[key]
}

//RequestInflight returns wether or not a request for a bucket key was already made
func (s *SimulationState) RequestInflight(key BucketKey) bool {
	return s.nextBucketRequestsInflight[key]
}

//CurrentReturnedBatchChan ...
func (s *SimulationState) CurrentReturnedBatchChan() chan *proto.CellComputeBatch {
	return s.currentReturnedBatchChan
}

//NextReturnedBatchChan ...
func (s *SimulationState) NextReturnedBatchChan() chan *proto.CellComputeBatch {
	return s.nextReturnedBatchChan
}

//CurrentWaitGroup ...
func (s *SimulationState) CurrentWaitGroup() *sync.WaitGroup {
	return s.currentWaitGroup
}

//NextWaitGroup ...
func (s *SimulationState) NextWaitGroup() *sync.WaitGroup {
	return s.nextWaitGroup
}

//Cycle closes the current channel, set it and the current waitgroup to the next one and create a new next ones
func (s *SimulationState) Cycle() {
	close(s.currentReturnedBatchChan)
	s.currentBucketRequestsInflight = s.nextBucketRequestsInflight
	s.nextBucketRequestsInflight = map[BucketKey]bool{}

	s.currentReturnedBatchChan = s.nextReturnedBatchChan
	s.nextReturnedBatchChan = make(chan *proto.CellComputeBatch)

}

func (s *SimulationState) intializeComplexFields() {
	s.currentBucketRequestsInflight = map[BucketKey]bool{}
	s.nextBucketRequestsInflight = map[BucketKey]bool{}

	s.currentReturnedBatchChan = make(chan *proto.CellComputeBatch)
	s.nextReturnedBatchChan = make(chan *proto.CellComputeBatch)
	s.currentWaitGroup = &sync.WaitGroup{}
	s.nextWaitGroup = &sync.WaitGroup{}
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
