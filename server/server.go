package server

import (
	"net"
	"net/http"
	"time"

	"github.com/etix/stoppableListener"
)

// Serve executes a HTTP server and blocks until the passed channel closes.
func Serve(terminated <-chan struct{}, address string, handler http.Handler) error {
	listener, err := net.Listen("tcp", address)
	stoppable := stoppableListener.Handle(listener)

	go func() {
		select {
		case <-terminated:
			stoppable.Stop <- true
		}
	}()

	err = http.Serve(stoppable, handler)
	if !stoppable.Stopped && err != nil {
		return err
	}

	// Give up to 5 seconds for in-flight requests to complete
	killAt := time.Now().Add(5 * time.Second)
	for time.Now().Before(killAt) && stoppable.ConnCount.Get() > 0 {
	}

	return nil
}
