package websocket

import (
	"github.com/codeuniversity/al-proto"
)

//Message that is sent through websocket to the client
type Message struct {
	Cells    []*proto.Cell `json:"cells"`
	Warnings []string      `json:"warnings"`
}
