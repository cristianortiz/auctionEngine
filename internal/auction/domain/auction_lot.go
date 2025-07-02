package domain

import (
	"sync"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// AuctinLotState represents the actual state of a lot auction
type AuctionLotState string

const (
	StatePending   AuctionLotState = "pending"
	StateActive    AuctionLotState = "active"
	StateFinished  AuctionLotState = "finished"
	StateCancelled AuctionLotState = "cancelled"
)

type AuctionLot struct {
	ID            uuid.UUID
	Title         string
	Description   string
	InitialPrice  float64
	CurrentPrice  float64
	EndTime       time.Time
	State         AuctionLotState
	LastBidTime   *time.Time    //for time extension logic
	TimeExtension time.Duration // time extension period  for bid
	CreatedAt     time.Time
	UpdatedAt     time.Time
	//to protect concurrent state of lot during bids flow
	//very important for thread safety in concurrent environment (websockets)
	mu sync.Mutex
	//list of bids associeted whit this lot, for simplicity we take it all in this MVP
	Bids []*Bid
}

func NewAuctionLot(id uuid.UUID, title, description string, initialPrice float64, endTime time.Time, timeExtension time.Duration) *AuctionLot {
	return &AuctionLot{
		ID:            id,
		Title:         title,
		InitialPrice:  initialPrice,
		CurrentPrice:  initialPrice, //current price starts at initial price
		EndTime:       endTime,
		State:         StatePending, //starts pendind
		TimeExtension: timeExtension,
		Bids:          []*Bid{},
	}
}

func (al *AuctionLot) PlaceBid(userID uuid.UUID, amount float64, minIncrement float64) (*Bid, error) {
	//blocks concurrent acces to lot state
	al.mu.Lock()
	//ensures the mutex is released when function ends
	defer al.mu.Unlock()
	//bussiles logic validations
	if al.State != StateActive {
		log.Warn("Bid rejected: Lot not active",
			zap.String("lotID", al.ID.String()),
			zap.String("state", string(al.State)),
			zap.Float64("bidAmount", amount),
			zap.String("userID", userID.String()),
		)
		return nil, ErrLotNotActive
	}

	if amount <= al.CurrentPrice {
		log.Warn("Bid rejected: Amount too low",
			zap.String("lotID", al.ID.String()),
			zap.Float64("bidAmount", amount),
			zap.Float64("currentPrice", al.CurrentPrice),
			zap.String("userID", userID.String()),
		)
		return nil, ErrBidAmountTooLow
	}

	// Optional: validates minimum increment
	// if amount < al.CurrentPrice + minIncrement {
	// 	log.Warn("Bid rejected: Increment too small",
	// 		zap.String("lotID", al.ID.String()),
	// 		zap.Float64("bidAmount", amount),
	// 		zap.Float64("currentPrice", al.CurrentPrice),
	// 		zap.Float64("minIncrement", minIncrement),
	// 		zap.String("userID", userID.String()),
	// 	)
	// 	return nil, ErrBidIncrementTooSmall
	// }

	//time extension logic, if the bid occurs near to the end
	originalEndTime := al.EndTime
	now := time.Now()
	if time.Now().Add(al.TimeExtension).After(al.EndTime) {
		al.EndTime = time.Now().Add(al.TimeExtension)
		//a log entry musy be useful, consider it
		log.Info("Auction time extended",
			zap.String("lotID", al.ID.String()),
			zap.Time("originalEndTime", originalEndTime),
			zap.Time("newEndTime", al.EndTime),
			zap.Duration("extension", al.TimeExtension),
			zap.String("userID", userID.String()),
		)
	}

	//updates lot state
	al.CurrentPrice = amount
	al.LastBidTime = &now
	//cretes new bid
	newBid := NewBid(uuid.New(), al.ID, userID, amount, now)
	// adds the bid to the list, remember this is a simplyfied way to do it
	al.Bids = append(al.Bids, newBid)

	log.Info("Bid placed successfully",
		zap.String("lotID", al.ID.String()),
		zap.String("bidID", newBid.ID.String()),
		zap.String("userID", userID.String()),
		zap.Float64("amount", amount),
		zap.Float64("newCurrentPrice", al.CurrentPrice),
		zap.Time("newEndTime", al.EndTime),
	)

	return newBid, nil

}

// Start initiate he auction if is pending
func (al *AuctionLot) Start() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.State != StatePending {
		log.Warn("Attempted to start lot that is not pending",
			zap.String("lotID", al.ID.String()),
			zap.String("state", string(al.State)),
		)
		return ErrLotAlreadyStartedOrFinished
	}
	al.State = StateActive
	log.Info("Auction lot started",
		zap.String("lotID", al.ID.String()),
		zap.Time("endTime", al.EndTime),
	)
	//maybe set EndTime here, if was not defined at lot creation
	return nil
}

// Finish ends an active lot
func (al *AuctionLot) Finish() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.State != StateActive {
		log.Warn("Attempted to finish lot that is not active",
			zap.String("lotID", al.ID.String()),
			zap.String("state", string(al.State)),
		)
		return ErrLotNotActive
	}
	al.State = StateFinished
	log.Info("Auction lot finished",
		zap.String("lotID", al.ID.String()),
		zap.Float64("finalPrice", al.CurrentPrice),
	)
	return nil
}

// Cancel auction
func (al *AuctionLot) Cancel() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.State == StateFinished || al.State == StateCancelled {
		log.Warn("Attempted to cancel lot that is already finished or cancelled",
			zap.String("lotID", al.ID.String()),
			zap.String("state", string(al.State)),
		)
		return ErrLotAlreadyFinishedOrCancelled
	}

	al.State = StateCancelled
	log.Info("Auction lot cancelled",
		zap.String("lotID", al.ID.String()),
		zap.String("state", string(al.State)),
	)
	return nil
}
