package domain

import (
	"context"

	"github.com/google/uuid"
)

type AuctionLotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (AuctionLot, error)
	Save(ctx context.Context, lot *AuctionLot) error
}

// BidRepository interface and repo is necesary for provide access to bid ops, even if Bid is and agreggates of AuctionLot type
// for now a save() methos is enough
type BidRepository interface {
	Save(ctx context.Context, bid *Bid) error
	//get all the bids for a specific lot
	GetByLotID(ctx context.Context, lotID uuid.UUID, limit int) ([]*Bid, error)
}
