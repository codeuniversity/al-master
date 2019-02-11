package master

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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

	websocketConnectionsHandler *websocket.ConnectionsHandler
}

//NewServer with address to cis
func NewServer(cisAddress string, port int) *Server {
	return &Server{
		port:                        port,
		cisAddress:                  cisAddress,
		websocketConnectionsHandler: websocket.NewConnectionsHandler(),
	}
}

//InitUniverse gets initial set of cells
func (s *Server) InitUniverse() {
	s.fetchBigBang()
}

//Run offloads the computation of changes to cis
func (s *Server) Run() {
	go s.listen()
	s.withCellInteractionClient(func(c proto.CellInteractionServiceClient) {
		for {
			withTimeout(100*time.Second, func(ctx context.Context) {
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
	})
}

func (s *Server) fetchBigBang() {
	s.withCellInteractionClient(func(c proto.CellInteractionServiceClient) {
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
				s.cells = append(s.cells, cell)
			}
		})
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

func withTimeout(timeout time.Duration, f func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	f(ctx)
}
