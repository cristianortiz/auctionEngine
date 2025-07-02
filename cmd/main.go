package main

import (
	"context"
	"os"

	"github.com/cristianortiz/auctionEngine/internal/auction/application"
	"github.com/cristianortiz/auctionEngine/internal/auction/infra/repository/postgres"
	wsh "github.com/cristianortiz/auctionEngine/internal/auction/infra/websocket"
	"github.com/cristianortiz/auctionEngine/internal/shared/db"
	"github.com/cristianortiz/auctionEngine/internal/shared/db/migrations"
	"github.com/cristianortiz/auctionEngine/internal/shared/httpserver"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/cristianortiz/auctionEngine/internal/shared/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("HTTP_PORT")
	log := logger.GetLogger()
	defer log.Sync()

	log.Info("Starting AuctionEngine server...")

	log.Info("Running database migrations...")
	if err := migrations.RunMigrations(); err != nil {
		log.Fatal("Database migration failed", zap.Error(err))
	}
	log.Info("Database migrations completed successfully.")

	dbPool, err := db.GetPostgresDBPool(context.Background())
	if err != nil {
		log.Fatal("failed to connect to DBpool", zap.Error(err))
	}
	defer dbPool.Close()
	log.Info("DB pool connected")

	//--- Init repositorys ----
	lotRepo := postgres.NewAuctionLotRepository(dbPool)
	log.Info("Lot repository initialized")
	bidRepo := postgres.NewBidRepository(dbPool)
	log.Info("Lot repository initialized")

	//--- Init uses cases
	placeBidUC := application.NewPlaceBidUseCase(lotRepo, bidRepo, dbPool)
	getLostStateUC := application.NewGetLotStateUseCase(lotRepo, bidRepo)

	//---Init app service
	auctionService := application.NewAuctionService(placeBidUC, getLostStateUC)

	//-- Init webSocket hub and runs it in a goroutine
	hub := websocket.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	//-- init handler, remember this came from Ws handler internal/infra/websocket
	auctionWSHandler := wsh.NewAuctionWSHandler(auctionService, hub)
	go auctionWSHandler.ListenForMessages(ctx)
	log.Info("WebSocket Hub started.")

	server := httpserver.NewServer(":"+port, hub, ctx)
	if err := server.Start(":" + port); err != nil {
		log.Fatal("HTTP server failed", zap.Error(err))
	}
}
