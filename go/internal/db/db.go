package db

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// TODO: better abstraction...
func NewBunPostgresClient(connectionString string) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connectionString)))

	db := bun.NewDB(sqldb, pgdialect.New())

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	return db
}
