package application

import (
	"context"
	"fmt"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// PlaceBidDTO is DTO input for PlaceBid useCase, contains the necesary data to make a bid
type PlaceBidDTO struct {
	LotID  uuid.UUID
	UserID uuid.UUID
	Amount float64
}

// PlaceBidUseCase is useCase to make a bid in an auction lot, orchestrate bussines logic and persistence
type PlaceBidUseCase struct {
	lotRepo domain.AuctionLotRepository
	bidRepo domain.BidRepository
	dbPool  *pgxpool.Pool
	// userRepo domain.UserRepository // maybe useful to validates the UserID existence
}

// NewPlaceBidUseCase creates a new instace of PlaceBidUseCase struct, it receives dependency through injection
func NewPlaceBidUseCase(lotRepo domain.AuctionLotRepository,
	bidRepo domain.BidRepository,
	dbPool *pgxpool.Pool) *PlaceBidUseCase {

	return &PlaceBidUseCase{
		lotRepo: lotRepo,
		bidRepo: bidRepo,
		dbPool:  dbPool,
	}

}

func (uc *PlaceBidUseCase) Execute(ctx context.Context, cmd PlaceBidDTO) (*domain.Bid, error) {
	log.Info("Executing PlaceBidUseCase",
		zap.String("lotID", cmd.LotID.String()),
		zap.String("userID", cmd.UserID.String()),
		zap.Float64("amount", cmd.Amount),
	)
	// 1. validates input DTO (basics validations, relative to the input data, not bussiles logic)
	if cmd.Amount <= 0 {
		log.Warn("PlaceBidUseCase: Invalid bid amount",
			zap.String("lotID", cmd.LotID.String()),
			zap.String("userID", cmd.UserID.String()),
			zap.Float64("amount", cmd.Amount),
		)
		return nil, domain.ErrInvalidAmount // Asegúrate de que ErrInvalidAmount esté definido en domain/errors.go
	}
	//TODO: maybe validates if UserID exists using userRepo.GetByID()

	//2. starts a DB TX, to ensures an atomic operations for save the bid and upates de lot
	tx, err := uc.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("PlaceBidUseCase: Failed to begin transaction",
			zap.String("lotID", cmd.LotID.String()),
			zap.String("userID", cmd.UserID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("place bid use case: failed to begin transaction: %w", err)
	}

}
