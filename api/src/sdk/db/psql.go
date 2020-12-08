package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // import psql driver

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

// NewPSQLConn initiates a new connection to a postgresql server
func NewPSQLConn(
	dbURL string,
	maxOpenConns int,
	maxIdleConns int,
	connMaxLifetime time.Duration,
) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("could not open conn to postgresql (%v)", err)
	}

	// IddleConns and OpenConns should evolve depending of each other
	// we don't want a service to monopolise connections, we find more secure
	// to open new conn because we don't have performance issue nowadays
	db.SetMaxIdleConns(maxIdleConns)
	// some service might need a higher number of max open conns according to their purposee
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database (%v)", err)
	}
	return db, nil
}

// Rollback ...
func Rollback(ctx context.Context, tx *sql.Tx, msg string) {
	err := tx.Rollback()
	if err != nil {
		logger.FromCtx(ctx).Error().Msgf("could not rollback '%s': %s", msg, err.Error())
	}
}
