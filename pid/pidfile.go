package pid

import (
	"io/ioutil"
	"strconv"
	"strings"
)

type BadPidfile struct {
	msg string
}

func (e BadPidfile) Error() string {
	return e.msg
}

type Pidfile struct {
	pid int
}

func NewPidfile(pidfilePath string) (Pidfile, error) {
	pidfileContents, err := ioutil.ReadFile(pidfilePath)
	if err != nil {
		return Pidfile{}, err
	}

	// TODO: test
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidfileContents)))
	if err != nil {
		return Pidfile{}, BadPidfile{msg: "pidfile does not contain a valid PID"}
	}

	return Pidfile{pid: pid}, nil
}

func (p Pidfile) PID() int {
	return p.pid
}
