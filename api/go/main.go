package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/public"
	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/transcoder"
	"github.com/joho/godotenv"
)

func main() {
	var server base.Server

	err := godotenv.Load("/usr/local/bin/.env")

	if err != nil {
		log.Fatal("Error loading .env files")
	}

	var serverMode string = os.Getenv("SERVER_MODE")

	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))

	if err != nil {
		log.Fatal("Error getting server port from .env " + err.Error())
	}

	var serverIP string = os.Getenv("SERVER_IP")

	switch serverMode {
	case "transcoder":
		server = transcoder.InitServer(serverPort, serverIP)

	case "api":
		server = public.InitServer(serverPort, serverIP)

	default:
		panic("Unrecognized server mode.")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		server.Start()
	}()

	fmt.Printf("Server is running on %s:%d\n", server.IP, server.Port)

	wg.Wait()
}
