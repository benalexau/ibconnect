package gateway

import (
	"database/sql"
	"github.com/benalexau/ibconnect/core"
	"github.com/gofinance/ib"
)

func FeedFactories(c core.Config) []FeedFactory {
	f := []FeedFactory{}
	f = append(f, &AccountFeedFactory{c.AccountRefresh})
	return f
}

// Feed handles individual data exchange between IB API and the database.
// It must notify of any error by reporting to to a FeedError channel passed
// to the FeedFactory.NewFeed(..) function.
type Feed interface {
	Close()
}

// FeedFactory returns a Feed that will use the passed values. A FeedFactory
// must not send an error to the FeedError channel from the same goroutine as
// invoked FeedFactory.NewFeed(..), as this may block delivery of the error.
type FeedFactory interface {
	NewFeed(ctx *FeedContext) *Feed
	Done() core.NtType
}

// FeedError holds error information passed on the FeedContext errors channel.
type FeedError struct {
	Error error
	Feed  Feed
}

// FeedContext provides access to values commonly needed when writing Feeds.
type FeedContext struct {
	Errors chan FeedError
	DB     *sql.DB
	N      *core.Notifier
	Eng    *ib.Engine
}
