package core

import (
	"database/sql"
	"sync"
)

// Context initializes key application dependencies and makes them available.
type Context struct {
	closed sync.Once
	N      *Notifier
	DL     *DistLock
	DB     *sql.DB
}

// NewContext prepares the application context using the passed Config.
func NewContext(c Config) (*Context, error) {
	n, err := NewNotifier(c.DbUrl)
	if err != nil {
		return nil, err
	}

	err = n.RegisterAll(NtTypes())
	if err != nil {
		return nil, err
	}

	dl, err := NewDistLock(c.DbUrl)
	if err != nil {
		return nil, err
	}

	db, err := InitMeddler(c.DbUrl)
	if err != nil {
		return nil, err
	}

	return &Context{
		N:  n,
		DL: dl,
		DB: db}, nil
}

// Close releases all resources. It is safe to call multiple times.
func (c *Context) Close() {
	c.closed.Do(func() {
		defer c.N.Close()
		defer c.DL.Close()
		defer c.DB.Close()
	})
}
