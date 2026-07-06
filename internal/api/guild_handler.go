package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"market-dragon/internal/gold"
	"market-dragon/internal/guild"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type guildResponse struct {
	ID         string      `json:"id"`
	Gold       gold.Amount `json:"gold"`
	Reserved   gold.Amount `json:"reserved"`
	DailySpent gold.Amount `json:"daily_spent"`
	DailyLimit gold.Amount `json:"daily_limit"`
	Available  gold.Amount `json:"available"`
}

func newGuildResponse(g *guild.Guild) guildResponse {
	return guildResponse{
		ID:         g.ID,
		Gold:       g.Gold,
		Reserved:   g.Reserved,
		DailySpent: g.DailySpent,
		DailyLimit: g.DailyLimit,
		Available:  g.Gold - g.Reserved,
	}
}

type GuildWalletService interface {
	CreateGuild(ctx context.Context, id string, initialGold, dailyLimit gold.Amount) (*guild.Guild, error)
	GetGuild(ctx context.Context, id string) (*guild.Guild, error)
}

type createGuildRequest struct {
	ID         string      `json:"id"`
	Gold       gold.Amount `json:"gold"`
	DailyLimit gold.Amount `json:"daily_limit"`
}

func (h *guildHandler) Create(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createGuildRequest
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
	g, err := h.svc.CreateGuild(r.Context(), req.ID, req.Gold, req.DailyLimit)
	if writeDomainError(w, err) {
		return
	}
	if err != nil {
		slog.Error("create guild", "error", err)
		http.Error(w, "failed to create guild", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newGuildResponse(g)); err != nil {
		slog.Error("encode guild response", "error", err)
	}
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
