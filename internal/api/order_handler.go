package api

import (
	"net/http"

	"market-dragon/internal/order"
)

type orderHandler struct {
	svc *order.OrderServiceImpl
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
