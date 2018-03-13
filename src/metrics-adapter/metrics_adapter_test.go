package adapter_test

import (
	"encoding/json"
	adapter "metrics-adapter"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("MetricsAdapter", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CollectMetrics", func() {
		var (
			collectedMetrics adapter.DatadogSeries
			collectErr       error
			url              string
		)

		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/"),
				ghttp.RespondWith(http.StatusOK, `{"numGoRoutines": 19,"memstats":{"Alloc": 12345}}`),
			))
			url = server.URL()
		})

		JustBeforeEach(func() {
			collectedMetrics, collectErr = adapter.CollectMetrics(url)
		})

		It("does not return an error", func() {
			Expect(collectErr).NotTo(HaveOccurred())
		})

		It("collects metrics from the debug server", func() {
			expected := adapter.DatadogSeries{
				adapter.DatadogMetric{
					Metric: "garden.numGoroutines",
					Points: []adapter.DatadogMetricPoint{{Timestamp: time.Now(), Value: 19.0}},
					Host:   "",
					Tags:   []string{},
				},
				adapter.DatadogMetric{
					Metric: "garden.memory",
					Points: []adapter.DatadogMetricPoint{{Timestamp: time.Now(), Value: 12345.0}},
					Host:   "",
					Tags:   []string{},
				},
			}

			expectMetricsToBeEqual(collectedMetrics, expected)
		})

		Context("when getting the metrics fails", func() {
			BeforeEach(func() {
				url = "foo"
			})

			It("returns an error", func() {
				Expect(collectErr).To(HaveOccurred())
			})
		})

		Context("when the response is not valid JSON", func() {
			BeforeEach(func() {
				server.Reset()
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/fail"),
					ghttp.RespondWith(http.StatusOK, `totally not json`),
				))
				url = server.URL() + "/fail"
			})

			It("returns an error", func() {
				Expect(collectErr).To(HaveOccurred())
			})
		})
	})

	Describe("EmitMetrics", func() {
		var (
			emitErr        error
			body           []byte
			emittedMetrics = adapter.DatadogSeries{
				adapter.DatadogMetric{
					Metric: "garden.numGoroutines",
					Points: []adapter.DatadogMetricPoint{{Timestamp: time.Now(), Value: 1.0}},
					Host:   "",
					Tags:   []string{},
				},
				adapter.DatadogMetric{
					Metric: "garden.memory",
					Points: []adapter.DatadogMetricPoint{{Timestamp: time.Now(), Value: 1.0}},
					Host:   "",
					Tags:   []string{},
				},
			}
		)

		BeforeEach(func() {
			server.AppendHandlers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				GinkgoRecover()
				body = readAll(r.Body)
			}))
		})

		JustBeforeEach(func() {
			emitErr = adapter.EmitMetrics(emittedMetrics, server.URL()+"/emit", "abc")
		})

		It("does not return an error", func() {
			Expect(emitErr).NotTo(HaveOccurred())
		})

		It("posts json", func() {
			Expect(server.ReceivedRequests()[0].Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("emits valid metrics", func() {
			var receivedMetrics adapter.DatadogSeries
			Expect(json.Unmarshal(body, &receivedMetrics)).To(Succeed())
			expectMetricsToBeEqual(receivedMetrics, emittedMetrics)
		})

		It("encodes the api_key in the request URL", func() {
			Expect(server.ReceivedRequests()[0].URL.Query().Get("api_key")).To(Equal("abc"))
		})
	})
})
