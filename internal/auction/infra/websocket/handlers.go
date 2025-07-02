package websocket

import (
	"context"
	"encoding/json"

	"github.com/cristianortiz/auctionEngine/internal/auction/application"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/cristianortiz/auctionEngine/internal/shared/websocket"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// AuctionWSHandler handles the ws inbound msgs wich are specific for auction module (remember is a bounded context)
type AuctionWSHandler struct {
	auctionService application.AuctionService // application layer dependency
	hub            *websocket.Hub             // shared hub dependency to send msgs
}

// NewAuctionWSHandler creates a new instance of AuctionWSHandler
func NewAuctionWSHandler(auctionService application.AuctionService, hub *websocket.Hub) *AuctionWSHandler {
	return &AuctionWSHandler{
		auctionService: auctionService,
		hub:            hub,
	}
}

// ListenForMessages starts a go routine that listen the Hub inbound channel for messages and proccess every one of them
func (h *AuctionWSHandler) ListenForMessages(ctx context.Context) {
	log.Info("AuctionWSHandler started listening for inbound messages from hub")
	for {
		select {
		case <-ctx.Done():
			log.Info("AuctionWSHandler stopped listening for inbound messages from hub")
			return
		case msg := <-h.hub.InboundMessages:
			go h.processMessage(ctx, msg.Client, msg.Data)
		}
	}

}

// processMesssage dispatch the message by this type
func (h *AuctionWSHandler) processMessage(ctx context.Context, client *websocket.Client, data []byte) {
	var baseMsg BaseMessage
	if err := json.Unmarshal(data, &baseMsg); err != nil {
		h.sendErrorToClient(client, "invalid message format")
		return
	}
	switch baseMsg.Type {
	case MessageTypeClientBid:
		h.handleClientBidMessage(ctx, client, data)
	//adds more case for other types of messages
	default:
		h.sendErrorToClient(client, "unknown message type")
	}
}

func (h *AuctionWSHandler) handleClientBidMessage(ctx context.Context, client *websocket.Client, data []byte) {
	var bidMsg ClientBidMessage
	if err := json.Unmarshal(data, &bidMsg); err != nil {
		h.sendErrorToClient(client, "invalid bid message format")
		return
	}

	//validates LotId
	if bidMsg.Payload.LotID.String() != client.LotID {
		h.sendErrorToClient(client, "lot ID mismatch")
		return
	}

	cmd := application.PlaceBidDTO{
		LotID:  bidMsg.Payload.LotID,
		UserID: bidMsg.Payload.UserID,
		Amount: bidMsg.Payload.Amount,
	}
	_, err := h.auctionService.PlaceBid(ctx, cmd)
	if err != nil {
		h.sendErrorToClient(client, err.Error())
		return
	}

	//1. get updated lot state
	lotState, err := h.auctionService.GetLotState(ctx, cmd.LotID)
	if err != nil {
		h.sendErrorToClient(client, "failed to get updated lost state")
		return
	}
	//2. build update message
	updateMsg := ServerLotUpdateMessage{
		BaseMessage: BaseMessage{
			Type: MessageTypeServerLotUpdate,
		},
	}
	updateMsg.Payload.LotID = lotState.LotID
	updateMsg.Payload.CurrentPrice = lotState.CurrentPrice
	updateMsg.Payload.EndTime = lotState.EndTime
	updateMsg.Payload.State = lotState.State
	updateMsg.Payload.LastBidAmount = lotState.LastBidAmount
	updateMsg.Payload.LastBidUserID = lotState.LastBidUserID
	updateMsg.Payload.LastBidTime = lotState.LastBidTime

	// 3. serialize and send to all lot clients
	updateDate, err := json.Marshal(updateMsg)
	if err != nil {
		h.sendErrorToClient(client, "failed to serialize lot update")
		return
	}
	h.hub.BroadcastMessageToLot(client.LotID, updateDate)

}

// sendErrorToClient serializes and sends an error msg to a specific client
func (h *AuctionWSHandler) sendErrorToClient(client *websocket.Client, errorMessage string) {
	errMsg := ServerErrorMessage{
		BaseMessage: BaseMessage{MessageTypeServerError},
	}
	errMsg.Payload.Error = errorMessage
	data, err := json.Marshal(errMsg)
	if err != nil {
		log.Error("failed to marshal ServerErrorMessage", zap.Error(err))
		return
	}
	select {
	case client.Send <- data:
		log.Debug("sent error message to client")
	default:
		log.Warn("client send channel full or closed, could not send error msg")
	}
}
