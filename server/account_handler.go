package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/benalexau/ibconnect/core"
	"github.com/russross/meddler"
)

type AccountHandler struct {
	db *sql.DB
	n  *core.Notifier
	u  *Util
}

type AccountReport struct {
	Balance   core.AccountAmount
	Positions []*core.AccountPositionView
}

func (a *AccountHandler) GetAll(w rest.ResponseWriter, r *rest.Request) {
	err := RefreshIfNeeded(a.n, r, core.NtAccountRefresh, core.NtAccountFeedDone, 15*time.Second)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	var accounts []*core.Account
	err = meddler.QueryAll(a.db, &accounts, "SELECT * FROM account")
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}
	w.Header().Add("Cache-Control", "private, max-age=60")
	w.WriteJson(&accounts)
}

func (a *AccountHandler) GetLatest(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("accountCode")
	latest := new(core.AccountSnapshotLatest)
	err := meddler.QueryRow(a.db, latest, "SELECT * FROM v_account_snapshot_latest WHERE account_code = $1", code)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	path := fmt.Sprintf("/v1/accounts/%s/%s", code, latest.Latest.Format(time.RFC3339Nano))
	url := r.UrlFor(path, make(map[string][]string))
	w.Header().Add("Location", url.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (a *AccountHandler) GetReport(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("accountCode")
	timestamp := r.PathParam("timestamp")
	existing := new(core.Account)
	err := meddler.QueryRow(a.db, existing, "SELECT * FROM account WHERE account_code = $1", code)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	created, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	var snap core.AccountSnapshot
	err = meddler.QueryRow(a.db, &snap, "SELECT * FROM account_snapshot WHERE account_id = $1 AND created = $2", existing.Id, created)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	var report AccountReport
	err = meddler.QueryAll(a.db, &report.Positions, "SELECT * FROM v_account_position WHERE account_snapshot_id = $1", snap.Id)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	err = meddler.QueryRow(a.db, &report.Balance, "SELECT * FROM account_amount WHERE account_snapshot_id = $1", snap.Id)
	if err != nil {
		a.u.HandleError(err, w, r)
		return
	}

	w.Header().Add("Cache-Control", "private, max-age=31556926")
	w.WriteJson(&report)
}
