package pid

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type BadPidfileError struct {
	pidfile string
}

func (e BadPidfileError) Error() string {
	return fmt.Sprintf("pidfile (%s) does not contain a valid pid")
}

type Pidfile struct {
	pid int
}

func NewPidfile(pidfilePath string) (Pidfile, error) {
	pidfileContents, err := ioutil.ReadFile(pidfilePath)
	if err != nil {
		return Pidfile{}, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidfileContents)))
	if err != nil {
		return Pidfile{}, BadPidfileError{pidfile: pidfilePath}
	}

	return Pidfile{pid: pid}, nil
}

func (p Pidfile) PID() int {
	return p.pid
}
