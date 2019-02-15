package master

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/codeuniversity/al-master/websocket"
	"github.com/codeuniversity/al-proto"
	websocketConn "github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

//Server that manages cell changes
type Server struct {
	port       int
	cisAddress string

	cells    []*proto.Cell
	timeStep uint64

	cisClientPool               *CISClientPool
	websocketConnectionsHandler *websocket.ConnectionsHandler
}

//NewServer with address to cis
func NewServer(cisAddress string, cisConcurrentConns, connBufferSize int, port int) *Server {
	clientPool := NewCISClientPool(connBufferSize)
	for i := 0; i < cisConcurrentConns; i++ {
		clientPool.AddClient(createCellInteractionClient(cisAddress))
	}

	return &Server{
		port:                        port,
		cisAddress:                  cisAddress,
		websocketConnectionsHandler: websocket.NewConnectionsHandler(),
		cisClientPool:               clientPool,
	}
}

//InitUniverse gets initial set of cells
func (s *Server) InitUniverse() {
	s.fetchBigBang()
}

//Run offloads the computation of changes to cis
func (s *Server) Run() {
	go s.listen()
	for {
		s.step()
	}
}

func (s *Server) fetchBigBang() {
	withTimeout(100*time.Second, func(ctx context.Context) {
		c := s.cisClientPool.GetClient()
		defer s.cisClientPool.AddClient(c)
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
			s.cells = append(s.cells, cell)
		}
	})
}

func (s *Server) withCellInteractionClient(f func(c proto.CellInteractionServiceClient)) {
	conn, err := grpc.Dial(s.cisAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewCellInteractionServiceClient(conn)

	f(c)
}

var upgrader = websocketConn.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (s *Server) listen() {
	http.HandleFunc("/", s.websocketHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil))
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
	s.websocketConnectionsHandler.BroadcastCells(s.cells)
}

func (s *Server) step() {
	buckets := CreateBuckets(s.cells, 1000)
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
			TimeStep:         s.timeStep,
		}
		go s.callCIS(batch, wg, returnedBatchChan)
	}

	wg.Wait()
	close(returnedBatchChan)
	<-doneChan
	s.timeStep++
	fmt.Println(s.timeStep, ": ", len(s.cells))
	s.broadcastCurrentState()
}

func (s *Server) callCIS(batch *proto.CellComputeBatch, wg *sync.WaitGroup, returnedBatchChan chan *proto.CellComputeBatch) {
	looping := true
	for looping {
		withTimeout(10*time.Second, func(ctx context.Context) {
			c := s.cisClientPool.GetClient()
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
	s.cells = newCells
	doneChan <- struct{}{}
}

func withTimeout(timeout time.Duration, f func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f(ctx)
}

func createCellInteractionClient(address string) proto.CellInteractionServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return proto.NewCellInteractionServiceClient(conn)
}
