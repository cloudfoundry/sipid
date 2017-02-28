package pid

import (
	"io/ioutil"
	"strconv"
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

	pid, err := strconv.Atoi(string(pidfileContents))
	if err != nil {
		return Pidfile{}, BadPidfile{msg: "Pidfile does not contain a valid PID"}
	}

	return Pidfile{pid: pid}, nil
}

func (p Pidfile) PID() int {
	return p.pid
}
