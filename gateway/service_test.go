package gateway

import (
	"testing"

	"github.com/benalexau/ibconnect/core"
)

// This file only tests areas not already covered by controller_test.go

func TestServiceCloseCanExecuteTwiceWithoutIssues(t *testing.T) {
	c := core.NewTestConfig(t)

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	terminate := make(chan struct{})
	errs := make(chan GatewayError)
	go func() {
		for {
			select {
			case <-errs:
			case <-terminate:
				return
			}
		}
	}()

	ffs := FeedFactories(c)
	ibGw := "127.0.0.0:0123"
	service := NewGatewayService(errs, ffs, ctx.DB, ctx.N, ibGw, c.IbClientId)
	service.Close()
	service.Close()
	close(terminate)
}
