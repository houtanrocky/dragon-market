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

	"github.com/go-chi/chi/v5"
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
func newItemsResponse(items []*item.Item) []itemResponse {
	responses := make([]itemResponse, 0, len(items))

	for _, it := range items {
		responses = append(responses, newItemResponse(it))
	}

	return responses
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

// List GET /items
func (h *itemHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListFree(r.Context())
	if err != nil {
		slog.Error("list free items", "error", err)
		http.Error(w, "failed to list items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(newItemsResponse(items)); err != nil {
		slog.Error("encode item list response", "error", err)
	}
}

// Get /items/{id}
func (h *itemHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "item ID is required", http.StatusBadRequest)
		return
	}

	it, err := h.svc.GetItem(r.Context(), id)
	if errors.Is(err, item.ErrItemNotFound) {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("get item", "item_id", id, "error", err)
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}
	if it == nil {
		slog.Error("get item returned nil", "item_id", id)
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(newItemResponse(it)); err != nil {
		slog.Error("encode item response", "error", err)
	}
}
