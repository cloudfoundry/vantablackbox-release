package main_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var (
		vantaagentCmd *exec.Cmd

		gardenDebugServer *httptest.Server
		statsdServer      net.Listener
		statsdEndpoint    string
		pollingInterval   int

		statsdServerPayload chan string
		vantaagentBinary    string

		session *gexec.Session
	)

	BeforeEach(func() {
		pollingInterval = 1
		statsdServerPayload = make(chan string)

		gardenDebugServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "{\"numGoRoutines\": 19}")
		}))

		var err error
		statsdServer, err = net.Listen("tcp", "localhost:0")
		Expect(err).NotTo(HaveOccurred())
		go fakeStatsdAcceptor(statsdServer, statsdServerPayload)

		statsdEndpoint = statsdServer.Addr().String()

		vantaagentBinary, err = gexec.Build("vantaagent")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		vantaagentCmd = exec.Command(vantaagentBinary,
			"--interval", fmt.Sprintf("%d", pollingInterval),
			"--statsd", statsdEndpoint,
			"--garden", gardenDebugServer.URL)
		var err error
		session, err = gexec.Start(vantaagentCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		session.Terminate()
		gardenDebugServer.Close()
		Expect(statsdServer.Close()).To(Succeed())
	})

	Describe("Parameters validation", func() {
		Context("when all parameters are provided", func() {
			It("should not exit", func() {
				Consistently(session).ShouldNot(gexec.Exit())
			})
		})

		Context("when no parameters are provided", func() {
			JustBeforeEach(func() {
				vantaagentCmd = exec.Command(vantaagentBinary)
				var err error
				session, err = gexec.Start(vantaagentCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should exit with non-zero code", func() {
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when a single parameters is provided", func() {
			JustBeforeEach(func() {
				vantaagentCmd = exec.Command(vantaagentBinary,
					"--garden", gardenDebugServer.URL)
				var err error
				session, err = gexec.Start(vantaagentCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should exit with non-zero code", func() {
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})

	Describe("Goroutines count", func() {
		It("reports goroutines count", func() {
			Eventually(statsdServerPayload).Should(Receive(Equal("asd")))
		})
	})

})

func fakeStatsdAcceptor(server net.Listener, ch chan string) {
	defer GinkgoRecover()
	for {
		conn, err := server.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		panic("foo")

		payload, err := ioutil.ReadAll(conn)
		Expect(err).NotTo(HaveOccurred())

		ch <- string(payload)
	}
}
