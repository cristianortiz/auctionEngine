package postgres

import (
	"context"
	"errors"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BidRepository implements domain.BidRepository interface
type BidRepository struct {
	pool *pgxpool.Pool
}

// NewBidRepository creates new instance of BidRepository.
func NewBidRepository(pool *pgxpool.Pool) *BidRepository {
	return &BidRepository{pool: pool}
}

// this method only inserts a new bid, the logic for the transaction for update the lot, will be created in application layer
func (r *BidRepository) Save(ctx context.Context, tx pgx.Tx, bid *domain.Bid) error {
	query := `
        INSERT INTO bids (id, lot_id, user_id, amount, timestamp, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := tx.Exec(ctx, query,
		bid.ID,
		bid.LotID,
		bid.UserID,
		bid.Amount,
		bid.Timestamp,
	)
	return err
}

func (r *BidRepository) GetBidsByLotID(ctx context.Context, lotID uuid.UUID) ([]*domain.Bid, error) {
	query := `
        SELECT id, lot_id, user_id, amount, timestamp, created_at
        FROM bids
        WHERE lot_id = $1
        ORDER BY timestamp ASC 
    `
	rows, err := r.pool.Query(ctx, query, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []*domain.Bid
	for rows.Next() {
		bid := &domain.Bid{}
		err := rows.Scan(
			&bid.ID,
			&bid.LotID,
			&bid.UserID,
			&bid.Amount,
			&bid.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bids, nil
}

func (r *BidRepository) GetLatestBidByLotID(ctx context.Context, lotID uuid.UUID) (*domain.Bid, error) {
	query := `
        SELECT id, lot_id, user_id, amount, timestamp, created_at
        FROM bids
        WHERE lot_id = $1
        ORDER BY timestamp DESC
        LIMIT 1
    `
	bid := &domain.Bid{}
	err := r.pool.QueryRow(ctx, query, lotID).Scan(
		&bid.ID,
		&bid.LotID,
		&bid.UserID,
		&bid.Amount,
		&bid.Timestamp,
	)

	if err != nil {
		//if there is any bid por this lot
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return bid, nil
}
