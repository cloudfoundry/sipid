// +build !windows

package pid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

type processExistsError struct {
	Filename string
	PID      int
}

func (err processExistsError) Error() string {
	return fmt.Sprintf("process %d (%s) already exists", err.PID, err.Filename)
}

type processInPidfileError struct {
	Filename string
	PID      int
}

func (err processInPidfileError) Error() string {
	return fmt.Sprintf("process %d (%s) is already in the pidfile", err.PID, err.Filename)
}

func Claim(desiredPID int, pidfilePath string) error {
	if err := os.MkdirAll(filepath.Dir(pidfilePath), 0700); err != nil {
		return err
	}

	return withLock(pidfilePath, func() error {
		pidfile, err := NewPidfile(pidfilePath)
		if err == nil && pidRunning(pidfile.PID()) {
			if desiredPID == pidfile.PID() {
				return processInPidfileError{Filename: pidfilePath, PID: pidfile.PID()}
			}

			return processExistsError{Filename: pidfilePath, PID: pidfile.PID()}
		}

		return ioutil.WriteFile(pidfilePath, []byte(strconv.Itoa(desiredPID)), 0600)
	})
}

func withLock(filePath string, f func() error) error {
	lock := flock{path: filePath}
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	return f()
}

func pidRunning(pid int) bool {
	process, _ := os.FindProcess(pid)

	err := process.Signal(syscall.Signal(0))
	return err == nil
}

type flock struct {
	path string
	fh   *os.File
}

func (f *flock) Lock() error {
	var err error

	if f.fh, err = os.OpenFile(f.path, os.O_RDONLY|os.O_CREATE, 0600); err != nil {
		return err
	}

	if err = syscall.Flock(int(f.fh.Fd()), syscall.LOCK_NB|syscall.LOCK_EX); err != nil {
		return errors.New("another process is locking")
	}

	return nil
}

func (f *flock) Unlock() {
	f.fh.Close()
}
