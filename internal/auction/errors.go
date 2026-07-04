package auction

import "errors"

var (
	ErrAuctionNotFound      = errors.New("auction not found")
	ErrAuctionNotActive     = errors.New("auction is not active")
	ErrAuctionNotFinished   = errors.New("auction has not reached its deadline")
	ErrAuctionFinished      = errors.New("auction deadline has passed")
	ErrActiveAuctionExists  = errors.New("active auction already exists for item")
	ErrBidNotFound          = errors.New("bid not found")
	ErrBidNotCancellable    = errors.New("bid cannot be cancelled")
	ErrBidTooLow            = errors.New("bid does not meet minimum increment")
	ErrInvalidBidAmount     = errors.New("bid amount must be positive")
	ErrSellerCannotBid      = errors.New("seller cannot bid on own item")
	ErrItemNotLegendary     = errors.New("only legendary items can be auctioned")
	ErrItemNotOwnedBySeller = errors.New("item is not owned by seller")
	ErrItemNotAvailable     = errors.New("item is not available")
)
