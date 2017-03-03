package kill

import (
	"context"
	"os"
	"syscall"
	"time"
)

// Kill attempts to gracefully kill a process, and will violently
// kill if the process takes longer than the context deadline.
//
// If showStacks is true then SIGQUIT will be sent to attempt a
// retrieval of the process's stack before it is violently killed.
func Kill(ctx context.Context, pid int, showStacks bool) error {
	p := findProcess(pid)

	exited := p.setupExitWaiter()

	if err := p.tryKill(); err != nil {
		return err
	}

	select {
	case <-exited:
		return nil
	case <-ctx.Done():
		if showStacks {
			return p.showStacksAndKill()
		} else {
			return p.kill()
		}
	}
}

func running(process *os.Process) bool {
	err := process.Signal(syscall.Signal(0))
	return err == nil
}

func findProcess(pid int) *process {
	p, _ := os.FindProcess(pid)

	return &process{
		proc: p,
	}
}

type process struct {
	proc *os.Process
}

func (p *process) setupExitWaiter() chan struct{} {
	exited := make(chan struct{})

	go func() {
		for running(p.proc) {
			time.Sleep(10 * time.Millisecond)
		}

		close(exited)
	}()

	return exited
}

func (p *process) tryKill() error {
	// SIGTERM attempts a graceful kill, allowing a process to clean up
	return p.proc.Signal(syscall.SIGTERM)
}

func (p *process) kill() error {
	// SIGKILL kills a process immediately
	return p.proc.Signal(syscall.SIGKILL)
}

func (p *process) showStacksAndKill() error {
	// SIGQUIT prompts some frameworks to dump stack
	if err := p.proc.Signal(syscall.SIGQUIT); err != nil {
		return err
	}

	// allow time for stacks to be dumped
	time.Sleep(100 * time.Millisecond)

	// SIGKILL might error if SIGQUIT caused a quit; ignore error
	p.proc.Signal(syscall.SIGKILL)

	return nil
}
