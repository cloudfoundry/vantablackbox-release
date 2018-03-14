package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type GardenMemStats struct {
	Alloc float64 `json:"Alloc"`
}

type GardenDebugMetrics struct {
	NumGoroutines int            `json:"numGoroutines"`
	Memstats      GardenMemStats `json:"memstats"`
}

type DatadogMetrics []DatadogMetric

type DatadogSeries struct {
	Series DatadogMetrics `json:"series"`
}

type DatadogMetric struct {
	Metric string              `json:"metric"`
	Points DatadogMetricPoints `json:"points"`
	Host   string              `json:"host"`
	Tags   []string            `json:"tags"`
}

type DatadogMetricPoints [][2]float64

func fromGardenDebugMetrics(m GardenDebugMetrics) DatadogSeries {
	now := time.Now().Unix()
	return DatadogSeries{
		Series: DatadogMetrics{
			DatadogMetric{
				Metric: "garden.numGoroutines",
				Points: DatadogMetricPoints{[2]float64{float64(now), float64(m.NumGoroutines)}},
				Host:   "",
				Tags:   []string{},
			},
			DatadogMetric{
				Metric: "garden.memory",
				Points: DatadogMetricPoints{[2]float64{float64(now), float64(m.Memstats.Alloc)}},
				Host:   "",
				Tags:   []string{},
			},
		},
	}
}

func CollectMetrics(url string) (DatadogSeries, error) {
	body, err := getResponseBody(url)
	if err != nil {
		return DatadogSeries{}, err
	}

	var gardenDebugMetrics GardenDebugMetrics
	err = json.Unmarshal(body, &gardenDebugMetrics)
	if err != nil {
		return DatadogSeries{}, err
	}

	return fromGardenDebugMetrics(gardenDebugMetrics), nil
}

func getResponseBody(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

func EmitMetrics(metrics DatadogSeries, url, apiKey string) error {
	content, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	res, err := http.Post(fmt.Sprintf("%s?api_key=%s", url, apiKey), "application/json", bytes.NewBuffer(content))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("expected %d response but got %d", http.StatusAccepted, res.StatusCode)
	}

	return nil
}
