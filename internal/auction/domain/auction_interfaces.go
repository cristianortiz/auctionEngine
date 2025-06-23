package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuctionLotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*AuctionLot, error)
	Save(ctx context.Context, tx pgx.Tx, lot *AuctionLot) error
	GetActiveLots(ctx context.Context) ([]*AuctionLot, error)
	GetLotsEndingSoon(ctx context.Context, threshold time.Duration) ([]*AuctionLot, error)
}

type BidRepository interface {
	Save(ctx context.Context, tx pgx.Tx, bid *Bid) error
	GetBidsByLotID(ctx context.Context, lotID uuid.UUID) ([]*Bid, error)
	GetLatestBidByLotID(ctx context.Context, lotID uuid.UUID) (*Bid, error)
}
