package main

import (
	"fmt"
	"os"
	"strconv"

	"istio.io/istio/mixer/adapter/lightstep"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Printf("Missing command line args. Please provide serverAddress reporterID accessToken hostPort")
		os.Exit(-1)
	}
	serverAddr := os.Args[1]
	reporterID, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		fmt.Printf("reporterID is not a valid uint64: %v", err)
		os.Exit(-1)
	}
	accessToken := os.Args[3]
	hostPort := os.Args[4]

	fmt.Println(os.Args[1])
	fmt.Println(os.Args[2])
	fmt.Println(os.Args[3])
	fmt.Println(os.Args[4])


	s, err := lightstep.NewLightStepAdapter(
		lightstep.AdapterOptions{
			Server: lightstep.ServerOptions{
				Address: serverAddr,
			},
			Client: lightstep.ClientOptions{
				ReporterID:  reporterID,
				AccessToken: accessToken,
				HostPort:    hostPort,
			},
		})
	if err != nil {
		fmt.Printf("unable to start adapter: %v", err)
		os.Exit(-1)
	}

	shutdown := make(chan error, 1)
	go func() {
		s.Run(shutdown)
	}()
	_ = <-shutdown
}
