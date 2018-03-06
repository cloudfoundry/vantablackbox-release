package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVantaagent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vantaagent Suite")
}
