package gateway

import (
	"database/sql"
	"log"
	"time"

	"github.com/benalexau/ibconnect/core"
)

const lockManagerKey int64 = 9063409683409876463

// GatewayController ensures this node will execute a GatewayService for each
// IB API endpoint (if no other node in the cluster is doing so). It will
// automatically restart any failed GatewayService.
type GatewayController struct {
	exit       chan bool
	terminated chan struct{}
	db         *sql.DB
	n          *core.Notifier
	distLock   *core.DistLock
	ibGws      []string
	ibClientId int
	ffs        []FeedFactory
	restarts   int
}

func NewGatewayController(ffs []FeedFactory, db *sql.DB, n *core.Notifier, distLock *core.DistLock, ibGws []string, ibClientId int) (*GatewayController, error) {
	g := &GatewayController{
		exit:       make(chan bool),
		terminated: make(chan struct{}),
		db:         db,
		n:          n,
		distLock:   distLock,
		ibGws:      ibGws,
		ibClientId: ibClientId,
		ffs:        ffs,
	}
	g.initGatewayController()
	return g, nil // never returns error, but declared for consistency
}

// Close terminates the GatewayController, including any GatewayService
// instances it may be controlling. Close can be called multiple times safely,
// and it will block until the GatewayController has been closed.
func (g *GatewayController) Close() {
	select {
	case <-g.terminated:
		return
	case g.exit <- true:
	}
	<-g.terminated
}

// Restarts reports how many times the GatewayController has restarted a GatewayService.
// Zero may indicate an absence of errors, or that the controller is not the leader.
func (g *GatewayController) Restarts() int {
	return g.restarts
}

type GatewayError struct {
	Error error
	IbGw  string
}

func (g *GatewayController) initGatewayController() {
	go func() {
		lock := lockManagerKey
		abandonLock := make(chan struct{})
		errorReports := make(chan GatewayError)
		lockReply := g.distLock.Request(lock, abandonLock)
		gateways := make(map[string]*GatewayService)
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
					case <-errorReports:
					}
				}()
				for _, gwservice := range gateways {
					gwservice.Close()
				}
				close(errsink)
				close(abandonLock)
				close(g.terminated)
			case acquiredLock, ok := <-lockReply:
				if ok && acquiredLock {
					for _, ibGw := range g.ibGws {
						gateways[ibGw] = NewGatewayService(errorReports, g.ffs, g.db, g.n, ibGw, g.ibClientId)
					}
				}
			case gwerr := <-errorReports:
				// close erroneous gateway and open a fresh replacement
				time.Sleep(100 * time.Millisecond)
				g.restarts++
				log.Printf("%s %s", gwerr.IbGw, gwerr.Error.Error())
				gateways[gwerr.IbGw].Close()
				log.Printf("%s restarting", gwerr.IbGw)
				gateways[gwerr.IbGw] = NewGatewayService(errorReports, g.ffs, g.db, g.n, gwerr.IbGw, g.ibClientId)
			}
		}
	}()
}
