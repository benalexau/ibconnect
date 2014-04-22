package gateway

import (
	"testing"
	"time"

	"github.com/benalexau/ibconnect/core"
)

func TestAccountFeedInsertsDataOnStartup(t *testing.T) {
	c := core.NewTestConfig(t)
	var ff FeedFactory = &AccountFeedFactory{c.AccountRefresh}
	TestSimpleFeedInsertsDataOnStartup(t, &ff, "account_snapshot", 15*time.Second)
}

func TestAccountFeedHandlesEngineTermination(t *testing.T) {
	c := core.NewTestConfig(t)
	var ff FeedFactory = &AccountFeedFactory{c.AccountRefresh}
	TestSimpleFeedHandlesEngineTermination(t, &ff, 15*time.Second)
}

func TestAccountFeedHandlesNoEngine(t *testing.T) {
	c := core.NewTestConfig(t)
	var ff FeedFactory = &AccountFeedFactory{c.AccountRefresh}
	TestSimpleFeedHandlesNoEngine(t, &ff)
}

func TestAccountFeedPublishesDoneMessage(t *testing.T) {
	c := core.NewTestConfig(t)
	var ff FeedFactory = &AccountFeedFactory{c.AccountRefresh}
	TestSimpleFeedPublishesDoneMessage(t, &ff, 15*time.Second)
}
