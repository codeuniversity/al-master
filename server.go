package master

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/codeuniversity/al-proto"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

//Server that manages cell changes
type Server struct {
	port       int
	cisAddress string

	cells    []*proto.Cell
	timeStep uint64

	conns    []*websocket.Conn
	connLock sync.Mutex
}

//NewServer with address to cis
func NewServer(cisAddress string, port int) *Server {
	return &Server{
		port:       port,
		cisAddress: cisAddress,
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
		s.withCellInteractionClient(2*time.Second, func(c proto.CellInteractionServiceClient, ctx context.Context) {
			batch := &proto.CellComputeBatch{
				CellsToCompute: s.cells,
				TimeStep:       s.timeStep,
			}
			returnedBatch, err := c.ComputeCellInteractions(ctx, batch)
			if err != nil {
				panic(err)
			}

			s.cells = returnedBatch.CellsToCompute
			s.timeStep++
			fmt.Println(s.timeStep)
			s.broadcastCurrentState()
		})

	}
}

func (s *Server) fetchBigBang() {
	s.withCellInteractionClient(10*time.Second, func(c proto.CellInteractionServiceClient, ctx context.Context) {
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

func (s *Server) withCellInteractionClient(timeout time.Duration, f func(c proto.CellInteractionServiceClient, ctx context.Context)) {
	conn, err := grpc.Dial(s.cisAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewCellInteractionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f(c, ctx)
}

var upgrader = websocket.Upgrader{
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

	s.connLock.Lock()
	defer s.connLock.Unlock()
	s.conns = append(s.conns, conn)
}

func (s *Server) broadcastCurrentState() {
	s.connLock.Lock()
	defer s.connLock.Unlock()
	indicesToRemove := []int{}
	for index, conn := range s.conns {
		err := conn.WriteJSON(s.cells)
		if err != nil {
			//assume connection is dead
			fmt.Println(err)
			indicesToRemove = append(indicesToRemove, index)
		}
	}

	if len(indicesToRemove) == 0 {
		return
	}

	newSlice := []*websocket.Conn{}
	for index, conn := range s.conns {
		if !isIncluded(index, indicesToRemove) {
			newSlice = append(newSlice, conn)
		}
	}
	s.conns = newSlice
}

func isIncluded(element int, arr []int) bool {
	for _, arrayElement := range arr {
		if arrayElement == element {
			return true
		}
	}
	return false
}
