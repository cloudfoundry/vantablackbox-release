package main

import (
	"flag"
	"os"

	"github.com/cactus/go-statsd-client/statsd"
)

func main() {
	var gardenEndpoint string
	var statsdEndpoint string
	var pollingInterval int

	flag.StringVar(&gardenEndpoint, "garden", "", "Enpoint of the garden server debug API")
	flag.StringVar(&statsdEndpoint, "statsd", "", "Enpoint of the statsd API")
	flag.IntVar(&pollingInterval, "interval", 0, "Poll interval, seconds")
	flag.Parse()

	if gardenEndpoint == "" || statsdEndpoint == "" || pollingInterval <= 0 {
		os.Exit(1)
	}

	client, err := statsd.NewClient(statsdEndpoint, "test-client")
	if err != nil {
		panic("Error creating statsd client: " + err.Error())
	}
	defer client.Close()

	if err := client.SetInt("goroutines", 19, 1); err != nil {
		panic("Error setting goroutines count: " + err.Error())
	}
	for {
	}
}
