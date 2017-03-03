package pid_test

import (
	"github.com/cloudfoundry/sipid/pid"

	"io/ioutil"
	"os"
	"strconv"

	"path/filepath"

	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Claim", func() {
	var (
		pidfilePath string
	)

	BeforeEach(func() {
		dir, _ := ioutil.TempDir("", "pidfile")
		pidfilePath = filepath.Join(dir, "my.pid")
	})

	itWritesThePid := func() {
		It("writes the pid", func() {
			err := pid.Claim(17, pidfilePath)
			Expect(err).ToNot(HaveOccurred())

			pidfileContents, err := ioutil.ReadFile(pidfilePath)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(pidfileContents)).To(Equal(strconv.Itoa(17)))
		})
	}

	Context("when the pidfile does not exist", func() {
		itWritesThePid()

		Context("when the directory does not exist", func() {
			BeforeEach(func() {
				os.RemoveAll(filepath.Dir(pidfilePath))
			})

			itWritesThePid()
		})
	})

	Context("when the pidfile does exist", func() {
		Context("when the pidfile is empty", func() {
			BeforeEach(func() {
				ioutil.WriteFile(pidfilePath, []byte{}, 0600)
			})

			itWritesThePid()
		})

		Context("when the pidfile is not empty", func() {
			Context("when the PID is a running process", func() {
				preexistingPid := os.Getpid()
				newPid := os.Getpid() - 1

				BeforeEach(func() {
					ioutil.WriteFile(pidfilePath, []byte(strconv.Itoa(preexistingPid)), 0600)
				})

				It("returns an error", func() {
					err := pid.Claim(newPid, pidfilePath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("process %d (%s) already exists", preexistingPid, pidfilePath))
				})

				Context("when the PID is yourself", func() {
					It("returns an error", func() {
						err := pid.Claim(preexistingPid, pidfilePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("process %d (%s) is already in the pidfile", preexistingPid, pidfilePath))
					})
				})
			})

			Context("when the PID is not a running process", func() {
				BeforeEach(func() {
					ioutil.WriteFile(pidfilePath, []byte("-100"), 0600)
				})

				itWritesThePid()
			})
		})
	})

	Context("locking", func() {
		Context("when the file is already locked", func() {
			var (
				fh *os.File
			)

			BeforeEach(func() {
				err := ioutil.WriteFile(pidfilePath, []byte("18"), 0600)
				Expect(err).NotTo(HaveOccurred())

				fh, err = os.Open(pidfilePath)
				Expect(err).NotTo(HaveOccurred())

				err = syscall.Flock(int(fh.Fd()), syscall.LOCK_NB|syscall.LOCK_EX)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				fh.Close()
			})

			It("fails", func() {
				err := pid.Claim(17, pidfilePath)
				Expect(err).To(MatchError(ContainSubstring("another process is locking")))
			})
		})
	})
})
