package api

import (
	"context"
	"market-dragon/internal/gold"
	"net/http"

	"market-dragon/internal/order"
)

type OrderService interface {
	List(ctx context.Context, itemID, sellerID string, price gold.Amount) (*order.LimitOrder, error)
	Buy(ctx context.Context, orderID, buyerID string) error
	Cancel(ctx context.Context, orderID, sellerID string) error
}

type orderHandler struct {
	svc order.OrderService
}

// POST /orders
func (h *orderHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// POST /orders/{id}/buy
func (h *orderHandler) Buy(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// DELETE /orders/{id}
func (h *orderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
