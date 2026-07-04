package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	guildSvc GuildWalletService,
	itemSvc ItemService,
	auctionSvc AuctionService,
	orderSvc OrderService,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	gh := &guildHandler{svc: guildSvc}
	ih := &itemHandler{svc: itemSvc}
	ah := &auctionHandler{svc: auctionSvc}
	oh := &orderHandler{svc: orderSvc}

	r.Get("/guilds/{id}/wallet", gh.Get)

	r.Post("/items", ih.Create)
	r.Get("/items", ih.List)
	r.Get("/items/{id}", ih.Get)

	r.Post("/auctions", ah.Start)
	r.Get("/auctions/{id}", ah.GetBid)
	r.Post("/auctions/{id}/bids", ah.PlaceBid)
	r.Delete("/auctions/{auctionID}/bids/{bidID}", ah.CancelBid)

	r.Post("/orders", oh.Create)
	r.Post("/orders/{id}/buy", oh.Buy)
	r.Delete("/orders/{id}", oh.Cancel)

	return r
}
