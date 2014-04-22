package gateway

import (
	"database/sql"
	"fmt"
	"github.com/benalexau/ibconnect/core"
	"github.com/gofinance/ib"
	"sync"
	"testing"
	"time"
)

// TestFeedContext supports writing tests that require a FeedContext.
type TestFeedContext struct {
	closed sync.Once
	ctx    *core.Context
	FC     *FeedContext
}

// NewTestFeedContext provides a simple way of producing a FeedContext for tests.
func NewTestFeedContext(t *testing.T) *TestFeedContext {
	c := core.NewTestConfig(t)

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}

	engine, err := ib.NewEngine(ib.NewEngineOptions{Gateway: c.IbGws[0]})
	if err != nil {
		t.Fatal(err)
	}

	fc := &FeedContext{
		Errors: make(chan FeedError),
		DB:     ctx.DB,
		N:      ctx.N,
		Eng:    engine,
	}

	return &TestFeedContext{
		ctx: ctx,
		FC:  fc,
	}
}

func (f *TestFeedContext) Close() {
	f.closed.Do(func() {
		if f.FC.Eng != nil {
			defer f.FC.Eng.Stop()
		}
		defer f.ctx.Close()
	})
}

func Count(t *testing.T, db *sql.DB, table string) int {
	count := 0
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	row := db.QueryRow(query)
	if err := row.Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

// TestSimpleFeedInsertsDataOnStartup ensures the Feed will grow the table before timeout.
// The table must grow by at least one row. Any error will fail the test.
func TestSimpleFeedInsertsDataOnStartup(t *testing.T, ff *FeedFactory, table string, timeout time.Duration) {
	tfc := NewTestFeedContext(t)
	defer tfc.Close()
	before := Count(t, tfc.FC.DB, table)

	feed := (*ff).NewFeed(tfc.FC)
	defer (*feed).Close()

	kill := time.Now().Add(timeout)
	for {
		select {
		case err := <-tfc.FC.Errors:
			t.Fatal(err)
			return
		default:
			count := Count(t, tfc.FC.DB, table)
			if count >= before+1 {
				return
			}
			if time.Now().After(kill) {
				t.Fatalf("Table %s did not update in time (was %d; now %d)", table, before, count)
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// TestSimpleFeedHandlesEngineTermination ensures the Feed will detect Engine failures.
// The Feed must report an error before the timeout.
func TestSimpleFeedHandlesEngineTermination(t *testing.T, ff *FeedFactory, timeout time.Duration) {
	tfc := NewTestFeedContext(t)
	defer tfc.Close()

	tfc.FC.Eng.Stop()

	feed := (*ff).NewFeed(tfc.FC)
	defer (*feed).Close()
	for {
		select {
		case <-tfc.FC.Errors:
			return
		case <-time.After(timeout):
			t.Fatal("Timeout reached and engine failure never reported")
		}
	}
}

// TestSimpleFeedHandlesNoEngine ensures the Feed will detect a missing Engine.
// The Feed must report an error within 1 second of being started.
func TestSimpleFeedHandlesNoEngine(t *testing.T, ff *FeedFactory) {
	tfc := NewTestFeedContext(t)
	defer tfc.Close()

	tfc.FC.Eng = nil

	feed := (*ff).NewFeed(tfc.FC)
	defer (*feed).Close()
	for {
		select {
		case <-tfc.FC.Errors:
			return
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout reached and engine nil never reported")
		}
	}
}

// TestSimpleFeedPublishesDoneMessage ensures the Feed will report it is done.
func TestSimpleFeedPublishesDoneMessage(t *testing.T, ff *FeedFactory, timeout time.Duration) {
	tfc := NewTestFeedContext(t)
	defer tfc.Close()

	feed := (*ff).NewFeed(tfc.FC)
	defer (*feed).Close()

	notifications := make(chan *core.Notification)
	tfc.FC.N.Subscribe(notifications)
	defer tfc.FC.N.Unsubscribe(notifications)

	for {
		select {
		case err := <-tfc.FC.Errors:
			t.Fatal(err)
			return
		case event := <-notifications:
			if event.Type == (*ff).Done() {
				return
			}
		case <-time.After(timeout):
			t.Fatal("Timeout reached and feed never reported as done")
		}
	}
}
