package api

import (
	"context"
	"market-dragon/internal/gold"
	"net/http"

	"market-dragon/internal/auction"
)

type AuctionService interface {
	StartAuction(ctx context.Context, itemID, sellerID string) (*auction.Auction, error)
	PlaceBid(ctx context.Context, auctionID, bidderID string, amount gold.Amount) error
	CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error
	EndAuction(ctx context.Context, auctionID string) error
}

type auctionHandler struct {
	svc AuctionService
}

// Start POST /auctions
func (h *auctionHandler) Start(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// PlaceBid POST /auctions/{id}/bids
func (h *auctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (h *auctionHandler) GetBid(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// CancelBid DELETE /auctions/{auctionID}/bids/{bidID}
func (h *auctionHandler) CancelBid(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
