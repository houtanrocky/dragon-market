package api

import (
	"net/http"

	"market-dragon/internal/item"
)

type itemHandler struct {
	svc *item.ItemServiceImpl
}

// GET /items
func (h *itemHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// GET /items/{id}
func (h *itemHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
