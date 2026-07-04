package api

import (
	"context"
	"market-dragon/internal/gold"
	"net/http"

	"market-dragon/internal/item"
)

type ItemService interface {
	Create(ctx context.Context, name string, typ item.Type, ownerID string, basePrice gold.Amount) (*item.Item, error)
	ListFree(ctx context.Context, itemID string) ([]*item.Item, error)
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
}

type itemHandler struct {
	svc ItemService
}

// GET /items
func (h *itemHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// GET /items/{id}
func (h *itemHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// POST /items
func (h *itemHandler) Post(w http.ResponseWriter, r *http.Request) { panic("implement me") }
