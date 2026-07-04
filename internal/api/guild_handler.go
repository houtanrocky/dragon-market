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
func (h *guildHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
