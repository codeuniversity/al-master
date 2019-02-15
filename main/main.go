package main

import master "github.com/codeuniversity/al-master"

const (
	address    = "localhost:5000"
	threads    = 10
	bufferSize = threads
	port       = 4000
)

func main() {
	s := master.NewServer(address, threads, bufferSize, port)
	s.InitUniverse()
	s.Run()
}
