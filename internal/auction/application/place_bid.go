package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	//config defer() to handles commit/rollback
	defer func() {
		if r := recover(); r != nil {
			log.Error("PlaceBidUseCase: Recovered from panic during transaction",
				zap.String("lotID", cmd.LotID.String()),
				zap.String("userID", cmd.UserID.String()),
				zap.Any("panic", r),
			)
			_ = tx.Rollback(ctx) // Rollback for panic case
			panic(r)
		}
		//if 'err' is not nil at the end of functions means an error occurs,
		// in some later step (GetByID, PlaceBid, Save), wich their own logs sentence,
		// here only logs the rollback
		if err != nil {
			log.Warn("PlaceBidUseCase: Rolling back transaction due to error",
				zap.String("lotID", cmd.LotID.String()),
				zap.String("userID", cmd.UserID.String()),
				zap.Error(err), // Log the error wich causes the error
			)
			_ = tx.Rollback(ctx) // Rollback if there is any error
			return               // Exit the defer func after rollback
		}
		// If we reach here, 'err' is nil, meaning no error occurred before the defer.
		// Attempt to commit the transaction.
		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			// if commits fails, log commit error
			log.Error("PlaceBidUseCase: Failed to commit transaction",
				zap.String("lotID", cmd.LotID.String()),
				zap.String("userID", cmd.UserID.String()),
				zap.Error(commitErr),
			)
			// Assign the commitError to 'err' variable to be returned by Execute() main function
			err = fmt.Errorf("place bid use case: failed to commit transaction: %w", commitErr)
		}
		//at this point the tx has beaing completed succefully
		log.Info("PlaceBidUseCase: Transaction committed successfully",
			zap.String("lotID", cmd.LotID.String()),
			zap.String("userID", cmd.UserID.String()))

	}()

	//3. Load AuctionLot aggregate inside TX
	lot, err := uc.lotRepo.GetByID(ctx, cmd.LotID)
	if err != nil {
		//if the error is ErrLotNotFound, is bussiner err, handled by infra layer
		// Si es otro error, logueamos aquí.
		if !errors.Is(err, domain.ErrLotNotFound) {
			log.Error("PlaceBidUseCase: Failed to get auction lot",
				zap.String("lotID", cmd.LotID.String()),
				zap.String("userID", cmd.UserID.String()),
				zap.Error(err),
			)
		}
		// Return the error (a domain or repository error)
		return nil, fmt.Errorf("place bid use case: failed to get auction lot %s: %w", cmd.LotID, err)
	}

	// 4. call domain method to make the bid, where the bussines logic is executed (validations, state updates
	// time extension). Domain returns new Bid entity if is succefully created
	minIncrement := 0.0 //temporal configuration
	newBid, err := lot.PlaceBid(cmd.UserID, cmd.Amount, minIncrement)
	if err != nil {
		return nil, fmt.Errorf("place bid use case: bid failed for lot %s: %w", cmd.LotID, err)
	}

	// 5. persist in repository methods inside TX
	err = uc.bidRepo.Save(ctx, tx, newBid)
	if err != nil {
		log.Error("PlaceBidUseCase: Failed to save new bid",
			zap.String("lotID", cmd.LotID.String()),
			zap.String("userID", cmd.UserID.String()),
			zap.String("bidID", newBid.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("place bid use case: failed to save new bid for lot %s: %w", cmd.LotID, err)
	}
	//save updated state of aggregate AuctionLot usin TX
	err = uc.lotRepo.Save(ctx, tx, lot)
	if err != nil {
		log.Error("PlaceBidUseCase: Failed to save updated auction lot",
			zap.String("lotID", cmd.LotID.String()),
			zap.String("userID", cmd.UserID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("place bid use case: failed to save updated auction lot %s: %w", cmd.LotID, err)
	}

	//6. if everthing goes right, defer() makes the commit, and then the newBid is returned
	return newBid, nil

}
