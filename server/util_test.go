package server

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/benalexau/ibconnect/core"
	"github.com/benalexau/ibconnect/gateway"
	"net/http"
	"testing"
	"time"
)

func TestUuid(t *testing.T) {
	u := Util{}
	uuid := u.uuid()
	if len(uuid) != 36 {
		t.Fatalf("unexpected length of UUID %s", uuid)
	}
}

func TestErrorHandlerInternalServerError(t *testing.T) {
	c := core.NewTestConfig(t)

	ctx, err := core.NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	handler := Handler(true, ctx.DB, ctx.N)

	var ff gateway.FeedFactory = &gateway.AccountFeedFactory{AccountRefresh: c.AccountRefresh}
	WaitForFeed(t, ctx, &ff, 5*time.Second)

	ctx.DB.Close() // this will cause an internal server error

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/v1/accounts", nil))
	recorded.CodeIs(http.StatusInternalServerError)
	recorded.ContentTypeIsJson()
}

func TestErrorHandlerDataNotFound(t *testing.T) {
	ctx, handler := NewTestHandler(t)
	defer ctx.Close()

	accountCode := "neverfind"
	url := fmt.Sprintf("http://1.2.3.4/v1/accounts/%s", accountCode)
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", url, nil))
	recorded.CodeIs(http.StatusNotFound)
	recorded.ContentTypeIsJson()
}
