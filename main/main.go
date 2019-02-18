package main

import master "github.com/codeuniversity/al-master"

const (
	bufferSize = 1000
	httpPort   = 4000
	grpcPort   = 3000
)

func main() {
	s := master.NewServer(bufferSize, httpPort, grpcPort)
	s.Init()
	s.Run()
}
