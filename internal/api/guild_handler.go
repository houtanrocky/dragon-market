package api

import (
	"context"
	"market-dragon/internal/gold"
	"net/http"
)

type GuildWalletService interface {
	Reserve(ctx context.Context, id string, amount gold.Amount) error
	Deduct(ctx context.Context, id string, amount gold.Amount) error
	Release(ctx context.Context, id string, amount gold.Amount) error
}

type guildHandler struct {
	svc GuildWalletService
}

// GET /guilds/{id}/wallet
func (h *guildHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := "..."
	amount := 100

	if err := h.svc.Reserve(ctx, id, gold.Amount(amount)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
