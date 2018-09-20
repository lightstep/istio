package main

import (
	"flag"
	"fmt"
	"istio.io/istio/mixer/adapter/lightstep"
	"math/rand"
	"os"
	"time"
)

func main() {
	serverAddress := flag.String("serverAddress", "", "the location the server listens on (host:port)")
	accessToken := flag.String("accessToken", "", "the access token used to communicate with the satellite pool")
	socketAddress := flag.String("socketAddress", "", "the upstream location of the satellite pool")
	flag.Parse()

	fmt.Println("The provided flags include...")
	fmt.Println("serverAddress:", *serverAddress)
	fmt.Println("accessToken:", *accessToken)
	fmt.Println("socketAddress:", *socketAddress)

	var flagErrs error
	lightstep.AppendErrors(flagErrs, lightstep.ValidateServerAddress(*serverAddress))
	lightstep.AppendErrors(flagErrs, lightstep.ValidateAccessToken(*accessToken))
	lightstep.AppendErrors(flagErrs, lightstep.ValidateSocketAddress(*socketAddress))
	if flagErrs != nil {
		fmt.Println(flagErrs)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Generate a randomized reporterID
	seededGUIDGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	reporterID := uint64(seededGUIDGen.Int63())

	s, err := lightstep.NewLightStepAdapter(
		lightstep.AdapterOptions{
			Server: lightstep.ServerOptions{
				Address: *serverAddress,
			},
			Client: lightstep.ClientOptions{
				ReporterID:    reporterID,
				AccessToken:   *accessToken,
				SocketAddress: *socketAddress,
			},
		})
	if err != nil {
		fmt.Println("unable to start adapter:", err)
		os.Exit(1)
	}

	shutdown := make(chan error, 1)
	go func() {
		s.Run(shutdown)
	}()
	_ = <-shutdown
}
