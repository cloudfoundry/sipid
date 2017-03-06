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
	"github.com/cloudfoundry/sipid/poll"
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
	case "wait-until-healthy":
		waitCmd()
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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := kill.Kill(ctx, pidfilePath, showStacks); err != nil {
		log.Printf("kill failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func waitCmd() {
	var healthcheckURL string
	var timeout time.Duration
	var pollingFrequency time.Duration

	flag.StringVar(&healthcheckURL, "url", "", "URL to poll for system health")
	flag.DurationVar(&timeout, "timeout", time.Minute, "timeout for system to become healthy")
	flag.DurationVar(&pollingFrequency, "polling-frequency", 5 * time.Second, "frequency to poll healthcheck endpoint")
	flag.CommandLine.Parse(os.Args[2:])

	if healthcheckURL == "" {
		usage()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := poll.Poll(ctx, healthcheckURL, pollingFrequency); err != nil {
		log.Printf("poll failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func usage() {
	indent := "       "
	claimUsage := fmt.Sprintf("%s claim --pid PID --pid-file PID_FILE", os.Args[0])
	killUsage := fmt.Sprintf("%s%s kill --pid-file PID_FILE [--show-stacks]", indent, os.Args[0])
	waitUsage := fmt.Sprintf("%s%s wait-until-healthy --url HEALTHCHECK_URL [--timeout DURATION (default 1m)] [--polling-frequency DURATION (default 5s)]", indent, os.Args[0])

	fmt.Fprintf(os.Stderr, "usage: %s\n%s\n%s\n", claimUsage, killUsage, waitUsage)
	fmt.Fprintf(os.Stderr, "\n%sDURATIONS must be specified with units (e.g. 10s, 4m, 500ms)\n", indent)
	os.Exit(1)
}
