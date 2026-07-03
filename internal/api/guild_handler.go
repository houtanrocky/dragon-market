package api

import (
	"context"
	"net/http"
)

type GuildWalletService interface {
	Reserve(ctx context.Context, id string, amount float64) error
	Deduct(ctx context.Context, id string, amount float64) error
	Release(ctx context.Context, id string, amount float64) error
}

type guildHandler struct {
	svc GuildWalletService
}

// GET /guilds/{id}/wallet
func (h *guildHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := "..."
	amount := 100.0

	if err := h.svc.Reserve(ctx, id, amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
