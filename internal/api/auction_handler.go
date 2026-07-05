package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"market-dragon/internal/gold"
	"mime"
	"net/http"
	"time"

	"market-dragon/internal/auction"

	"github.com/go-chi/chi/v5"
)

// -- start auction req
type startAuctionRequest struct {
	ItemID   string `json:"item_id"`
	SellerID string `json:"seller_id"`
}

type auctionResponse struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	SellerID string                `json:"type"`
	Status   auction.AuctionStatus `json:"status"`
}

func newAuctionResponse(a *auction.Auction) auctionResponse {
	return auctionResponse{
		ID:       a.ID,
		Name:     a.ItemID,
		SellerID: a.SellerID,
		Status:   a.Status,
	}
}

// -- place bid req
type placeBidRequest struct {
	AuctionID string      `json:"auction_id"`
	BidderID  string      `json:"bidder_id"`
	Amount    gold.Amount `json:"amount"`
}
type placeBidResponse struct {
	ID        string            `json:"id"`
	AuctionID string            `json:"auction_id"`
	BidderID  string            `json:"bidder_id"`
	Amount    gold.Amount       `json:"amount"`
	PlacedAt  time.Time         `json:"placed_at"`
	Status    auction.BidStatus `json:"status"`
}

func newPlaceBidResponse(b *auction.Bid) placeBidResponse {
	return placeBidResponse{
		ID:        b.ID,
		AuctionID: b.AuctionID,
		BidderID:  b.BidderID,
		Amount:    b.Amount,
		PlacedAt:  b.PlacedAt,
		Status:    b.Status,
	}
}

func newGetBidResponse(b *auction.Bid) placeBidResponse {
	return placeBidResponse{
		ID:        b.ID,
		AuctionID: b.AuctionID,
		BidderID:  b.BidderID,
		Amount:    b.Amount,
		PlacedAt:  b.PlacedAt,
		Status:    b.Status,
	}
}

type AuctionService interface {
	StartAuction(ctx context.Context, itemID, sellerID string) (*auction.Auction, error)
	PlaceBid(ctx context.Context, auctionID, bidderID string, amount gold.Amount) (*auction.Bid, error)
	CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error
	EndAuction(ctx context.Context, auctionID string) error
	GetBid(ctx context.Context, id string) (*auction.Bid, error)
}

type auctionHandler struct {
	svc AuctionService
}

// Start POST /auctions
func (h *auctionHandler) Start(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req startAuctionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	created, err := h.svc.StartAuction(r.Context(), req.ItemID, req.SellerID)

	if errors.Is(err, auction.ErrItemNotAvailable) {
		http.Error(w, "item is not free to be listed in auction", http.StatusBadRequest)
		return
	}
	if errors.Is(err, auction.ErrItemNotOwnedBySeller) {
		http.Error(w, "item is not owned by the seller", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "failed to create auction", http.StatusInternalServerError)
		slog.Error("failed to create auction", "error", err)
		return
	}

	if created == nil {
		http.Error(w, "failed to create auction", http.StatusInternalServerError)
		slog.Error("create returned nil without error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newAuctionResponse(created)); err != nil {
		slog.Error("encode auction response", "error", err)
	}

}

// PlaceBid POST /auctions/{id}/bids
func (h *auctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req placeBidRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "request body must contain one JSON object", http.StatusBadRequest)
		return
	}

	bid, err := h.svc.PlaceBid(r.Context(), req.AuctionID, req.BidderID, req.Amount)

	if err != nil {
		http.Error(w, "failed to place bid", http.StatusInternalServerError)
		slog.Error("failed to place bid", "error", err)
		return
	}

	if bid == nil {
		http.Error(w, "failed to place bid", http.StatusInternalServerError)
		slog.Error("PlaceBid returned nil without error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newPlaceBidResponse(bid)); err != nil {
		slog.Error("encode auction response", "error", err)
	}
}

func (h *auctionHandler) GetBid(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "bid ID is required", http.StatusBadRequest)
		return
	}

	b, err := h.svc.GetBid(r.Context(), id)
	if errors.Is(err, auction.ErrBidNotFound) {
		http.Error(w, "bid not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("get bid", "bid_id", id, "error", err)
		http.Error(w, "failed to get bid", http.StatusInternalServerError)
		return
	}
	if b == nil {
		slog.Error("get bid returned nil", "bid_id", id)
		http.Error(w, "failed to get bid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(newGetBidResponse(b)); err != nil {
		slog.Error("encode bid response", "error", err)
	}
}

// CancelBid handles DELETE /auctions/{auctionID}/bids/{bidID}.
func (h *auctionHandler) CancelBid(w http.ResponseWriter, r *http.Request) {
	auctionID := chi.URLParam(r, "auctionID")
	if auctionID == "" {
		http.Error(w, "auction ID is required", http.StatusBadRequest)
		return
	}

	bidID := chi.URLParam(r, "bidID")
	if bidID == "" {
		http.Error(w, "bid ID is required", http.StatusBadRequest)
		return
	}

	bid, err := h.svc.GetBid(r.Context(), bidID)
	switch {
	case errors.Is(err, auction.ErrBidNotFound):
		http.Error(w, "bid not found", http.StatusNotFound)
		return
	case err != nil:
		slog.Error("get bid", "bid_id", bidID, "error", err)
		http.Error(w, "failed to get bid", http.StatusInternalServerError)
		return
	case bid == nil:
		slog.Error("get bid returned nil", "bid_id", bidID)
		http.Error(w, "failed to get bid", http.StatusInternalServerError)
		return
	case bid.AuctionID != auctionID:
		http.Error(w, "bid not found for this auction", http.StatusNotFound)
		return
	}

	err = h.svc.CancelBid(r.Context(), auctionID, bidID, bid.BidderID)
	switch {
	case errors.Is(err, auction.ErrBidNotFound):
		http.Error(w, "bid not found", http.StatusNotFound)
		return
	case errors.Is(err, auction.ErrAuctionNotFound):
		http.Error(w, "auction not found", http.StatusNotFound)
		return
	case errors.Is(err, auction.ErrBidNotCancellable),
		errors.Is(err, auction.ErrAuctionNotActive):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		slog.Error(
			"cancel bid",
			"auction_id", auctionID,
			"bid_id", bidID,
			"error", err,
		)
		http.Error(w, "failed to cancel bid", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
