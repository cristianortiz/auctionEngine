package application

import (
	"context"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/google/uuid"
)

// LotStateDTO is the output DTO for exposing lot state to the UI/WS
type LotStateDTO struct {
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
}

// GetLotStateUseCase retrieves the current state of and auction lot
type GetLotStateUseCase struct {
	lotRepo domain.AuctionLotRepository
	bidRepo domain.BidRepository
}

// NewGetLotStateUseCase creates a new instance of GetLotStateUseCase.
func NewGetLotStateUseCase(lotRepo domain.AuctionLotRepository, bidRepo domain.BidRepository) *GetLotStateUseCase {
	return &GetLotStateUseCase{
		lotRepo: lotRepo,
		bidRepo: bidRepo,
	}
}

func (uc *GetLotStateUseCase) Execute(ctx context.Context, lotID uuid.UUID) (*LotStateDTO, error) {
	lot, err := uc.lotRepo.GetByID(ctx, lotID)
	if err != nil {
		return nil, err
	}

	dto := &LotStateDTO{
		LotID:        lot.ID,
		Title:        lot.Title,
		Description:  lot.Description,
		InitialPrice: lot.InitialPrice,
		CurrentPrice: lot.CurrentPrice,
		EndTime:      lot.EndTime,
		State:        string(lot.State),
		LastBidTime:  lot.LastBidTime,
	}

	// Optionally, get the latest bid for more details
	bid, err := uc.bidRepo.GetLatestBidByLotID(ctx, lotID)
	if err == nil && bid != nil {
		dto.LastBidAmount = bid.Amount
		dto.LastBidUserID = bid.UserID
		dto.LastBidTime = &bid.Timestamp
	}

	return dto, nil
}
