package main

import (
	"flag"
	master "github.com/codeuniversity/al-master"
)

const (
	bufferSize = 1000
	httpPort   = 4000
	grpcPort   = 3000
)

func main() {
	stringPtr := flag.String("stateName", "", "specify the state name you want to load")
	flag.Parse()

	s := master.NewServer(bufferSize, httpPort, grpcPort)
	s.Init(*stringPtr)
	s.Run()
}
