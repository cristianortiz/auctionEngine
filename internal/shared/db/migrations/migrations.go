package migrations

import (
	"github.com/cristianortiz/auctionEngine/internal/shared/db"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

var log = logger.GetLogger() // Instancia logger para el pakg

func RunMigrations() error {
	dbURL := db.BuildPostgresDSN()
	log.Info("RunMigrations",
		zap.String("posgresUrl", dbURL))
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
