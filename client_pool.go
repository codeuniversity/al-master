package master

import (
	"github.com/codeuniversity/al-proto"
)

//CISClientPool handles asynchronous access to clients
type CISClientPool struct {
	clientChan chan proto.CellInteractionServiceClient
}

//NewCISClientPool returns a ClientPool that is capable of holding poolBufferSize-clients
func NewCISClientPool(poolBufferSize int) *CISClientPool {
	return &CISClientPool{
		clientChan: make(chan proto.CellInteractionServiceClient, poolBufferSize),
	}
}

//AddClient to buffered clientChan
func (p *CISClientPool) AddClient(client proto.CellInteractionServiceClient) {
	p.clientChan <- client
}

//GetClient from buffered clientChan
func (p *CISClientPool) GetClient() proto.CellInteractionServiceClient {
	return <-p.clientChan
}
