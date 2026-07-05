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

	"market-dragon/internal/order"
)

type createOrderRequest struct {
	Name      string      `json:"name"`
	OwnerID   string      `json:"owner_id"`
	ItemID    string      `json:"item_id"`
	BasePrice gold.Amount `json:"base_price"`
}

type orderResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	OwnerID   string       `json:"owner_id"`
	Status    order.Status `json:"status"`
	BasePrice gold.Amount  `json:"base_price"`
}

func newOrderResponse(o *order.LimitOrder) orderResponse {
	return orderResponse{
		ID: o.ID,
	}
}

type OrderService interface {
	List(ctx context.Context, itemID, sellerID string, price gold.Amount) (*order.LimitOrder, error)
	Buy(ctx context.Context, orderID, buyerID string) error
	Cancel(ctx context.Context, orderID, sellerID string) error
}

type orderHandler struct {
	svc OrderService
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

// POST /orders/{id}/buy
func (h *orderHandler) Buy(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// DELETE /orders/{id}
func (h *orderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
