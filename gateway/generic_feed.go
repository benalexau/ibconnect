package gateway

import (
	"errors"
	"time"

	"github.com/benalexau/ibconnect/core"
	"github.com/gorhill/cronexpr"
)

// GenericFeed simplifies creation of Feed types that comply with the contract.
// Feeds simply use NewGenericFeed, passing it configuration information and a
// callback. The callback is fired at relevant times. The callback MUST send an
// error on the FeedContext errors channel or cleanly complete its work. It is
// strongly recommended that the callback use a single database transaction.
// The callback should not close any channels or objects in the FeedContext.
type GenericFeed struct {
	exit          chan bool
	terminated    chan struct{}
	refreshChan   chan bool
	ctx           *FeedContext
	cronRefresh   *cronexpr.Expression
	notifications []core.NtType
	callback      func(*FeedContext)
}

// NewGenericFeed returns an GenericFeed that will immediately start using the callback.
func NewGenericFeed(ctx *FeedContext, cronRefresh *cronexpr.Expression, notifications []core.NtType, callback func(*FeedContext)) *GenericFeed {
	a := GenericFeed{
		exit:          make(chan bool),
		terminated:    make(chan struct{}),
		refreshChan:   make(chan bool),
		ctx:           ctx,
		cronRefresh:   cronRefresh,
		notifications: notifications,
		callback:      callback,
	}
	a.init()
	return &a
}

// Close terminates the GenericFeed. Close can be called multiple times safely,
// and it will block until the GenericFeed has been closed.
func (a *GenericFeed) Close() {
	select {
	case <-a.terminated:
		return
	case a.exit <- true:
	}
	<-a.terminated
}

// Init starts the goroutines that performs the work of this Feed instance.
func (a *GenericFeed) init() {
	refreshTermination := make(chan struct{})
	go func() {
		a.refreshChan <- true
		now := time.Now().UTC()
		nextTime := a.cronRefresh.Next(now)
		durationUntil := nextTime.Sub(now)
		for {
			select {
			case <-refreshTermination:
				return
			case <-time.After(durationUntil):
				a.refreshChan <- true
			}
		}
	}()

	go func() {
		notifyChan := make(chan *core.Notification)
		a.ctx.N.Subscribe(notifyChan)
		defer a.ctx.N.Unsubscribe(notifyChan)
		for {
			select {
			case <-a.terminated:
				return
			case <-a.exit:
				close(refreshTermination)
				close(a.terminated)
			case <-a.refreshChan:
				a.callback(a.ctx)
			case notification, ok := <-notifyChan:
				if !ok {
					err := errors.New("notification system stopped")
					a.ctx.Errors <- FeedError{err, a}
					continue
				}
				for _, interested := range a.notifications {
					if interested == notification.Type {
						a.callback(a.ctx)
					}
				}
			}
		}
	}()
}
