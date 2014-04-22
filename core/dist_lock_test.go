package core

import (
	"testing"
	"time"
)

func TestNormalLockCycle(t *testing.T) {
	distLock := getLockManager(t)
	defer distLock.Close()

	lock := int64(2349875)
	abandon := make(chan struct{})

	reply := distLock.Request(lock, abandon)

	expectLock(t, reply)
	close(abandon)
	expectRelease(t, reply)
}

func TestCompetingLockNotGranted(t *testing.T) {
	lock := int64(2349875)

	// need 2 lock managers as one connection can acquire the same lock twice
	distLock1 := getLockManager(t)
	defer distLock1.Close()

	distLock2 := getLockManager(t)
	defer distLock2.Close()

	abandon1 := make(chan struct{})
	reply1 := distLock1.Request(lock, abandon1)
	expectLock(t, reply1)

	abandon2 := make(chan struct{})
	reply2 := distLock2.Request(lock, abandon2)
	expectNoLock(t, reply2, 100*time.Millisecond)

	close(abandon1)
	expectRelease(t, reply1)

	expectLock(t, reply2)
	close(abandon2)
	expectRelease(t, reply2)
}

func TestLockManagerClosureCancelsLocks(t *testing.T) {
	distLock := getLockManager(t)
	defer distLock.Close()

	lock := int64(2349875)
	abandon := make(chan struct{})

	reply := distLock.Request(lock, abandon)

	expectLock(t, reply)
	distLock.Close()

	expectRelease(t, reply)
}

func TestClosedLockManagerGivesClosedReplyForNewRequests(t *testing.T) {
	distLock := getLockManager(t)
	distLock.Close()

	lock := int64(2349875)
	abandon := make(chan struct{})

	reply := distLock.Request(lock, abandon)
	expectClosedReplyChannel(t, reply)
}

func getLockManager(t *testing.T) *DistLock {
	config := NewTestConfig(t)

	distLock, err := NewDistLock(config.DbUrl)
	if err != nil {
		t.Fatal(err)
	}

	return distLock
}

func expectLock(t *testing.T, reply <-chan bool) {
	select {
	case acquired, ok := <-reply:
		if !ok {
			t.Fatal("reply channel unexpectedly closed")
		}
		if !acquired {
			t.Fatal("server incorrectly sent false on reply channel")
		}
	case <-time.After(3000 * time.Millisecond):
		t.Fatal("lock manager did not sent reply to lock request")
	}
}

func expectNoLock(t *testing.T, reply <-chan bool, waitTime time.Duration) {
	select {
	case acquired, ok := <-reply:
		if !ok {
			t.Fatal("reply channel unexpectedly closed")
		}
		if acquired {
			t.Fatal("server incorrectly acquired the lock")
		}
	case <-time.After(waitTime):
		// good, this is what we want
	}
}

func expectClosedReplyChannel(t *testing.T, reply <-chan bool) {
	select {
	case _, ok := <-reply:
		if ok {
			t.Fatal("reply channel should have been closed")
		}
	}
}

func expectRelease(t *testing.T, reply <-chan bool) {
	select {
	case _, ok := <-reply:
		if ok {
			t.Fatal("reply channel should not have sent data")
		}
	case <-time.After(3000 * time.Millisecond):
		t.Fatal("lock manager did not confirm abandonment of lock")
	}
}
