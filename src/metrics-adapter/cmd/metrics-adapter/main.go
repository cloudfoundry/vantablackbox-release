package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/masters-of-cats/metricsadapter"
)

type flags struct {
	gardenDebugEndpoint string
	host                string
	wavefrontProxyPort  int
}

func initFlags() (flags, error) {
	var f flags
	flag.StringVar(&f.gardenDebugEndpoint, "garden-debug-endpoint", "", "Address of garden's debug endpoint")
	flag.StringVar(&f.host, "host", "", "Name of the host VM")
	flag.IntVar(&f.wavefrontProxyPort, "wavefront-proxy-port", 0, "Wavefront Proxy port")
	flag.Parse()

	if f.wavefrontProxyPort == 0 || f.gardenDebugEndpoint == "" || f.host == "" {
		return flags{}, errors.New("please provide all flags, see help for usage")
	}

	return f, nil
}

func main() {
	f, err := initFlags()
	exitOn(err)

	series, err := metricsadapter.CollectMetrics(f.gardenDebugEndpoint, f.host)
	exitOn(err)

	err = metricsadapter.EmitMetrics(series, fmt.Sprintf("localhost:%d", f.wavefrontProxyPort))
	exitOn(err)
}

func exitOn(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
