package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gorhill/cronexpr"
)

// Config represents the applicable configuration variables.
type Config struct {
	ErrInfo        bool
	IbGws          []string
	IbClientId     int
	DbUrl          string
	Port           int
	Host           string
	AccountRefresh *cronexpr.Expression
}

// Address returns the HTTP bind address.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// NewConfig parses the relevant environment variables, applying defaults if unspecified.
func NewConfig() (Config, error) {
	c := Config{}
	c.ErrInfo = os.Getenv("ERR_INFO") == "true"

	ibGw := os.Getenv("IB_GW")
	if ibGw == "" {
		ibGw = "127.0.0.1:4002"
	}
	c.IbGws = strings.Split(ibGw, ",")

	ibClientId := os.Getenv("IB_CID")
	if ibClientId == "" {
		ibClientId = "5555"
	}

	var err error
	c.IbClientId, err = strconv.Atoi(ibClientId)
	if err != nil {
		return c, fmt.Errorf("IB_CID '%s' not an integer")
	}

	c.DbUrl = os.Getenv("DB_URL")
	if c.DbUrl == "" {
		c.DbUrl = "postgres://ibc_dev@localhost/ibc_dev?sslmode=disable"
	}

	if !strings.HasPrefix(c.DbUrl, "postgres://") {
		return c, fmt.Errorf("DB_URL '%s' did not being with postgres://", c.DbUrl)
	}

	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "3000"
	}

	c.Port, err = strconv.Atoi(portString)
	if err != nil {
		return c, fmt.Errorf("PORT '%s' not an integer")
	}

	c.Host = os.Getenv("HOST")
	if c.Host == "" {
		c.Host = "localhost"
	}

	acctRefresh := os.Getenv("ACCT_REF")
	if acctRefresh == "" {
		acctRefresh = "@hourly"
	}
	c.AccountRefresh, err = cronexpr.Parse(acctRefresh)
	if err != nil {
		return c, err
	}

	return c, nil
}
