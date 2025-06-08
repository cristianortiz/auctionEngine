package main

import (
	"context"
	"os"

	"github.com/cristianortiz/auctionEngine/internal/shared/db"
	"github.com/cristianortiz/auctionEngine/internal/shared/db/migrations"
	"github.com/cristianortiz/auctionEngine/internal/shared/httpserver"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/cristianortiz/auctionEngine/internal/shared/websocket"
	userRepository "github.com/cristianortiz/auctionEngine/internal/user/infrastructure/repository/postgres"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("HTTP_PORT")
	logger := logger.GetLogger()
	defer logger.Sync()

	logger.Info("Starting AuctionEngine server...")

	logger.Info("Running database migrations...")
	if err := migrations.RunMigrations(); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully.")

	conn := db.GetPostgresDB()
	defer conn.Close(context.Background())

	//init user repository
	userRepo := userRepository.NewUserRepository(conn)
	logger.Info("User repository initialized")

	//init webSocket hub and runs it in a goroutine
	hub := websocket.NewHub()
	go hub.Run()
	logger.Info("WebSocket Hub started.")

	server := httpserver.NewServer(":"+port, hub)
	if err := server.Start(":" + port); err != nil {
		logger.Fatal("HTTP server failed", zap.Error(err))
	}
}
