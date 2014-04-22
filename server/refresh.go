package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/benalexau/ibconnect/core"
)

// RefreshIfNeeded detects a HTTP request for an immediate refresh of the gateway backend.
// If detected, it publishes a "request" notification and blocks awaiting the "completed"
// acknowledgement reply. An error is returned if acknowledgement exceeds the timeout.
func RefreshIfNeeded(n *core.Notifier, r *rest.Request, requestRefresh core.NtType, completedRefresh core.NtType, timeout time.Duration) error {
	if strings.Contains(r.Header.Get("Cache-Control"), "max-age=0") {
		notifications := make(chan *core.Notification)
		n.Subscribe(notifications)
		defer n.Unsubscribe(notifications)

		n.Publish(requestRefresh, 0)

		for {
			select {
			case msg := <-notifications:
				if msg == nil {
					return errors.New("Subscription channel unexpectedly closed; did another goroutine close the notifier?")
				}
				if msg.Type == completedRefresh {
					break
				}
			case <-time.After(timeout):
				return fmt.Errorf("Timeout %v waiting for '%v' response to '%v' request", timeout, completedRefresh, requestRefresh)
			}
		}
	}
	return nil
}
