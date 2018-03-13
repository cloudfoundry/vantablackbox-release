package metrics_adapter_integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("MetricsAdapterIntegration", func() {
	var (
		session *gexec.Session
		cmd     *exec.Cmd
	)

	BeforeEach(func() {
		cmd = exec.Command(metricsBinPath, "--datadog-api-key", "foo", "--garden-debug-endpoint", "localhost/foo")
	})

	JustBeforeEach(func() {
		session = gexecStart(cmd)
	})

	It("does not fail", func() {
		Expect(session.Wait()).To(gexec.Exit(0))
	})

	Context("when the datadog-api-key is omitted", func() {
		BeforeEach(func() {
			cmd = exec.Command(metricsBinPath, "--garden-debug-endpoint", "localhost/foo")
		})

		It("fails", func() {
			Expect(session.Wait()).NotTo(gexec.Exit(0))
		})
	})

	Context("when the garden-debug-endpoint is omitted", func() {
		BeforeEach(func() {
			cmd = exec.Command(metricsBinPath, "--datadog-api-key", "foo")
		})

		It("fails", func() {
			Expect(session.Wait()).NotTo(gexec.Exit(0))
		})
	})
})
