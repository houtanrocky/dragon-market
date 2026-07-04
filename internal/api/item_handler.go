package api

import (
	"context"
	"net/http"

	"market-dragon/internal/item"
)

type ItemService interface {
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
	MarkListedInAuction(ctx context.Context, itemID, sellerID string) error
	TransferFromAuction(ctx context.Context, itemID, sellerID, winnerID string) error
	ReleaseFromAuction(ctx context.Context, itemID string) error
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
