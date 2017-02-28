package pid

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func Guard(pidfilePath string) error {
	pidfile, err := NewPidfile(pidfilePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		err = os.Remove(pidfilePath)
		if err != nil {
			return err
		}
		return nil
	}

	process, err := os.FindProcess(pidfile.PID())
	if err != nil {
		return errors.New("Windows not supported")
	}

	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return fmt.Errorf("Process %d already exists", pidfile.PID())
	}

	err = os.Remove(pidfilePath)
	if err != nil {
		return err
	}

	return nil
}
