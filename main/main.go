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

	flag.StringVar(&config.StateFileName, "state_from_file", "", "input the state name you want to load")
	flag.BoolVar(&config.LoadLatestState, "load_latest_state", false, "specify if you want to load the "+
		"latest state, by default true")
	flag.Parse()

	s := master.NewServer(config)
	s.Init()
	s.Run()
}
