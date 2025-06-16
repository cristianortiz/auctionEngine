package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuctionLotRepository implements domain.AuctionLotRepository interface
type AuctionLotRepository struct {
	pool *pgxpool.Pool
}

// NewAuctionLotRepository creates a new instance of AuctionRepository
func NewAuctionLotRepository(pool *pgxpool.Pool) *AuctionLotRepository {
	return &AuctionLotRepository{pool: pool}
}

// Save guarda o actualiza un AuctionLot en la base de datos.
// Utiliza INSERT ON CONFLICT para manejar tanto la creación como la actualización.
// Omitimos created_at y updated_at en el INSERT inicial para usar los DEFAULT/TRIGGER de la DB.
func (r *AuctionLotRepository) Save(ctx context.Context, tx pgx.Tx, lot *domain.AuctionLot) error {
	query := `
        INSERT INTO auction_lots (id, title, description, initial_price, current_price, end_time, state, last_bid_time, time_extension)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO UPDATE
        SET
            title = EXCLUDED.title,
            description = EXCLUDED.description,
            initial_price = EXCLUDED.initial_price,
            current_price = EXCLUDED.current_price,
            end_time = EXCLUDED.end_time,
            state = EXCLUDED.state,
            last_bid_time = EXCLUDED.last_bid_time,
            time_extension = EXCLUDED.time_extension,
            updated_at = NOW(); 
    `
	_, err := tx.Exec(ctx, query,
		lot.ID,
		lot.Title,
		lot.Description,
		lot.InitialPrice,
		lot.CurrentPrice,
		lot.EndTime,
		lot.State,
		lot.LastBidTime,
		lot.TimeExtension,
	)
	return err
}

// GetByID recupera un AuctionLot por su ID.
// Incluimos created_at y updated_at en el SELECT y SCAN.
func (r *AuctionLotRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuctionLot, error) {
	query := `
        SELECT id, title, description, initial_price, current_price, end_time, state, last_bid_time, time_extension, created_at, updated_at
        FROM auction_lots
        WHERE id = $1
    `
	lot := &domain.AuctionLot{}
	var lastBidTime *time.Time // pointer to handle NULL

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&lot.ID,
		&lot.Title,
		&lot.Description,
		&lot.InitialPrice,
		&lot.CurrentPrice,
		&lot.EndTime,
		&lot.State,
		&lastBidTime, // scan pointer
		&lot.TimeExtension,
		&lot.CreatedAt, // Incluido en SCAN
		&lot.UpdatedAt, // Incluido en SCAN
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLotNotFound // Usar error del dominio
		}
		return nil, err
	}

	lot.LastBidTime = lastBidTime // Corregido: asignar el puntero directamente

	return lot, nil
}

// GetActiveLots recupera todos los lotes de subasta activos.
// Incluimos created_at y updated_at en el SELECT y SCAN.
func (r *AuctionLotRepository) GetActiveLots(ctx context.Context) ([]*domain.AuctionLot, error) {
	query := `
        SELECT id, title, description, initial_price, current_price, end_time, state, last_bid_time, time_extension, created_at, updated_at
        FROM auction_lots
        WHERE state = $1
    `
	rows, err := r.pool.Query(ctx, query, domain.StateActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lots []*domain.AuctionLot
	for rows.Next() {
		lot := &domain.AuctionLot{}
		var lastBidTime *time.Time
		err := rows.Scan(
			&lot.ID,
			&lot.Title,
			&lot.Description,
			&lot.InitialPrice,
			&lot.CurrentPrice,
			&lot.EndTime,
			&lot.State,
			&lastBidTime,
			&lot.TimeExtension,
			&lot.CreatedAt, // Incluido en SCAN
			&lot.UpdatedAt, // Incluido en SCAN
		)
		if err != nil {
			return nil, err
		}
		lot.LastBidTime = lastBidTime // Corregido: asignar el puntero directamente
		lots = append(lots, lot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lots, nil
}

// GetLotsEndingSoon recupera lotes activos que terminan pronto.
// 'threshold' define cuánto tiempo antes del fin se consideran "ending soon".
// Incluimos created_at y updated_at en el SELECT y SCAN.
func (r *AuctionLotRepository) GetLotsEndingSoon(ctx context.Context, threshold time.Duration) ([]*domain.AuctionLot, error) {
	query := `
        SELECT id, title, description, initial_price, current_price, end_time, state, last_bid_time, time_extension, created_at, updated_at
        FROM auction_lots
        WHERE state = $1 AND end_time <= NOW() + $2
    `
	rows, err := r.pool.Query(ctx, query, domain.StateActive, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lots []*domain.AuctionLot
	for rows.Next() {
		lot := &domain.AuctionLot{}
		var lastBidTime *time.Time
		err := rows.Scan(
			&lot.ID,
			&lot.Title,
			&lot.Description,
			&lot.InitialPrice,
			&lot.CurrentPrice,
			&lot.EndTime,
			&lot.State,
			&lastBidTime,
			&lot.TimeExtension,
			&lot.CreatedAt, // Incluido en SCAN
			&lot.UpdatedAt, // Incluido en SCAN
		)
		if err != nil {
			return nil, err
		}
		lot.LastBidTime = lastBidTime // Corregido: asignar el puntero directamente
		lots = append(lots, lot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lots, nil
}
