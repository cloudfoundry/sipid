package pid_test

import (
	"github.com/cloudfoundry/sipid/pid"

	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pidfile", func() {
	Context("when the file does not exist", func() {
		It("returns an IsNotExist error", func() {
			_, err := pid.NewPidfile("/tmp/bad/pidfile")
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when the pidfile does exist", func() {
		var (
			pidfile *os.File
		)

		BeforeEach(func() {
			pidfile, _ = ioutil.TempFile("", "pidfile")
		})

		Context("contains an integer", func() {
			BeforeEach(func() {
				pidfile.WriteString("7")
			})

			It("knows the PID", func() {
				p, err := pid.NewPidfile(pidfile.Name())
				Expect(err).ToNot(HaveOccurred())

				Expect(p.PID()).To(Equal(7))
			})
		})

		Context("contains an integer with some whitespace", func() {
			BeforeEach(func() {
				pidfile.WriteString("  7 \n")
			})

			It("knows the PID", func() {
				p, err := pid.NewPidfile(pidfile.Name())
				Expect(err).ToNot(HaveOccurred())

				Expect(p.PID()).To(Equal(7))
			})
		})

		Context("does not contain an integer", func() {
			BeforeEach(func() {
				pidfile.WriteString("a")
			})

			It("returns a BadPidfile error", func() {
				_, err := pid.NewPidfile(pidfile.Name())
				Expect(err).To(MatchError(ContainSubstring("does not contain a valid pid")))
			})
		})
	})
})
