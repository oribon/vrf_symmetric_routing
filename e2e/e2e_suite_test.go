package e2e

import (
	"flag"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func handleFlags() {
	flag.StringVar(&externalHostIP, "external-host-ip", "", "The IP of the external host to curl against and that would curl the services")
	flag.Parse()
}

func TestMain(m *testing.M) {
	// Register test flags, then parse flags.
	handleFlags()
	if testing.Short() {
		return
	}

	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	externalIP := net.ParseIP(externalHostIP)
	Expect(externalIP).ToNot(gomega.BeNil(), "could not parse external host ip: ", externalHostIP)

	ifaces, err := net.Interfaces()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	found := false
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.Equal(externalIP) {
				found = true
				break
			}
		}
	}
	gomega.Expect(found).To(BeTrue(), fmt.Sprintf("the host running the test is not the external host, did not find ip %s locally", externalHostIP))
})
