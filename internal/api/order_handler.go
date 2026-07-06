package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"market-dragon/internal/gold"
	"market-dragon/internal/idempotency"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"market-dragon/internal/order"
)

// --- create order request
type createOrderRequest struct {
	Name      string      `json:"name"`
	OwnerID   string      `json:"owner_id"`
	ItemID    string      `json:"item_id"`
	BasePrice gold.Amount `json:"base_price"`
}
type createOrderResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	OwnerID   string       `json:"owner_id"`
	Status    order.Status `json:"status"`
	BasePrice gold.Amount  `json:"base_price"`
}

// --- buy order request
type buyOrderRequest struct {
	BuyerID string `json:"buyer_id"`
}
type buyOrderResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	OwnerID   string       `json:"owner_id"`
	Status    order.Status `json:"status"`
	BasePrice gold.Amount  `json:"base_price"`
}

// --- cancel order request
type cancelOrderRequest struct {
	SellerID string `json:"seller_id"`
}

func newOrderResponse(o *order.LimitOrder) createOrderResponse {
	return createOrderResponse{
		ID: o.ID,
	}
}

type OrderService interface {
	List(ctx context.Context, itemID, sellerID string, price gold.Amount) (*order.LimitOrder, error)
	Buy(ctx context.Context, orderID, buyerID string) error
	Cancel(ctx context.Context, orderID, sellerID string) error
}

type orderHandler struct {
	svc     OrderService
	idemSvc *idempotency.IdempotencyService
}

// Create POST /orders
func (h *orderHandler) Create(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createOrderRequest
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

	created, err := h.svc.List(r.Context(), req.ItemID, req.OwnerID, req.BasePrice)

	if errors.Is(err, order.ErrInvalidOrder) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrOwnerNotFound) {
		http.Error(w, "owner not found", http.StatusNotFound)
		return
	}
	if writeDomainError(w, err) {
		return
	}

	if err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		slog.Error("failed to create order", "error", err)
		return
	}

	if created == nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		slog.Error("create returned nil without error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newOrderResponse(created)); err != nil {
		slog.Error("encode order response", "error", err)
	}
}

// Buy POST /orders/{id}/buy
func (h *orderHandler) Buy(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	var req buyOrderRequest
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

	orderID := chi.URLParam(r, "id")
	req.BuyerID = strings.TrimSpace(req.BuyerID)
	if orderID == "" || req.BuyerID == "" {
		http.Error(w, "order ID and buyer ID are required", http.StatusBadRequest)
		return
	}
	result, err := h.idemSvc.Execute(r.Context(), key, "order.buy",
		idempotency.Hash(orderID, req.BuyerID),
		func(ctx context.Context) (int, json.RawMessage, error) {
			return http.StatusNoContent, nil, h.svc.Buy(ctx, orderID, req.BuyerID)
		})
	if errors.Is(err, idempotency.ErrKeyConflict) || errors.Is(err, idempotency.ErrNotCompleted) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, order.ErrInvalidOrder) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrOrderAlreadySold) {
		http.Error(w, "order is already sold to someone else", http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrOrderNotListed) {
		http.Error(w, "order is already not listed for sale", http.StatusBadRequest)
		return
	}
	if writeDomainError(w, err) {
		return
	}
	if err != nil {
		http.Error(w, "failed to buy order", http.StatusInternalServerError)
		slog.Error("failed to buy order", "error", err)
		return
	}
	if result.Replayed {
		w.Header().Set("Idempotency-Replayed", "true")
	}
	w.WriteHeader(result.StatusCode)
}

// Cancel DELETE /orders/{id}
func (h *orderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req cancelOrderRequest
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

	orderID := chi.URLParam(r, "id")
	err = h.svc.Cancel(r.Context(), orderID, req.SellerID)
	if errors.Is(err, order.ErrInvalidOrder) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrOrderAlreadySold) {
		http.Error(w, "cannot cancel order that is already sold to someone else", http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrCancelOrderNotListed) {
		http.Error(w, "cannot cancel order that is not listed", http.StatusBadRequest)
		return
	}
	if errors.Is(err, order.ErrCancelOrderListedByAnother) {
		http.Error(w, "cannot cancel order that is listed by another user", http.StatusBadRequest)
		return
	}
	if writeDomainError(w, err) {
		return
	}
	if err != nil {
		http.Error(w, "failed to cancel order", http.StatusInternalServerError)
		slog.Error("failed to cancel order", "error", err)
		return
	}
}
