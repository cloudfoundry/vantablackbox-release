package adapter_test

import (
	"io"
	"io/ioutil"
	adapter "metrics-adapter"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetricsAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricsAdapter Suite")
}

func readAll(r io.Reader) []byte {
	content, err := ioutil.ReadAll(r)
	Expect(err).NotTo(HaveOccurred())
	return content
}

func expectMetricsToBeEqual(a, b adapter.DatadogSeries) {
	Expect(a).To(HaveLen(len(b)))
	for i := range a {
		Expect(a[i].Metric).To(Equal(b[i].Metric))
		Expect(a[i].Points[0].Value).To(Equal(b[i].Points[0].Value))
	}
}
