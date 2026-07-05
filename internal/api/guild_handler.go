package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"market-dragon/internal/gold"
	"market-dragon/internal/guild"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type guildResponse struct {
	ID         string      `json:"id"`
	Gold       gold.Amount `json:"gold"`
	Reserved   gold.Amount `json:"reserved"`
	DailySpent gold.Amount `json:"daily_spent"`
	DailyLimit gold.Amount `json:"daily_limit"`
}

func newGuildResponse(g *guild.Guild) guildResponse {
	return guildResponse{
		ID:         g.ID,
		Gold:       g.Gold,
		Reserved:   g.Reserved,
		DailySpent: g.DailySpent,
		DailyLimit: g.DailyLimit,
	}
}

type GuildWalletService interface {
	GetGuild(ctx context.Context, id string) (*guild.Guild, error)
}

type guildHandler struct {
	svc GuildWalletService
}

// Get /guilds/{id}/wallet
func (h *guildHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "guild ID is required", http.StatusBadRequest)
		return
	}

	g, err := h.svc.GetGuild(r.Context(), id)
	if errors.Is(err, guild.ErrGuildNotFound) {
		http.Error(w, "guild not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("get guild", "guild_id", id, "error", err)
		http.Error(w, "failed to get guild", http.StatusInternalServerError)
		return
	}
	if g == nil {
		slog.Error("get guild returned nil", "guild_id", id)
		http.Error(w, "failed to get guild", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(newGuildResponse(g)); err != nil {
		slog.Error("encode guild response", "error", err)
	}
}
