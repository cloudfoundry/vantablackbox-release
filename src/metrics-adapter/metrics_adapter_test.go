package metricsadapter_test

import (
	"errors"
	"net/http"
	"time"

	"github.com/masters-of-cats/metricsadapter"
	fakes "github.com/masters-of-cats/metricsadapter/metrics-adapterfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("MetricsAdapter", func() {
	Describe("CollectMetrics", func() {
		var (
			server           *ghttp.Server
			collectedMetrics metricsadapter.Series
			collectErr       error
			url              string
		)

		BeforeEach(func() {
			server = ghttp.NewServer()
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/"),
				ghttp.RespondWith(http.StatusOK, `{"numGoRoutines": 19,"memstats":{"Alloc": 12345}}`),
			))
			url = server.URL()
		})

		AfterEach(func() {
			server.Close()
		})

		JustBeforeEach(func() {
			collectedMetrics, collectErr = metricsadapter.CollectMetrics(url, "cactus")
		})

		It("does not return an error", func() {
			Expect(collectErr).NotTo(HaveOccurred())
		})

		It("collects metrics from the debug server", func() {
			expected := metricsadapter.Series{
				Series: metricsadapter.Metrics{
					metricsadapter.Metric{
						Metric: "garden.numGoroutines",
						Points: metricsadapter.MetricPoints{[2]float64{float64(time.Now().Unix()), float64(19)}},
						Host:   "cactus",
						Tags:   []string{},
					},
					metricsadapter.Metric{
						Metric: "garden.memory",
						Points: metricsadapter.MetricPoints{[2]float64{float64(time.Now().Unix()), float64(12345)}},
						Host:   "cactus",
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
			wfSender       *fakes.FakeSender
			emittedMetrics metricsadapter.Series
		)

		BeforeEach(func() {
			wfSender = new(fakes.FakeSender)

			emittedMetrics = metricsadapter.Series{
				Series: metricsadapter.Metrics{
					metricsadapter.Metric{
						Metric: "garden.numGoroutines",
						Points: metricsadapter.MetricPoints{[2]float64{float64(1000), float64(1)}},
						Host:   "cactus",
						Tags:   []string{},
					},
					metricsadapter.Metric{
						Metric: "garden.memory",
						Points: metricsadapter.MetricPoints{[2]float64{float64(2000), float64(2)}},
						Host:   "cactus",
						Tags:   []string{},
					},
				},
			}
		})

		JustBeforeEach(func() {
			emitErr = metricsadapter.EmitMetrics(emittedMetrics, wfSender)
		})

		It("does not return an error", func() {
			Expect(emitErr).NotTo(HaveOccurred())
		})

		It("flushes the wavefront sender", func() {
			Expect(wfSender.FlushCallCount()).To(Equal(1))
		})

		It("posts metric in wavefront proxy format", func() {
			Expect(wfSender.SendMetricCallCount()).To(Equal(2))
			actualMetricName, actualValue, actualTimestamp, actualHost, actualTags := wfSender.SendMetricArgsForCall(0)
			Expect(actualMetricName).To(Equal("garden.numGoroutines"))
			Expect(actualValue).To(Equal(1.0))
			Expect(actualTimestamp).To(Equal(int64(1000)))
			Expect(actualHost).To(Equal("cactus"))
			Expect(actualTags).To(BeNil())

			actualMetricName, actualValue, actualTimestamp, actualHost, actualTags = wfSender.SendMetricArgsForCall(1)
			Expect(actualMetricName).To(Equal("garden.memory"))
			Expect(actualValue).To(Equal(2.0))
			Expect(actualTimestamp).To(Equal(int64(2000)))
			Expect(actualHost).To(Equal("cactus"))
			Expect(actualTags).To(BeNil())
		})

		When("the wavefront sender fails", func() {
			BeforeEach(func() {
				wfSender.SendMetricReturns(errors.New("wf-error"))
			})

			It("returns the error", func() {
				Expect(emitErr).To(MatchError("wf-error"))
			})
		})
	})
})
