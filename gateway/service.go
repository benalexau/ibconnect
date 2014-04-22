package gateway

import (
	"database/sql"
	"fmt"

	"github.com/benalexau/ibconnect/core"
	"github.com/gofinance/ib"
)

// GatewayService represents an attempt at communication with a single IP API
// gateway, loading a series of Feed workers for individual data exchange use
// cases. Any failures are reported back via the errors channel passed when the
// value was created.
type GatewayService struct {
	exit       chan bool
	terminated chan struct{}
	errors     chan<- GatewayError
	ffs        []FeedFactory
	ibGw       string
	ibClientId int
	ctx        *FeedContext
}

// NewGatewayService loads a GatewayService. It guarantees any errors are reported
// to the passed error channel.
func NewGatewayService(errors chan<- GatewayError, ffs []FeedFactory, db *sql.DB, n *core.Notifier, ibGw string, ibClientId int) *GatewayService {
	ctx := &FeedContext{
		Errors: make(chan FeedError),
		DB:     db,
		N:      n,
	}
	g := &GatewayService{
		exit:       make(chan bool),
		terminated: make(chan struct{}),
		errors:     errors,
		ffs:        ffs,
		ibGw:       ibGw,
		ibClientId: ibClientId,
		ctx:        ctx,
	}
	g.initGatewayService()
	return g
}

// Close terminates the GatewayService, including any Feed instances it may be
// running. Close can be called multiple times safely, and it will block until
// the GatewayService has been closed.
func (g *GatewayService) Close() {
	select {
	case <-g.terminated:
		return
	case g.exit <- true:
	}
	<-g.terminated
}

func (g *GatewayService) initGatewayService() {
	go func() {
		feeds := []*Feed{}
		esl := make(chan ib.EngineState)
		var err error

		g.ctx.Eng, err = ib.NewEngine(ib.NewEngineOptions{Gateway: g.ibGw, Client: int64(g.ibClientId)})
		if err == nil {
			defer g.ctx.Eng.Stop()

			g.ctx.Eng.SubscribeState(esl)
			defer g.ctx.Eng.UnsubscribeState(esl)

			for _, ff := range g.ffs {
				feeds = append(feeds, ff.NewFeed(g.ctx))
			}
		} else {
			g.errors <- GatewayError{err, g.ibGw}
		}

		for {
			select {
			case <-g.terminated:
				return
			case <-g.exit:
				errsink := make(chan struct{})
				go func() {
					select {
					case <-errsink:
						return
					case <-g.ctx.Errors:
					}
				}()
				for _, feed := range feeds {
					(*feed).Close()
				}
				close(errsink)
				close(g.terminated)
			case feederr := <-g.ctx.Errors:
				g.errors <- GatewayError{feederr.Error, g.ibGw}
			case es := <-esl:
				if es != ib.EngineReady {
					// Engine should never report this state (in normal shutdown we've unsubscribed, so we would never receive this state change)
					err = g.ctx.Eng.FatalError()
					if err == nil {
						err = fmt.Errorf("%s without reporting fatal error", es.String())
					}
					g.errors <- GatewayError{err, g.ibGw}
				}
			}
		}
	}()
}
