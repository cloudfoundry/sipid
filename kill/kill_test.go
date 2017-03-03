package kill_test

import (
	"context"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"io"
	"os"
	"syscall"

	"github.com/cloudfoundry/sipid/kill"
)

var _ = Describe("Kill", func() {
	var (
		path string

		process *os.Process
		stderr  *gbytes.Buffer
	)

	JustBeforeEach(func() {
		process, stderr = startFixture(path)
	})

	//AfterEach(func() {
	//	// really make sure we do not leak anything
	//	process.Kill()
	//
	//})

	Context("when the process we're killing goes away easily", func() {
		BeforeEach(func() {
			path = easyPath
		})

		It("stops the process", func() {
			ctx := context.Background()

			err := kill.Kill(ctx, process.Pid, false)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())
		})
	})

	Context("when the process we're killing doesn't go away easily", func() {
		BeforeEach(func() {
			path = hardPath
		})

		It("eventually goes away too after we become more violent", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := kill.Kill(ctx, process.Pid, false)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())

			Expect(stderr).ToNot(gbytes.Say("SIGQUIT: quit"))
		})

		It("will show the stacks if desired", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			err := kill.Kill(ctx, process.Pid, true)
			Expect(err).NotTo(HaveOccurred())

			Eventually(running(process)).Should(BeFalse())

			Expect(stderr).To(gbytes.Say("SIGQUIT: quit"))
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
