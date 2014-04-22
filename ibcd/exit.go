package main

import (
	"os"
	"os/signal"
	"syscall"
)

// handleSignals registers signal handlers, returning a channel that will be
// closed on SIGTERM, SIGHUB or SIGINT.
func handleSignals() <-chan struct{} {
	terminated := make(chan struct{})
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	go func() {
		defer func() {
			close(terminated)
		}()

		select {
		case <-signalCh:
			return
		}
	}()

	return terminated
}
