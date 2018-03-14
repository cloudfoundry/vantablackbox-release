package main

import (
	"errors"
	"flag"
	"fmt"
	adapter "metrics-adapter"
	"os"
)

type flags struct {
	dataDogAPIKey       string
	gardenDebugEndpoint string
}

func initFlags() (flags, error) {
	var f flags
	flag.StringVar(&f.dataDogAPIKey, "datadog-api-key", "", "API key to Datadog account")
	flag.StringVar(&f.gardenDebugEndpoint, "garden-debug-endpoint", "", "Address of garden's debug endpoint")
	flag.Parse()

	if f.dataDogAPIKey == "" || f.gardenDebugEndpoint == "" {
		return flags{}, errors.New("please provide all flags, see help for usage")
	}

	return f, nil
}

func main() {
	f, err := initFlags()
	exitOn(err)

	datadogURL := "https://app.datadoghq.com/api/v1/series"
	datadogSeries, err := adapter.CollectMetrics(f.gardenDebugEndpoint)
	exitOn(err)

	err = adapter.EmitMetrics(datadogSeries, datadogURL, f.dataDogAPIKey)
	exitOn(err)
}

func exitOn(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
