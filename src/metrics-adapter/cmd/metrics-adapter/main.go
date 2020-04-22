package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/masters-of-cats/metricsadapter"
)

type flags struct {
	dataDogAPIKey       string
	gardenDebugEndpoint string
	host                string
}

func initFlags() (flags, error) {
	var f flags
	flag.StringVar(&f.dataDogAPIKey, "datadog-api-key", "", "API key to Datadog account")
	flag.StringVar(&f.gardenDebugEndpoint, "garden-debug-endpoint", "", "Address of garden's debug endpoint")
	flag.StringVar(&f.host, "host", "", "Name of the host VM")
	flag.Parse()

	if f.dataDogAPIKey == "" || f.gardenDebugEndpoint == "" || f.host == "" {
		return flags{}, errors.New("please provide all flags, see help for usage")
	}

	return f, nil
}

func main() {
	f, err := initFlags()
	exitOn(err)

	datadogURL := "https://app.datadoghq.com/api/v1/series"
	datadogSeries, err := metricsadapter.CollectMetrics(f.gardenDebugEndpoint, f.host)
	exitOn(err)

	err = metricsadapter.EmitMetrics(datadogSeries, datadogURL, f.dataDogAPIKey)
	exitOn(err)
}

func exitOn(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
