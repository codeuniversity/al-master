package master

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/codeuniversity/al-master/websocket"
	"github.com/codeuniversity/al-proto"
	websocketConn "github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	statesFolderName = "states"
)

type ServerConfig struct {
	ConnBufferSize int
	GrpcPort       int
	HttpPort       int
	StateFileName  string
}

//Server that manages cell changes
type Server struct {
	ServerConfig

	Cells    []*proto.Cell
	TimeStep uint64

	cisClientPool               *CISClientPool
	websocketConnectionsHandler *websocket.ConnectionsHandler
	grpcServer                  *grpc.Server
}

//NewServer with address to cis
func NewServer(config ServerConfig) *Server {
	clientPool := NewCISClientPool(config.ConnBufferSize)

	return &Server{
		ServerConfig:                config,
		websocketConnectionsHandler: websocket.NewConnectionsHandler(),
		cisClientPool:               clientPool,
	}
}

//Init starts the server
func (s *Server) Init() {
	go s.listen()
	if s.StateFileName != "" {
		err := s.loadState(filepath.Join(statesFolderName, s.StateFileName))
		if err != nil {
			fmt.Println("\nLoading state failed, exiting now", err)
			panic(err)
		}
	} else {
		s.fetchBigBang()
	}
}

//Run offloads the computation of changes to cis
func (s *Server) Run() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		if len(signals) != 0 {
			s.grpcServer.Stop()
			fmt.Println("Received Signal:", <-signals)
			break
		}
		s.step()
	}

	err := s.saveState()
	if err == nil {
		fmt.Println("\nState successfully saved")
	} else {
		fmt.Println("\nState could not be saved:", err)
	}
}

func createDirIfNotExist(dir string) error {
	_, err := os.Stat(filepath.Join(dir))
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *Server) saveState() error {
	err := createDirIfNotExist(statesFolderName)
	if err != nil {
		return err
	}
	file, err := os.Create(s.buildStateFilePath())
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return file.Close()
	}
	return err
}

func (s *Server) loadState(statePath string) error {
	file, err := os.Open(statePath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&s)
	if err != nil {
		return file.Close()
	}
	return err
}

func (s *Server) buildStateFilePath() string {
	return filepath.Join(statesFolderName, string(time.Now().Format("20060102150405")))
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.GrpcPort))
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
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.HttpPort), nil))
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
