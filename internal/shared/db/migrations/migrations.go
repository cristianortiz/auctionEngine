package migrations

import (
	"github.com/cristianortiz/auctionEngine/internal/shared/db"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() error {
	dbURL := db.BuildPostgresDSN()
	m, err := migrate.New(
		"file://internal/shared/db/migrations/sql",
		dbURL,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
