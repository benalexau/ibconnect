package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/benalexau/ibconnect/core"
)

// Handler returns an initialised Handler.
func Handler(errInfo bool, db *sql.DB, n *core.Notifier) http.Handler {
	u := &Util{
		ErrInfo: errInfo,
	}

	accountHandler := AccountHandler{u: u, db: db, n: n}
	null, _ := os.Open(os.DevNull)

	handler := rest.ResourceHandler{
		EnableGzip:               true,
		EnableRelaxedContentType: true,
		DisableJsonIndent:        false,
		EnableStatusService:      false,
		EnableResponseStackTrace: false,
		EnableLogAsJson:          true,
		Logger:                   log.New(null, "", 0),
	}

	var routes []*rest.Route

	routes = append(routes, &rest.Route{"GET", "/v1/accounts", accountHandler.GetAll})
	routes = append(routes, &rest.Route{"GET", "/v1/accounts/:accountCode", accountHandler.GetLatest})
	routes = append(routes, &rest.Route{"GET", "/v1/accounts/:accountCode/*timestamp", accountHandler.GetReport})

	handler.SetRoutes(routes...)

	var h http.Handler
	h = &handler
	return h
}
