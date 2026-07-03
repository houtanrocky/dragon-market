package api

import (
	"net/http"

	"market-dragon/internal/auction"
)

type auctionHandler struct {
	svc *auction.AuctionServiceImpl
}

// Start POST /auctions
func (h *auctionHandler) Start(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// PlaceBid POST /auctions/{id}/bids
func (h *auctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// CancelBid DELETE /auctions/{auctionID}/bids/{bidID}
func (h *auctionHandler) CancelBid(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
