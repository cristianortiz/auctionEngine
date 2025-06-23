package application

import (
	"context"

	"github.com/cristianortiz/auctionEngine/internal/auction/domain"
)

// AuctionService defines application interface layer of auction module
// exposes uses cases to external layer, aka infra
type AuctionService interface {
	// Placebid handles logic when a user makes a bid in a lot
	// receives a command with necesary data and returns the created bid or an error
	PlaceBid(ctx context.Context, cmd PlaceBidDTO) (*domain.Bid, error)
}

// concret implementation of AuctionService (struct)
type auctionService struct {
	placeBidUC *PlaceBidUseCase
}

func NewAuctionService(placeBidUC *PlaceBidUseCase) AuctionService {
	return &auctionService{
		placeBidUC: placeBidUC,
	}
}

// PlaceBid implements AuctionService.
func (as *auctionService) PlaceBid(ctx context.Context, cmd PlaceBidDTO) (*domain.Bid, error) {
	return as.placeBidUC.Execute(ctx, cmd)
}
