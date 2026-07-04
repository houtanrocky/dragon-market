package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"market-dragon/internal/gold"
	"market-dragon/internal/item"
)

type ItemService interface {
	Create(ctx context.Context, name string, typ item.Type, ownerID string, basePrice gold.Amount) (*item.Item, error)
	ListFree(ctx context.Context) ([]*item.Item, error)
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
}

type createItemRequest struct {
	Name      string      `json:"name"`
	Type      item.Type   `json:"type"`
	OwnerID   string      `json:"owner_id"`
	BasePrice gold.Amount `json:"base_price"`
}

type itemResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      item.Type   `json:"type"`
	OwnerID   string      `json:"owner_id"`
	Status    item.Status `json:"status"`
	BasePrice gold.Amount `json:"base_price"`
}

func newItemResponse(it *item.Item) itemResponse {
	return itemResponse{
		ID:        it.ID,
		Name:      it.Name,
		Type:      it.Type,
		OwnerID:   it.OwnerID,
		Status:    it.Status,
		BasePrice: it.BasePrice,
	}
}

type itemHandler struct {
	svc ItemService
}

// Create POST /items
func (h *itemHandler) Create(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createItemRequest
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

	created, err := h.svc.Create(r.Context(), req.Name, req.Type, req.OwnerID, req.BasePrice)

	if errors.Is(err, item.ErrInvalidItem) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, item.ErrOwnerNotFound) {
		http.Error(w, "owner not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		slog.Error("failed to create item", "error", err)
		return
	}

	if created == nil {
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		slog.Error("create returned nil without error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newItemResponse(created)); err != nil {
		slog.Error("encode item response", "error", err)
	}

}

// GET /items
func (h *itemHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// GET /items/{id}
func (h *itemHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
