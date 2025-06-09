package domain

import (
	"time"

	"github.com/google/uuid"
)

// bid represents individual bid in an auction lot
// is also an entity inside AuctionLot agreggate (DDD concepts)
type Bid struct {
	ID        uuid.UUID
	LotID     uuid.UUID
	UserID    uuid.UUID //users id who makes the bid
	Amount    float64
	Timestamp time.Time
}

// NewBid creates a new Bid instance
func NewBid(id, lotID, userID uuid.UUID, amount float64, timestamp time.Time) *Bid {
	return &Bid{
		ID:        id,
		LotID:     lotID,
		UserID:    userID,
		Amount:    amount,
		Timestamp: timestamp,
	}

}
