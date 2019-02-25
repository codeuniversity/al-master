package websocket

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/codeuniversity/al-proto"
)

//ConnectionsHandler holds all connections and handles removing dead connections
type ConnectionsHandler struct {
	conns    []*Connection
	connLock *sync.Mutex
}

//NewConnectionsHandler with initialized mutexes
func NewConnectionsHandler() *ConnectionsHandler {
	return &ConnectionsHandler{
		connLock: &sync.Mutex{},
	}
}

//Shutdown closes all active websocket connections
func (h *ConnectionsHandler) Shutdown() {
	h.closeActiveConnections()
}

func (h *ConnectionsHandler) closeActiveConnections() {
	for _, conn := range h.conns {
		if err := conn.Conn.Close(); err != nil {
			log.Println("Couldn't close websocket connection", err)
		}
	}
}

//AddConnection to the handler
func (h *ConnectionsHandler) AddConnection(conn *websocket.Conn) {
	connectionWrapper := NewConnection(conn)
	connectionWrapper.OnListenError(func(listenErr error) {
		fmt.Println("removing connection because: ", listenErr)
		h.removeConnection(connectionWrapper)
	})

	h.connLock.Lock()
	defer h.connLock.Unlock()

	h.conns = append(h.conns, connectionWrapper)
}

//BroadcastCells to all connected clients
func (h *ConnectionsHandler) BroadcastCells(cells []*proto.Cell) {
	h.connLock.Lock()
	defer h.connLock.Unlock()
	indicesToRemove := []int{}
	for index, conn := range h.conns {
		err := conn.WriteRequestedCells(cells)
		if err != nil {
			//assume connection is dead
			fmt.Println(err)
			indicesToRemove = append(indicesToRemove, index)
		}
	}

	if len(indicesToRemove) == 0 {
		return
	}

	newSlice := []*Connection{}
	for index, conn := range h.conns {
		if !isIncluded(index, indicesToRemove) {
			newSlice = append(newSlice, conn)
		}
	}
	h.conns = newSlice
}

func (h *ConnectionsHandler) removeConnection(connectionToBeRemoved *Connection) {
	h.connLock.Lock()
	defer h.connLock.Unlock()

	newSlice := []*Connection{}
	for _, conn := range h.conns {
		if conn != connectionToBeRemoved {
			newSlice = append(newSlice, conn)
		}
	}
	h.conns = newSlice
}

func isIncluded(element int, arr []int) bool {
	for _, arrayElement := range arr {
		if arrayElement == element {
			return true
		}
	}
	return false
}
