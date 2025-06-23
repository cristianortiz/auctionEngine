package websocket

import (
	"context"

	"github.com/cristianortiz/auctionEngine/internal/auction/application"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/cristianortiz/auctionEngine/internal/shared/websocket"
)

var log = logger.GetLogger()

// AuctionWSHandler handles the ws inbound msgs wich are specific for auction module (remember is a bounded context)
type AuctionWSHandler struct {
	auctionService application.AuctionService // application layer dependency
	hub            *websocket.Hub             // shared hub dependency to send msgs
}

// NewAuctionWSHandler creates a new instance of AuctionWSHandler
func NewAuctionWSHandler(auctionService *application.AuctionService, hub *websocket.Hub) *AuctionWSHandler {
	return &AuctionWSHandler{
		auctionService: *auctionService,
		hub:            hub,
	}
}

// ListenForMessages starts a go routine that listen the Hub inbound channel for messages and proccess every one of them
func (h *AuctionWSHandler) ListenForMessages(ctx context.Context) {
	log.Info("AuctionWSHandler started listening for inbound messages from hub")

}
