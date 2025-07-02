package websocket

import (
	"time"

	"github.com/google/uuid"
)

// MessageType defines ws type message
type MessageType string

const (
	MessageTypeClientBid          MessageType = "client_bid"           // client msg to make a bid
	MessageTypeServerLotUpdate    MessageType = "server_lot_update"    // server  msg with lot update
	MessageTypeServerError        MessageType = "server_error"         // server msg indicating error
	MessageTypeServerInfo         MessageType = "server_info"          // server msg with general info
	MessageTypeClientJoinLot      MessageType = "client_join_lot"      // client msg to join a lot (optional if the path is no used)
	MessageTypeServerInitialState MessageType = "server_initial_state" // server msgw with lot initial state
)

// BaseMessage is base struct for all the WS messages, includes a Type field for identify the message type
type BaseMessage struct {
	Type MessageType `json:"type"`
}

// ClientBidMessage is DTO for a bid message sended vy the client
type ClientBidMessage struct {
	BaseMessage
	Payload struct {
		LotID  uuid.UUID `json:"lot_id"`
		UserID uuid.UUID `json:"user_id"`
		Amount float64   `json:"amount"`
	} `json:"payload"`
}

// ServerLotUpdateMessage is DTO for a lot update msg sended by the server
type ServerLotUpdateMessage struct {
	BaseMessage
	Payload struct {
		LotID         uuid.UUID  `json:"lot_id"`
		CurrentPrice  float64    `json:"current_price"`
		EndTime       time.Time  `json:"end_time"`
		State         string     `json:"state"` // Use string for domain state
		LastBidAmount float64    `json:"last_bid_amount,omitempty"`
		LastBidUserID uuid.UUID  `json:"last_bid_user_id,omitempty"`
		LastBidTime   *time.Time `json:"last_bid_time,omitempty"`
	} `json:"payload"`
}

type ServerErrorMessage struct {
	BaseMessage
	Payload struct {
		Error string `json:"error"`
	} `json:"payload"`
}

// ServerInfoMessage es el DTO para un mensaje de información general enviado por el servidor.
type ServerInfoMessage struct {
	BaseMessage
	Payload struct {
		Message string `json:"message"`
	} `json:"payload"`
}

// ServerInitialStateMessage es el DTO para el estado inicial del lote enviado al cliente al conectarse.
type ServerInitialStateMessage struct {
	BaseMessage
	Payload struct {
		LotID         uuid.UUID  `json:"lot_id"`
		Title         string     `json:"title"`
		Description   string     `json:"description"`
		InitialPrice  float64    `json:"initial_price"`
		CurrentPrice  float64    `json:"current_price"`
		EndTime       time.Time  `json:"end_time"`
		State         string     `json:"state"`
		LastBidAmount float64    `json:"last_bid_amount,omitempty"`
		LastBidUserID uuid.UUID  `json:"last_bid_user_id,omitempty"`
		LastBidTime   *time.Time `json:"last_bid_time,omitempty"`
		// maybe include a list of recents bids here
		// RecentBids []*BidDTO `json:"recent_bids,omitempty"` //BidDTO needed
	} `json:"payload"`
}

// type BidDTO struct {
// 	ID uuid.UUID `json:"id"`
// 	UserID uuid.UUID `json:"user_id"`
// 	Amount float64 `json:"amount"`
// 	Timestamp time.Time `json:"timestamp"`
// }
