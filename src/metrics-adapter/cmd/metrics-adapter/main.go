package main

import (
	"errors"
	"flag"
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
		return flags{}, errors.New("blah")
	}

	return f, nil
}

func main() {
	f, err := initFlags()
	if err != nil {
		os.Exit(1)
	}

	datadogURL := "https://app.datadoghq.com/api/v1/series"
	datadogSeries, err := adapter.CollectMetrics(f.gardenDebugEndpoint)
	if err != nil {
		panic(err)
	}
	if err := adapter.EmitMetrics(datadogSeries, datadogURL, f.dataDogAPIKey); err != nil {
		panic(err)
	}
}
