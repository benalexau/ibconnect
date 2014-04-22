package core

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

// distlock provides a simple distributed lock manager that is backed by
// Postgres session-level advisory locks. Such locks are held until explicitly
// released or the Postgres connection ends. Corner cases like the loss of
// connectivity between this node and the database resulting in loss of the
// lock (as the database considers the session has ended) are currently
// unhandled, although the API contract enables this to be easily added in the
// future without any existing client changes.
type DistLock struct {
	exit       chan bool
	terminated chan struct{}
	db         *sql.DB
}

// NewDistLock returns a distributed lock manager.
func NewDistLock(dbUrl string) (*DistLock, error) {
	db, err := sql.Open("postgres", dbUrl)

	if err != nil {
		return nil, err
	}

	n := &DistLock{
		exit:       make(chan bool),
		terminated: make(chan struct{}),
		db:         db,
	}
	if err := n.initLockManager(); err != nil {
		return nil, err
	}
	return n, nil
}

func (d *DistLock) initLockManager() error {
	go func() {
		for {
			select {
			case <-d.terminated:
				return
			case <-d.exit:
				d.db.Exec("SELECT pg_advisory_unlock_all()")
				d.db.Close()
				close(d.terminated)
			}
		}
	}()

	return nil
}

// Request attempts to acquire a lock for the passed id. The caller can abandon
// the lock request (or acquired lock) by closing the abandon channel. The
// server will send true on the reply channel if the lock is acquired. It will
// close the reply channel if the lock is abandoned (usually due to the client
// requesting abandonment by closing the abandon channel, or due to the lock
// manager closing or already being closed). Future connection monitoring
// features will also tie into the reply channel being closed on unexpected loss
// of the lock.
func (d *DistLock) Request(id int64, abandon <-chan struct{}) <-chan bool {
	reply := make(chan bool)
	acquired := false

	doNonBlockRequest := func() {
		str := fmt.Sprintf("SELECT pg_try_advisory_lock(%d)", id)
		row := d.db.QueryRow(str)
		row.Scan(&acquired)
		if acquired {
			acquired = true
			reply <- true
		}

	}

	doLockAbandon := func() {
		if acquired {
			str := fmt.Sprintf("SELECT pg_advisory_unlock(%d)", id)
			d.db.Exec(str)
			// we ignore result here, as db conn might have closed
			// (in which case our lock is automatically gone anyway)
		}
	}

	go func() {
		defer close(reply)
		select {
		case <-d.terminated:
			return
		case <-abandon:
			doLockAbandon()
			return
		default:
			doNonBlockRequest()
		}

		for {
			select {
			case <-d.terminated:
				return
			case <-abandon:
				doLockAbandon()
				return
			case <-time.After(100 * time.Millisecond):
				if acquired {
					// TODO: monitor connection failure and doAbandonLock
				} else {
					doNonBlockRequest()
				}
			}

		}
	}()
	return reply
}

// Close terminates the lock manager and all locks. It will cause all reply
// channels to close. Close can be called multiple times safely, and it will
// block until the lock manager has been closed.
func (d *DistLock) Close() {
	select {
	case <-d.terminated:
		return
	case d.exit <- true:
	}
	<-d.terminated
}
