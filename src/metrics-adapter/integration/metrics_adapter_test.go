package metrics_adapter_integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("MetricsAdapterIntegration", func() {
	var (
		session           *gexec.Session
		cmd               *exec.Cmd
		gardenDebugServer *httptest.Server
	)

	BeforeEach(func() {
		gardenDebugServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "{\"numGoRoutines\": 19}")
		}))

		cmd = exec.Command(metricsBinPath, "--datadog-api-key", "foo", "--garden-debug-endpoint", gardenDebugServer.URL)
	})

	JustBeforeEach(func() {
		session = gexecStart(cmd)
	})

	AfterEach(func() {
		gardenDebugServer.Close()
	})

	Context("when the datadog-api-key is omitted", func() {
		BeforeEach(func() {
			cmd = exec.Command(metricsBinPath, "--garden-debug-endpoint", gardenDebugServer.URL, "--host", "bar")
		})

		It("fails", func() {
			Expect(session.Wait()).NotTo(gexec.Exit(0))
		})
	})

	Context("when the garden-debug-endpoint is omitted", func() {
		BeforeEach(func() {
			cmd = exec.Command(metricsBinPath, "--datadog-api-key", "foo", "--host", "bar")
		})

		It("fails", func() {
			Expect(session.Wait()).NotTo(gexec.Exit(0))
		})
	})

	Context("when the host is omitted", func() {
		BeforeEach(func() {
			cmd = exec.Command(metricsBinPath, "--datadog-api-key", "foo", "--garden-debug-endpoint", gardenDebugServer.URL)
		})

		It("fails", func() {
			Expect(session.Wait()).NotTo(gexec.Exit(0))
		})
	})
})
