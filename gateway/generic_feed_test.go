package gateway

import (
	"github.com/benalexau/ibconnect/core"
	"github.com/gofinance/ib"
	"github.com/gorhill/cronexpr"
	"testing"
	"time"
)

func TestGenericHourlyFeed(t *testing.T) {
	if err := runGenericFeedTest(t, nil, cronexpr.MustParse("@hourly"), 1*time.Second, 1); err != nil {
		t.Fatal(err)
	}
}

func TestGenericFeedRefreshInterval(t *testing.T) {
	oncePerSecond := "* * * * * * *"
	if err := runGenericFeedTest(t, nil, cronexpr.MustParse(oncePerSecond), 3*time.Second, 2); err != nil {
		t.Fatal(err)
	}
}

func TestGenericFeedFiredOnRefreshAllEvent(t *testing.T) {
	done := false
	fun := func(ctx *FeedContext) {
		if !done {
			done = true
			ctx.N.Publish(core.NtRefreshAll, 0)
		}
	}
	// expect 2 events (1 due to startup, 1 due to NtRefreshAll request)
	if err := runGenericFeedTest(t, fun, cronexpr.MustParse("@hourly"), 1*time.Second, 2); err != nil {
		t.Fatal(err)
	}
}

func TestGenericFeedKilledNotificationSystem(t *testing.T) {
	fun := func(ctx *FeedContext) {
		ctx.N.Close() // easiest way to kill it
	}
	if err := runGenericFeedTest(t, fun, cronexpr.MustParse("@hourly"), 1*time.Second, 100000); err == nil {
		t.Fatal("killing notifier should have reported back an error")
	}
}

// runGenericFeedTest returns any error reported to the error channel. It fails
// the test if the expected count is not reached within one second of loading.
func runGenericFeedTest(t *testing.T, fun func(*FeedContext), cronRefresh *cronexpr.Expression, waitTime time.Duration, expectedCount int) error {
	c := core.NewTestConfig(t)

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	errors := make(chan FeedError)
	var lastError error
	terminateErrors := make(chan struct{})
	defer close(terminateErrors)
	go func() {
		for {
			select {
			case e := <-errors:
				lastError = e.Error
			case <-terminateErrors:
				return
			}
		}
	}()

	var engine *ib.Engine
	fc := &FeedContext{errors, ctx.DB, ctx.N, engine}
	gft := newTestGenericFeed(t, fc, fun, cronRefresh)
	defer gft.Close()

	killAt := time.Now().Add(waitTime)
	for gft.counter < expectedCount {
		if lastError != nil {
			return lastError
		}
		if time.Now().After(killAt) {
			t.Fatal("Insufficient callbacks (%d) before timeout", gft.counter)
		}
	}

	gft.Close()
	return lastError
}

type TestGenericFeed struct {
	fun     func(*FeedContext)
	generic *GenericFeed
	counter int
}

func newTestGenericFeed(t *testing.T, ctx *FeedContext, fun func(*FeedContext), cronRefresh *cronexpr.Expression) *TestGenericFeed {
	g := &TestGenericFeed{}
	notifications := []core.NtType{core.NtRefreshAll}
	g.fun = fun
	g.generic = NewGenericFeed(ctx, cronRefresh, notifications, g.callback)
	return g
}

func (g *TestGenericFeed) callback(ctx *FeedContext) {
	g.counter++
	if g.fun != nil {
		g.fun(ctx)
	}
}

func (g *TestGenericFeed) Close() {
	g.generic.Close()
}
