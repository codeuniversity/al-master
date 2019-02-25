package main

import (
	"flag"
	"github.com/codeuniversity/al-master"
	"log"
)

const (
	bufferSize = 1000
	httpPort   = 4000
	grpcPort   = 3000
)

func main() {
	config := master.ServerConfig{
		ConnBufferSize: bufferSize,
		GRPCPort:       grpcPort,
		HTTPPort:       httpPort,
	}

	flag.StringVar(&config.StateFileName, "state_from_file", "", "input the state name you want to load")
	flag.BoolVar(
		&config.LoadLatestState,
		"load_latest_state",
		false,
		"specify if you want to load the latest state",
	)
	flag.Parse()

	if config.StateFileName != "" && config.LoadLatestState {
		log.Fatal("You shouldn't use the flags -state_from_file and -load_latest_state at the same time")
	}

	s := master.NewServer(config)
	s.Init()
	s.Run()
}
