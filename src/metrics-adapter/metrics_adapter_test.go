package metricsadapter_test

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/masters-of-cats/metricsadapter"
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
			server         net.Listener
			addr           string
			requestsChan   chan string
			emittedMetrics metricsadapter.Series
		)

		BeforeEach(func() {
			var err error
			server, err = net.Listen("tcp", "localhost:0")
			Expect(err).NotTo(HaveOccurred())
			addr = server.Addr().String()

			requestsChan = make(chan string)

			go func() {
				defer GinkgoRecover()

				for {
					conn, err := server.Accept()
					if err != nil {
						close(requestsChan)
						return
					}

					Expect(err).NotTo(HaveOccurred())

					go func() {
						defer GinkgoRecover()

						connReader := bufio.NewReader(conn)
						for {
							line, err := connReader.ReadString('\n')
							if err == io.EOF {
								return
							}
							Expect(err).NotTo(HaveOccurred())
							requestsChan <- line
						}
					}()
				}
			}()

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

		AfterEach(func() {
			server.Close()
			Eventually(requestsChan).Should(BeClosed())
		})

		JustBeforeEach(func() {
			emitErr = metricsadapter.EmitMetrics(emittedMetrics, addr)
		})

		It("does not return an error", func() {
			Expect(emitErr).NotTo(HaveOccurred())
		})

		It("posts metric in wavefront proxy format", func() {
			var metricsLine string
			Eventually(requestsChan).Should(Receive(&metricsLine))
			Expect(metricsLine).To(Equal(fmt.Sprintf("garden.numGoroutines %f %f source=cactus\n", 1.0, 1000.0)))

			Eventually(requestsChan).Should(Receive(&metricsLine))
			Expect(metricsLine).To(Equal(fmt.Sprintf("garden.memory %f %f source=cactus\n", 2.0, 2000.0)))
		})
	})
})
