package gateway

import (
	"fmt"
	"testing"
	"time"

	"github.com/benalexau/ibconnect/core"
)

func TestControllerNormalOperation(t *testing.T) {
	tcff := NewTestControllerFeedFactory(nil)
	runGatewayController(t, tcff, 1, 0)
}

func TestControllerErrorGenerator(t *testing.T) {
	counter := 0
	errorCount := 3
	tcff := NewTestControllerFeedFactory(func(ctf *TestControllerFeed) {
		if counter < errorCount {
			counter++
			ctf.ctx.Errors <- FeedError{
				fmt.Errorf("intentional error %d of %d", counter, errorCount),
				ctf,
			}
		}
	})
	runGatewayController(t, tcff, errorCount+1, errorCount)
}

func TestControllerGatewayLoss(t *testing.T) {
	counter := 0
	errorCount := 2
	tcff := NewTestControllerFeedFactory(func(ctf *TestControllerFeed) {
		if counter < errorCount {
			counter++
			ctf.ctx.Eng.Stop() // easiest way to kill it
		}
	})
	runGatewayController(t, tcff, errorCount+1, errorCount)
}

func TestControllerEngineUnavailable(t *testing.T) {
	c := core.NewTestConfig(t)
	c.IbGws = []string{"127.0.0.0:0000"}

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	tcff := NewTestControllerFeedFactory(func(ctf *TestControllerFeed) {
		t.Fatal("Should never have opened feed")
	})
	ffs := []FeedFactory{tcff}
	gatewayController, err := NewGatewayController(ffs, ctx.DB, ctx.N, ctx.DL, c.IbGws, c.IbClientId)
	if err != nil {
		t.Fatal(err)
	}
	defer gatewayController.Close()

	failAt := time.Now().Add(1 * time.Second)
	for gatewayController.Restarts() <= 5 {
		if time.Now().After(failAt) {
			t.Fatal("GatewayController failed to restart before timeout")
		}
	}
}

// runGatewayController executes the GatewayController, expecting the
// TestControllerFeed to be created and closed the specified number of times
// prior to the GatewayController being closed.
func runGatewayController(t *testing.T, tcff *TestControllerFeedFactory, expectedOpens int, expectedCloses int) {
	c := core.NewTestConfig(t)

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ffs := []FeedFactory{tcff}
	gatewayController, err := NewGatewayController(ffs, ctx.DB, ctx.N, ctx.DL, c.IbGws, c.IbClientId)
	if err != nil {
		t.Fatal(err)
	}
	defer gatewayController.Close()

	terminating := false
	for {
		select {
		case ct := <-tcff.openClose:
			if ct[0] < expectedOpens {
				continue
			}
			if ct[1] < expectedCloses {
				continue
			}
			if !terminating && ct[0] > expectedOpens {
				t.Fatalf("too many opens (saw %d, expected %d)",
					ct[0], expectedOpens)
			}
			if !terminating && ct[1] > expectedCloses {
				t.Fatalf("too many closes (saw %d, expected %d)",
					ct[1], expectedCloses)
			}
			if !terminating {
				terminating = true
				gatewayController.Close()
			}
			if ct[0] == ct[1] {
				// enough events seen, and everything closed
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout after %d opens (expected %d) and %d closes (expected %dd)",
				tcff.opened, expectedOpens, tcff.closed, expectedCloses)
			return
		}
	}
}

// TestControllerFeedFactory allows creation of a fake FeedFactory which will
// run the specified function on NewFeed, and offer a channel to report the
// number of open/close operations performed on the TestControllerFeed.
type TestControllerFeedFactory struct {
	openClose chan [2]int // channel with {opened,closed} on change
	opened    int
	closed    int
	fun       func(*TestControllerFeed) // goroutine to load on NewFeed
}

func NewTestControllerFeedFactory(fun func(*TestControllerFeed)) *TestControllerFeedFactory {
	tcff := &TestControllerFeedFactory{
		openClose: make(chan [2]int, 100),
		fun:       fun,
	}
	return tcff
}

func (c *TestControllerFeedFactory) pending() bool {
	return c.closed < c.opened
}

func (c *TestControllerFeedFactory) Done() core.NtType {
	return core.NtRefreshAll
}

func (c *TestControllerFeedFactory) NewFeed(ctx *FeedContext) *Feed {
	ctf := &TestControllerFeed{ctx: ctx}
	// pointers in TestControllerFeed allow tracking open/close counts
	ctf.opened = &c.opened
	ctf.closed = &c.closed
	ctf.openClose = &c.openClose
	c.opened++
	// simulate what a feed would do, ie go and run a function
	if c.fun != nil {
		go c.fun(ctf)
	}
	c.openClose <- [2]int{c.opened, c.closed}
	var f Feed = ctf
	return &f
}

type TestControllerFeed struct {
	ctx       *FeedContext
	opened    *int
	closed    *int
	openClose *chan [2]int
}

func (c *TestControllerFeed) Close() {
	if *c.opened == *c.closed {
		return
	}
	*c.closed++
	*c.openClose <- [2]int{*c.opened, *c.closed}
}
