package server

import (
	"github.com/benalexau/ibconnect/core"
	"github.com/benalexau/ibconnect/gateway"
	"net/http"
	"testing"
	"time"
)

// NewTestHandler returns the same handler chain as the server will usually use.
// Be sure to defer Context.Close().
func NewTestHandler(t *testing.T) (*core.Context, http.Handler) {
	c, err := core.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}

	return ctx, Handler(c.ErrInfo, ctx.DB, ctx.N)
}

// WaitForFeed blocks the goroutine until the FeedFactory has sent a Done event.
// This is useful for ensuring the database has some content.
func WaitForFeed(t *testing.T, ctx *core.Context, ff *gateway.FeedFactory, timeout time.Duration) {
	// Prepare feed context
	fct := gateway.NewTestFeedContext(t)
	defer fct.Close()

	// Prepare channel to monitor when feed has completed its work
	notifications := make(chan *core.Notification)
	fct.FC.N.Subscribe(notifications)
	defer fct.FC.N.Unsubscribe(notifications)

	// Run feed
	feed := (*ff).NewFeed(fct.FC)
	defer (*feed).Close()

	// Wait
	for {
		select {
		case err := <-fct.FC.Errors:
			t.Fatalf("error while waiting: %v", err)
		case event := <-notifications:
			if event.Type == (*ff).Done() {
				return
			}
		case <-time.After(timeout):
			t.Fatalf("timeout waiting for %v", (*ff).Done())
		}
	}
}
