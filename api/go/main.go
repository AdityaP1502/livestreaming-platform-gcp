package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/transcoder"
)

func isFlagValid(mode string, port int, ip string) error {
	// TODO: Create an error when the flag is wrong
	return nil
}

func main() {
	var serverMode string
	var serverPort int
	var serverIP string

	flag.StringVar(&serverMode, "server-mode", "", "The server mode that going to be run (required)")
	flag.StringVar(&serverIP, "ip", "", "The server ip (required)")
	flag.IntVar(&serverPort, "port", -1, "The server port (required)")

	flag.Parse()

	err := isFlagValid(serverMode, serverPort, serverIP)

	// TODO: Create a better exit when there are errors
	if err != nil {
		os.Exit(1)
	}

	server := transcoder.InitServer(serverPort, serverIP)
	go server.Start()
	fmt.Printf("Server is running on %s:%d\n", server.IP, server.Port)

	select {}
}
