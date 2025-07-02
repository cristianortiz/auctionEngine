package application

import (
	"context"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
	"github.com/google/uuid"
)

// AuctionService defines application interface layer of auction module
// exposes uses cases to external layer, aka infra
type AuctionService interface {
	// Placebid handles logic when a user makes a bid in a lot
	// receives a command with necesary data and returns the created bid or an error
	PlaceBid(ctx context.Context, cmd PlaceBidDTO) (*domain.Bid, error)
	GetLotState(ctx context.Context, lotID uuid.UUID) (*LotStateDTO, error)
}

// concret implementation of AuctionService (struct)
type auctionService struct {
	placeBidUC    *PlaceBidUseCase
	getLotStateUC *GetLotStateUseCase
}

func NewAuctionService(placeBidUC *PlaceBidUseCase, getLotStateUC *GetLotStateUseCase) AuctionService {
	return &auctionService{
		placeBidUC:    placeBidUC,
		getLotStateUC: getLotStateUC,
	}
}

// PlaceBid implements AuctionService.
func (as *auctionService) PlaceBid(ctx context.Context, cmd PlaceBidDTO) (*domain.Bid, error) {
	return as.placeBidUC.Execute(ctx, cmd)
}

// GetLotState to implementss AuctionService
func (as *auctionService) GetLotState(ctx context.Context, lotID uuid.UUID) (*LotStateDTO, error) {
	return as.getLotStateUC.Execute(ctx, lotID)
}
