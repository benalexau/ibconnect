package core

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// Notifier provides an inter-process notification mechanism that is backed
// by Postgres notifications. A client can send a notification using publish()
// and it will be asynchronously delivered to all subscribers on all nodes. A
// notification may include a uint64 payload, which is commonly a primary key
// identifier. On termination of the notifier, all subscription channels will be
// closed.
type Notifier struct {
	types       map[string]NtType
	exit        chan bool
	terminated  chan struct{}
	ch          chan command
	subscribers []chan<- *Notification
	db          *sql.DB
	l           *pq.Listener
}

// NewNotifier creates a new Notifier and correctly initializes it.
func NewNotifier(dbUrl string) (*Notifier, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}

	n := &Notifier{
		types:      make(map[string]NtType),
		exit:       make(chan bool),
		terminated: make(chan struct{}),
		ch:         make(chan command),
		db:         db,
	}
	if err := n.initSubscribers(dbUrl); err != nil {
		return nil, err
	}
	return n, nil
}

// Publish transmits a notification to all subscribers.
func (n *Notifier) Publish(ntType NtType, id int64) {
	n.sendCommand(func() {
		str := fmt.Sprintf("NOTIFY %v, '%d'", ntType, id)
		_, err := n.db.Exec(str)
		if err != nil {
			log.Printf("Notification %v (%d) publish error: %v", ntType, id, err)
		}
	})
}

// RegisterAll is a convenience function to register all present ntTypes. The
// method will immediately return on any error being detected.
func (n *Notifier) RegisterAll(t []NtType) error {
	for _, nt := range t {
		err := n.Register(nt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Register registers a Postgres channel name that should be listened to.
func (n *Notifier) Register(t NtType) error {
	str := fmt.Sprintf("%v", t)
	n.types[str] = t
	if err := n.l.Listen(str); err != nil {
		return err
	}
	return nil
}

// Subscribe blocks until the passed channel is registered to receive
// notifications or the notifier has terminated.
func (n *Notifier) Subscribe(c chan<- *Notification) {
	n.sendCommand(func() {
		n.subscribers = append(n.subscribers, c)
	})
}

// Unsubscribe blocks until the passed channel will no longer receive
// notifications or the notifier has terminated. It also maintains a goroutine
// to sink the channel until the unsubscribe is finalised, which frees the
// caller from handling this.
func (n *Notifier) Unsubscribe(c chan *Notification) {
	terminated := make(chan struct{})
	go func() {
		for {
			select {
			case <-c:
			case <-terminated:
				return
			}
		}
	}()
	n.sendCommand(func() {
		newSubscribers := make([]chan<- *Notification, 0)
		for _, existing := range n.subscribers {
			if existing != c {
				newSubscribers = append(newSubscribers, existing)
			}
		}
		n.subscribers = newSubscribers
	})
	close(terminated)
}

// Close must be called when the Notifier is no longer required. It blocks until
// the Notifier has closed, and is safe to call multiple times.
func (n *Notifier) Close() {
	select {
	case <-n.terminated:
		return
	case n.exit <- true:
	}
	<-n.terminated
}

// NtType represents a notification that can be sent by a Notifier.
type NtType string

const (
	NtErrorFlag NtType = "__error__"
)

// Notification identifies the type of notification and optional identifier.
type Notification struct {
	Type NtType
	Id   int64
}

// initSubscribers performs one-time initialization of the Postgres listener and
// goroutine for event delivery, termination and subscription management.
func (n *Notifier) initSubscribers(dbUrl string) error {
	n.l = pq.NewListener(dbUrl, 20*time.Millisecond, time.Hour, nil)
	go func() {
		for {
			select {
			case <-n.terminated:
				return
			case <-n.exit:
				n.l.UnlistenAll()
				n.l.Close()
				n.db.Close()
				for _, localL := range n.subscribers {
					close(localL)
				}
				close(n.terminated)
			case cmd := <-n.ch:
				cmd.fun()
				close(cmd.ack)
			case pgn := <-n.l.Notify:
				if pgn != nil {
					localN, err := n.makeNotification(pgn)
					if err != nil {
						log.Printf("Error parsing inbound notification %v: %v", pgn, err)
					} else {
						for _, sub := range n.subscribers {
							sub <- localN
						}
					}
				}
			}
		}
	}()

	return nil
}

// makeNotification converts a Postgres notification into a local notification.
func (n *Notifier) makeNotification(pn *pq.Notification) (*Notification, error) {
	localN := Notification{}

	id, err := strconv.Atoi(pn.Extra)
	if err == nil {
		localN.Id = int64(id)
	}

	localN.Type, err = n.getNotificationType(pn.Channel)
	return &localN, err
}

// command allows thread-safe subscribe/unsubscribe management.
type command struct {
	fun func()
	ack chan struct{}
}

// sendCommand delivers the func to the notifier, blocking the calling goroutine
// until the command is acknowledged as completed or the notifier exits.
func (n *Notifier) sendCommand(c func()) {
	cmd := command{c, make(chan struct{})}

	// send cmd
	select {
	case <-n.terminated:
		return
	case n.ch <- cmd:
	}

	// await ack (also handle termination)
	select {
	case <-n.terminated:
		return
	case <-cmd.ack:
		return
	}
}

// getNotificationType resolves the ntType presented by a string. It is
// symmetric with ntType.String(), which in turn is the Postgres-side
// notification channel name.
func (n *Notifier) getNotificationType(s string) (NtType, error) {
	value, ok := n.types[s]
	if !ok {
		return NtErrorFlag, fmt.Errorf("unregistered type '%s'", s)
	}
	return value, nil
}
