package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudfoundry/sipid/kill"
	"github.com/cloudfoundry/sipid/pid"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "claim":
		claimCmd()
	case "kill":
		killCmd()
	default:
		usage()
	}
}

func claimCmd() {
	var desiredPID int
	var pidfilePath string

	flag.IntVar(&desiredPID, "pid", -1, "process ID to write to pid-file")
	flag.StringVar(&pidfilePath, "pid-file", "", "pid-file to write process ID to")
	flag.CommandLine.Parse(os.Args[2:])

	if desiredPID == -1 {
		usage()
	}

	if pidfilePath == "" {
		usage()
	}

	if err := pid.Claim(desiredPID, pidfilePath); err != nil {
		log.Printf("claim failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func killCmd() {
	var pidfilePath string
	var showStacks bool

	flag.StringVar(&pidfilePath, "pid-file", "", "pid-file containing process id to kill")
	flag.BoolVar(&showStacks, "show-stacks", false, "when true, try to dump the stack")
	flag.CommandLine.Parse(os.Args[2:])

	if pidfilePath == "" {
		usage()
	}

	desiredPID, err := pid.NewPidfile(pidfilePath)
	if err != nil {
		log.Printf("kill failed: %s\n", err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := kill.Kill(ctx, desiredPID.PID(), showStacks); err != nil {
		log.Printf("kill failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func usage() {
	claimUsage := fmt.Sprintf("%s claim --pid PID --pid-file PID_FILE", os.Args[0])
	killUsage := fmt.Sprintf("%s kill --pid PID [--show-stacks]", os.Args[0])

	fmt.Fprintf(os.Stderr, "usage: %s\n       %s\n", claimUsage, killUsage)
	os.Exit(1)
}
