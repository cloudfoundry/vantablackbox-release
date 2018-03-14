package adapter_test

import (
	"io"
	"io/ioutil"
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
