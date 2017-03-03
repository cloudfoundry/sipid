package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("claim", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "claim-pid")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("writes the pidfile", func() {
		pidfilePath := filepath.Join(tmpDir, "my.pid")
		cmd := exec.Command(binPath, "claim", "--pid", "17", "--pid-file", pidfilePath)

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		contents, err := ioutil.ReadFile(pidfilePath)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(contents)).To(Equal("17"))
	})
})
