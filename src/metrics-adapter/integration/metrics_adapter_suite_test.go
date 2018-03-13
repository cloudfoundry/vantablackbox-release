package metrics_adapter_integration_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMetricsAdapterIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricsAdapter Integration Suite")
}

var metricsBinPath string

var _ = BeforeSuite(func() {
	metricsBinPath = gexecBuild("metrics-adapter/cmd/metrics-adapter")
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func gexecBuild(path string) string {
	binPath, err := gexec.Build(path)
	Expect(err).NotTo(HaveOccurred())
	return binPath
}

func gexecStart(cmd *exec.Cmd) *gexec.Session {
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
