package server

import (
	"github.com/benalexau/ibconnect/core"
	"testing"
	"time"
)

func TestServeExitsOnChannelClose(t *testing.T) {
	ctx, handler := NewTestHandler(t)
	defer ctx.Close()

	c, err := core.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	terminated := make(chan struct{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(terminated)
	}()

	err = Serve(terminated, c.Address(), handler)
	if err != nil {
		t.Fatal(err)
	}
}
