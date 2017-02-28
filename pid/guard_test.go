package pid_test

import (
	"github.com/cloudfoundry/sipid/pid"

	"io/ioutil"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Guard", func() {
	Context("when the pidfile does not exist", func() {
		It("does not return an error", func() {
			err := pid.Guard("/tmp/bad/pidfile")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when the pidfile does exist", func() {
		var (
			pidfile *os.File
		)

		BeforeEach(func() {
			pidfile, _ = ioutil.TempFile("", "pidfile")
		})

		Context("when the pidfile is empty", func() {
			It("removes the pidfile", func() {
				_, err := os.Stat(pidfile.Name())
				Expect(os.IsNotExist(err)).To(BeFalse())

				err = pid.Guard(pidfile.Name())
				Expect(err).ToNot(HaveOccurred())

				_, err = os.Stat(pidfile.Name())
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})

		Context("when the pidfile is not empty", func() {
			Context("when the PID is a running process", func() {
				BeforeEach(func() {
					pidfile.WriteString(strconv.Itoa(os.Getpid()))
				})

				It("returns an error", func() {
					err := pid.Guard(pidfile.Name())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Process %d already exists", os.Getpid()))
				})

				It("does not remove the pidfile", func() {
					pid.Guard(pidfile.Name())

					_, err := os.Stat(pidfile.Name())
					Expect(os.IsNotExist(err)).To(BeFalse())
				})
			})

			Context("when the PID is not a running process", func() {
				BeforeEach(func() {
					pidfile.WriteString("-100")
				})

				It("removes the pidfile", func() {
					_, err := os.Stat(pidfile.Name())
					Expect(os.IsNotExist(err)).To(BeFalse())

					err = pid.Guard(pidfile.Name())
					Expect(err).ToNot(HaveOccurred())

					_, err = os.Stat(pidfile.Name())
					Expect(os.IsNotExist(err)).To(BeTrue())
				})
			})
		})
	})
})
