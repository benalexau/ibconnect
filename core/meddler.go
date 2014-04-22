package core

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/russross/meddler"
)

// InitMeddler configures Meddler, including registration of custom Meddlers.
// It runs a database connection for general-purpose use.
func InitMeddler(dbUrl string) (*sql.DB, error) {
	meddler.Default = meddler.PostgreSQL
	meddler.Debug = true
	meddler.Register("monetary", MonetaryMeddler{})
	return sql.Open("postgres", dbUrl)
}
