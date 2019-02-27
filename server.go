package master

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/codeuniversity/al-master/websocket"
	"github.com/codeuniversity/al-proto"
	websocketConn "github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

const (
	statesFolderName = "states"
)

//ServerConfig contains config data for Server
type ServerConfig struct {
	ConnBufferSize  int
	GRPCPort        int
	HTTPPort        int
	StateFileName   string
	LoadLatestState bool
}

//Server that manages cell changes
type Server struct {
	ServerConfig

	Cells    []*proto.Cell
	TimeStep uint64

	cisClientPool               *CISClientPool
	websocketConnectionsHandler *websocket.ConnectionsHandler

	grpcServer *grpc.Server
	httpServer *http.Server
}

//NewServer with given config
func NewServer(config ServerConfig) *Server {
	clientPool := NewCISClientPool(config.ConnBufferSize)

	return &Server{
		ServerConfig:                config,
		websocketConnectionsHandler: websocket.NewConnectionsHandler(),
		cisClientPool:               clientPool,
	}
}

//Init loads state from a file or by asking a cis instance for a new BigBang depending on ServerConfig
func (s *Server) Init() {
	go s.listen()

	if s.StateFileName != "" {
		if err := s.loadState(filepath.Join(statesFolderName, s.StateFileName)); err != nil {
			fmt.Println("\nLoading state from filepath failed, exiting now", err)
			panic(err)
		}
		return
	}

	if s.LoadLatestState {
		if err := s.loadLatestState(); err != nil {
			fmt.Println("\nLoading latest state failed, exiting now", err)
			panic(err)
		}
		return
	}

	s.fetchBigBang()
}

//Run offloads the computation of changes to cis
func (s *Server) Run() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		if len(signals) != 0 {
			fmt.Println("Received Signal:", <-signals)
			break
		}
		s.step()
	}
	s.shutdown()
}

func (s *Server) shutdown() {
	s.websocketConnectionsHandler.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		fmt.Println("Couldn't shutdown http server", err)
	}
	s.grpcServer.Stop()

	err = s.saveState()
	if err == nil {
		fmt.Println("\nState successfully saved")
	} else {
		fmt.Println("\nState could not be saved:", err)
	}
}

func (s *Server) saveState() error {
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

func (s *Server) loadState(statePath string) error {
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

func (s *Server) loadLatestState() error {
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

//Register cis-slave and create clients to make the slave useful
func (s *Server) Register(ctx context.Context, registration *proto.SlaveRegistration) (*proto.SlaveRegistrationResponse, error) {
	for i := 0; i < int(registration.Threads); i++ {
		client, err := createCellInteractionClient(registration.Address)
		if err != nil {
			return nil, err
		}
		s.cisClientPool.AddClient(client)
	}
	return &proto.SlaveRegistrationResponse{}, nil
}

func (s *Server) fetchBigBang() {
	c := s.cisClientPool.GetClient()
	defer s.cisClientPool.AddClient(c)
	withTimeout(100*time.Second, func(ctx context.Context) {
		stream, err := c.BigBang(ctx, &proto.BigBangRequest{})
		if err != nil {
			panic(err)
		}

		for {
			cell, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			s.Cells = append(s.Cells, cell)
		}
	})
}

var upgrader = websocketConn.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (s *Server) listen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s.grpcServer = grpc.NewServer()
	proto.RegisterSlaveRegistrationServiceServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	http.HandleFunc("/", s.websocketHandler)
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%v", s.HTTPPort), Handler: nil}
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s.websocketConnectionsHandler.AddConnection(conn)
}

func (s *Server) broadcastCurrentState() {
	s.websocketConnectionsHandler.BroadcastCells(s.Cells)
}

func (s *Server) step() {
	buckets := CreateBuckets(s.Cells, 1000)
	fmt.Println(len(buckets))
	wg := &sync.WaitGroup{}
	returnedBatchChan := make(chan *proto.CellComputeBatch)
	doneChan := make(chan struct{})

	go s.processReturnedBatches(returnedBatchChan, doneChan)

	for key, bucket := range buckets {
		surroundingCells := []*proto.Cell{}
		for _, otherKey := range key.SurroundingKeys() {
			if otherBucket, ok := buckets[otherKey]; ok {
				surroundingCells = append(surroundingCells, otherBucket...)
			}
		}
		wg.Add(1)
		batch := &proto.CellComputeBatch{
			CellsToCompute:   bucket,
			CellsInProximity: surroundingCells,
			TimeStep:         s.TimeStep,
		}
		go s.callCIS(batch, wg, returnedBatchChan)
	}

	wg.Wait()
	close(returnedBatchChan)
	<-doneChan
	s.TimeStep++
	fmt.Println(s.TimeStep, ": ", len(s.Cells))
	s.broadcastCurrentState()
}

func (s *Server) callCIS(batch *proto.CellComputeBatch, wg *sync.WaitGroup, returnedBatchChan chan *proto.CellComputeBatch) {
	looping := true
	for looping {
		c := s.cisClientPool.GetClient()
		withTimeout(10*time.Second, func(ctx context.Context) {
			returnedBatch, err := c.ComputeCellInteractions(ctx, batch)
			s.cisClientPool.AddClient(c)
			if err == nil {
				returnedBatchChan <- returnedBatch
				looping = false
			}
		})
	}
	wg.Done()
}

func (s *Server) processReturnedBatches(returnedBatchChan chan *proto.CellComputeBatch, doneChan chan struct{}) {
	newCells := []*proto.Cell{}
	for returnedBatch := range returnedBatchChan {
		newCells = append(newCells, returnedBatch.CellsToCompute...)
	}
	s.Cells = newCells
	doneChan <- struct{}{}
}

func withTimeout(timeout time.Duration, f func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f(ctx)
}

func createCellInteractionClient(address string) (proto.CellInteractionServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewCellInteractionServiceClient(conn), nil
}
