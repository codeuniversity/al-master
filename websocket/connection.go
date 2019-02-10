package websocket

import (
	"fmt"
	"sync"

	"github.com/codeuniversity/al-proto"

	"github.com/codeuniversity/al-master/filters"
	"github.com/gorilla/websocket"
)

//Connection is a wrapper around a websocket conn that includes handling for filtering
type Connection struct {
	Conn      *websocket.Conn
	FilterSet filters.Set

	writeMutex           *sync.Mutex
	filterSetMutex       *sync.Mutex
	onListenErrorHandler func(error)
}

//NewConnection from websocket connection.
// Starts a goroutine to listen for definitions of a filterset coming in from the websocket.conn
func NewConnection(conn *websocket.Conn) *Connection {
	c := &Connection{
		Conn:           conn,
		writeMutex:     &sync.Mutex{},
		filterSetMutex: &sync.Mutex{},
	}
	go c.Listen()
	return c
}

//OnListenError call f with the error that occured in the Listen loop.
//the listen loop will already be broken when f is called.
//If the error was able to be handled in a way where we can continue to use the connection,
//Listen() should be started again.
func (c *Connection) OnListenError(f func(error)) {
	c.onListenErrorHandler = f
}

//WriteRequestedCells checks all given cells with the filterset that the client has sent.
func (c *Connection) WriteRequestedCells(cells []*proto.Cell) error {
	c.filterSetMutex.Lock()
	defer c.filterSetMutex.Unlock()

	if c.FilterSet == nil {
		//we don't want to write anything if the client hasn't told us yet what it wants.
		return nil
	}

	message := &Message{}
	for _, cell := range cells {
		passes, warnings := c.FilterSet.Eval(cell)
		if len(warnings) > 0 {
			message.Warnings = append(message.Warnings, warnings...)
		}

		if passes {
			message.Cells = append(message.Cells, cell)
		}
	}

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	return c.Conn.WriteJSON(message)
}

//Listen for incoming definitions
func (c *Connection) Listen() {
	for {
		definitions := &[]*filters.FilterDefinition{}
		err := c.Conn.ReadJSON(definitions)
		if err != nil {
			if c.onListenErrorHandler != nil {
				c.onListenErrorHandler(err)
			} else {
				fmt.Println(err, " not given to error handler")
			}
			break
		}

		if definitions != nil && len(*definitions) > 0 {
			c.filterSetMutex.Lock()
			c.FilterSet = filters.SetFromDefinitions(*definitions)
			c.filterSetMutex.Unlock()
		}
	}
}
