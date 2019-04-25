package main

import (
	"flag"
	"log"
	_ "net/http/pprof"

	"github.com/codeuniversity/al-master"
)

const (
	bufferSize = 1000
)

func main() {
	serverConfig := master.ServerConfig{
		ConnBufferSize: bufferSize,
	}

	var subMasterData master.SubMasterData

	flag.IntVar(
		&serverConfig.GRPCPort,
		"grpc_port",
		3000,
		"the grpc port")

	flag.IntVar(
		&serverConfig.HTTPPort,
		"http_port",
		4000,
		"the http_port")

	flag.StringVar(
		&subMasterData.ChiefMasterAddress,
		"chief_master_address",
		"",
		"provide the address of the chief master you want to register to")

	flag.StringVar(
		&serverConfig.StateFileName,
		"state_from_file",
		"",
		"input the state name you want to load")

	flag.BoolVar(
		&serverConfig.LoadLatestState,
		"load_latest_state",
		false,
		"specify if you want to load the latest state",
	)
	flag.StringVar(
		&serverConfig.BigBangConfigPath,
		"big_bang_config_path",
		"./big_bang_config.yaml",
		"Path to the Big-Bang Config")

	flag.IntVar(
		&serverConfig.BucketWidth,
		"bucket_width",
		500,
		"defines the edge length of a bucket")

	flag.Parse()

	if serverConfig.StateFileName != "" && serverConfig.LoadLatestState {
		log.Fatal("You shouldn't use the flags -state_from_file and -load_latest_state at the same time")
	}

	// TODO: find a better way to check if invalid flags were provided than checking if their actual value differs from the default value
	if subMasterData.ChiefMasterAddress != "" {
		if serverConfig.StateFileName != "" {
			log.Fatal("state_from_file flag not allowed as sub master")
		}
		if serverConfig.LoadLatestState {
			log.Fatal("load_latest_state flag not allowed as sub master")
		}
		if serverConfig.BucketWidth != 500 {
			log.Fatal("bucket_width can't be set by sub master, will be set by chief master")
		}
		if serverConfig.BigBangConfigPath != "./big_bang_config.yaml" {
			log.Fatal("big_bang_config_path flag is not allowed as sub master")
		}
	}

	if subMasterData.ChiefMasterAddress != "" {
		serverConfig.SubMaster = true
	} else {
		serverConfig.StandAloneMaster = true
	}

	s := master.NewServer(serverConfig, subMasterData)
	s.Init()
	s.Run()
}
