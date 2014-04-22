package server

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

type Util struct {
	ErrInfo bool
}

// HandleError will return an appropriate HTTP message in response to an error.
// This includes special handling of SQL not found errors and emitting log
// messages with correlation identifiers.
func (u *Util) HandleError(err error, w rest.ResponseWriter, r *rest.Request) {
	if err == sql.ErrNoRows {
		rest.NotFound(w, r)
	}
	id := u.uuid()
	log.Printf("%v [%s] [%v]", err, id, r.URL)

	errInfo := make(map[string]string)
	errInfo["error_id"] = id

	if u.ErrInfo {
		errInfo["details"] = err.Error()
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.WriteJson(errInfo)
}

// uuid makes a simple random UUID.
func (u *Util) uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
