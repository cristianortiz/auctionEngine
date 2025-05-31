package main

import (
	"context"

	"github.com/cristianortiz/auctionEngine/internal/shared/db"
	"github.com/cristianortiz/auctionEngine/internal/shared/db/migrations"
	"github.com/cristianortiz/auctionEngine/internal/shared/httpserver"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"go.uber.org/zap"
)

func main() {
	// Inicializa logger
	logger := logger.GetLogger()
	defer logger.Sync()

	logger.Info("Starting AuctionEngine server...")

	// Ejecuta migraciones de base de datos
	logger.Info("Running database migrations...")
	if err := migrations.RunMigrations(); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully.")

	// Conexión a la base de datos (singleton)
	conn := db.GetPostgresDB()
	defer conn.Close(context.Background())

	// Aquí podrías pasar 'conn' a tus repositorios/servicios

	// Arranca el servidor HTTP
	server := httpserver.NewServer()
	if err := server.Start(":9000"); err != nil {
		logger.Fatal("HTTP server failed", zap.Error(err))
	}
}
