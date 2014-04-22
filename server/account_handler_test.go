package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/benalexau/ibconnect/core"
	"github.com/benalexau/ibconnect/gateway"
)

func TestAccountHandlerGetAll(t *testing.T) {
	ctx, handler := NewTestHandler(t)
	defer ctx.Close()

	c := core.NewTestConfig(t)
	var ff gateway.FeedFactory = &gateway.AccountFeedFactory{AccountRefresh: c.AccountRefresh}
	WaitForFeed(t, ctx, &ff, 15*time.Second)

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/v1/accounts", nil))
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("Cache-Control", "private, max-age=60")
}

func TestAccountHandlerGetLatest(t *testing.T) {
	ctx, handler := NewTestHandler(t)
	defer ctx.Close()

	c := core.NewTestConfig(t)
	var ff gateway.FeedFactory = &gateway.AccountFeedFactory{AccountRefresh: c.AccountRefresh}
	WaitForFeed(t, ctx, &ff, 15*time.Second)

	accountCode := ""
	row := ctx.DB.QueryRow("SELECT account_code FROM account LIMIT 1")
	if err := row.Scan(&accountCode); err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("http://1.2.3.4/v1/accounts/%s", accountCode)
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", url, nil))
	recorded.CodeIs(http.StatusSeeOther)
	target := recorded.Recorder.Header().Get("Location")

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", target, nil))
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("Cache-Control", "private, max-age=31556926")
}

func TestAccountHandlerGetAllRefresh(t *testing.T) {
	ctx, handler := NewTestHandler(t)
	defer ctx.Close()

	c := core.NewTestConfig(t)
	var ff gateway.FeedFactory = &gateway.AccountFeedFactory{AccountRefresh: c.AccountRefresh}
	WaitForFeed(t, ctx, &ff, 15*time.Second)

	nc := make(chan *core.Notification)
	ctx.N.Subscribe(nc)
	defer ctx.N.Unsubscribe(nc)

	go func() {
		req := test.MakeSimpleRequest("GET", "http://1.2.3.4/v1/accounts", nil)
		req.Header.Add("Cache-Control", "private; max-age=0")
		test.RunRequest(t, handler, req)
		// NB: This request will not return data, as there's no feed running.
		// Long timeouts avoided as the ctx.Close() closes the Notifier,
		// which in turn closes the RefreshIfNeeded listener.
	}()

outer:
	for {
		select {
		case msg := <-nc:
			if msg.Type == core.NtAccountRefresh {
				// controller requested an account refresh, like it should have
				break outer
			}
		case <-time.After(15 * time.Second):
			t.Fatal("controller didn't request update before timeout")
		}
	}
}
