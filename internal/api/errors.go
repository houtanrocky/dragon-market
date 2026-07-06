package api

import (
	"errors"
	"market-dragon/internal/auction"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"market-dragon/internal/order"
	"net/http"
)

// writeDomainError translates expected business outcomes into 4xx responses.
// It returns false for unexpected infrastructure or invariant failures.
func writeDomainError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, item.ErrItemNotFound),
		errors.Is(err, order.ErrOrderNotFound),
		errors.Is(err, auction.ErrAuctionNotFound),
		errors.Is(err, auction.ErrBidNotFound),
		errors.Is(err, guild.ErrGuildNotFound),
		errors.Is(err, item.ErrOwnerNotFound),
		errors.Is(err, order.ErrOwnerNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, item.ErrInvalidItem),
		errors.Is(err, order.ErrInvalidOrder),
		errors.Is(err, auction.ErrInvalidBidAmount),
		errors.Is(err, auction.ErrItemNotLegendary),
		errors.Is(err, guild.ErrInvalidAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, auction.ErrItemNotOwnedBySeller),
		errors.Is(err, auction.ErrSellerCannotBid),
		errors.Is(err, order.ErrOrderItemNotOwnedBySeller),
		errors.Is(err, order.ErrCannotBuyOwnOrder),
		errors.Is(err, order.ErrCancelOrderListedByAnother):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, auction.ErrItemNotAvailable),
		errors.Is(err, auction.ErrActiveAuctionExists),
		errors.Is(err, auction.ErrAuctionNotActive),
		errors.Is(err, auction.ErrAuctionFinished),
		errors.Is(err, auction.ErrAuctionNotFinished),
		errors.Is(err, auction.ErrBidNotCancellable),
		errors.Is(err, auction.ErrBidTooLow),
		errors.Is(err, order.ErrOrderAlreadySold),
		errors.Is(err, order.ErrOrderNotListed),
		errors.Is(err, order.ErrOrderItemNotAvailable),
		errors.Is(err, order.ErrCantListOrderListedInAuction),
		errors.Is(err, order.ErrCancelOrderNotListed),
		errors.Is(err, guild.ErrInsufficientBalance),
		errors.Is(err, guild.ErrInsufficientReserve),
		errors.Is(err, guild.ErrDailyLimitReached):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		return false
	}
	return true
}
