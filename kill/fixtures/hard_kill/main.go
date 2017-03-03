package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	fmt.Printf("Running as %d\n", os.Getpid())

	for {
		select {
		case sig := <-c:
			handleSignal(sig)
		case now := <-tick.C:
			fmt.Printf("%v+\n", now)
		}
	}
}

func handleSignal(sig os.Signal) {
	switch sig {
	case syscall.SIGTERM:
		fmt.Println("GOT A", sig)
	default:
		fmt.Println("Default?")
	}
}
