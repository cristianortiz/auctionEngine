package domain

import "errors"

var (
	ErrLotNotFound                   = errors.New("auction lo not found")
	ErrLotNotActive                  = errors.New("auction lot is not active")
	ErrBidAmountTooLow               = errors.New("bid amount is too low")
	ErrInvalidAmount                 = errors.New("bid amount cannot be zero o less than zero")
	ErrBidIncrementTooSmall          = errors.New("bid increment is too small") // if increment validations is implemented later
	ErrLotAlreadyStartedOrFinished   = errors.New("auction lot is already started or finished")
	ErrLotAlreadyFinishedOrCancelled = errors.New("auction lot is already finished or cancelled")
)
