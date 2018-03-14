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
				Series: adapter.DatadogMetrics{
					adapter.DatadogMetric{
						Metric: "garden.numGoroutines",
						Points: adapter.DatadogMetricPoints{[2]float64{float64(time.Now().Unix()), float64(19)}},
						Host:   "",
						Tags:   []string{},
					},
					adapter.DatadogMetric{
						Metric: "garden.memory",
						Points: adapter.DatadogMetricPoints{[2]float64{float64(time.Now().Unix()), float64(12345)}},
						Host:   "",
						Tags:   []string{},
					},
				},
			}

			Expect(expected.Series).To(Equal(collectedMetrics.Series))
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
				Series: adapter.DatadogMetrics{
					adapter.DatadogMetric{
						Metric: "garden.numGoroutines",
						Points: adapter.DatadogMetricPoints{[2]float64{float64(time.Now().Unix()), float64(1)}},
						Host:   "",
						Tags:   []string{},
					},
					adapter.DatadogMetric{
						Metric: "garden.memory",
						Points: adapter.DatadogMetricPoints{[2]float64{float64(time.Now().Unix()), float64(2)}},
						Host:   "",
						Tags:   []string{},
					},
				},
			}
		)

		BeforeEach(func() {
			server.AppendHandlers(ghttp.CombineHandlers(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				GinkgoRecover()
				body = readAll(r.Body)
			}),
				ghttp.RespondWith(http.StatusAccepted, `u r cool`),
			))
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
			Expect(receivedMetrics.Series).To(Equal(emittedMetrics.Series))
		})

		It("encodes the api_key in the request URL", func() {
			Expect(server.ReceivedRequests()[0].URL.Query().Get("api_key")).To(Equal("abc"))
		})

		Context("when the HTTP response code is not a 202", func() {
			BeforeEach(func() {
				server.Reset()
				server.AppendHandlers(ghttp.RespondWith(http.StatusServiceUnavailable, `not here`))
			})

			It("returns an error", func() {
				Expect(emitErr).To(MatchError("expected 202 response but got 503"))
			})
		})
	})
})
