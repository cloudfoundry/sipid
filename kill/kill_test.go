package kill_test

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/sipid/kill"
	"strconv"
)

var _ = Describe("Kill", func() {
	var (
		processPath string
		pidfilePath string

		process *os.Process
		stderr  *gbytes.Buffer
	)

	JustBeforeEach(func() {
		process, stderr = startFixture(processPath)

		pidfile, err := ioutil.TempFile("", "pidfile")
		Expect(err).ToNot(HaveOccurred())
		pidfile.WriteString(strconv.Itoa(process.Pid))
		pidfile.Close()

		pidfilePath = pidfile.Name()
	})

	Context("when the process is not running", func() {
		BeforeEach(func() {
			processPath = easyPath
		})

		JustBeforeEach(func() {
			process.Kill()
			process.Wait()
		})

		It("does not return an error", func() {
			ctx := context.Background()

			err := kill.Kill(ctx, pidfilePath, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the process we're killing goes away easily", func() {
		BeforeEach(func() {
			processPath = easyPath
		})

		It("stops the process", func() {
			ctx := context.Background()

			err := kill.Kill(ctx, pidfilePath, false)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())
		})

		It("removes the pidfile", func() {
			ctx := context.Background()

			err := kill.Kill(ctx, pidfilePath, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(pidfilePath).NotTo(BeAnExistingFile())
		})
	})

	Context("when the process we're killing doesn't go away easily", func() {
		BeforeEach(func() {
			processPath = hardPath
		})

		It("eventually goes away too after we become more violent", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := kill.Kill(ctx, pidfilePath, false)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())

			Expect(stderr).ToNot(gbytes.Say("SIGQUIT: quit"))
		})

		It("will show the stacks if desired", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := kill.Kill(ctx, pidfilePath, true)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())

			Expect(stderr).To(gbytes.Say("SIGQUIT: quit"))
		})

		It("removes the pidfile", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := kill.Kill(ctx, pidfilePath, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(pidfilePath).NotTo(BeAnExistingFile())
		})
	})
})

func startFixture(path string) (*os.Process, *gbytes.Buffer) {
	cmd := exec.Command(path)
	stdout := gbytes.NewBuffer()
	stderr := gbytes.NewBuffer()
	cmd.Stdout = io.MultiWriter(stdout, GinkgoWriter)
	cmd.Stderr = io.MultiWriter(stderr, GinkgoWriter)

	go func() {
		defer GinkgoRecover()

		cmd.Run()
	}()

	// make sure process is running and signal handler has been installed
	Eventually(stdout).Should(gbytes.Say("Running as"))

	return cmd.Process, stderr
}

func running(process *os.Process) func() bool {
	return func() bool {
		err := process.Signal(syscall.Signal(0))
		return err == nil
	}
}
