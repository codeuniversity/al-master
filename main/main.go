package main

import (
	"flag"
	"github.com/codeuniversity/al-master"
)

const (
	bufferSize = 1000
	httpPort   = 4000
	grpcPort   = 3000
)

func main() {
	config := master.ServerConfig{
		ConnBufferSize: bufferSize,
		GrpcPort:       grpcPort,
		HttpPort:       httpPort,
	}

	flag.StringVar(&config.StateFileName, "state_from_file", "", "specify the state name you want to load")
	flag.Parse()

	s := master.NewServer(config)
	s.Init()
	s.Run()
}
