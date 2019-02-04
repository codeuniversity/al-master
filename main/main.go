package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codeuniversity/al-proto"

	"google.golang.org/grpc"
)

const (
	address = "localhost:5000"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewCellInteractionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	batch := &proto.CellComputeBatch{}
	batch.TimeStep = 10
	newBatch, err := c.ComputeCellInteractions(ctx, batch)
	if err != nil {
		log.Fatalf("could not call: %v", err)
	}
	fmt.Print(newBatch.CellsInProximity, newBatch.CellsToCompute, newBatch.TimeStep)
}
