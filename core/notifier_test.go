package core

import (
	"testing"
	"time"
)

func TestNotifications(t *testing.T) {
	config := NewTestConfig(t)

	notifier, err := NewNotifier(config.DbUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer notifier.Close()

	var ntPlay NtType = "plaything"
	err = notifier.RegisterAll([]NtType{ntPlay})
	if err != nil {
		t.Fatal(err)
	}

	nc1 := make(chan *Notification)
	nc2 := make(chan *Notification)
	notifier.Subscribe(nc1)
	notifier.Subscribe(nc2)
	defer notifier.Unsubscribe(nc1)
	defer notifier.Unsubscribe(nc2)

	payload := int64(4340986482)
	go func() {
		notifier.Publish(ntPlay, payload)
	}()

	received := 0
	for {
		select {
		case msg := <-nc1:
			received++
			if msg.Id != payload || msg.Type != ntPlay {
				t.Fatalf("Incorrect payload")
			}
			if received == 2 {
				return
			}
		case msg := <-nc2:
			received++
			if msg.Id != payload || msg.Type != ntPlay {
				t.Fatalf("Incorrect payload")
			}
			if received == 2 {
				return
			}
		case <-time.After(3000 * time.Millisecond):
			t.Fatalf("Did not receive 2 notifications (received %d)", received)
		}
	}
}
