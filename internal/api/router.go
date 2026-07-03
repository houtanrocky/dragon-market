package api

import (
	"net/http"

	"market-dragon/internal/auction"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"market-dragon/internal/order"
)

func NewRouter(
	guildSvc *guild.WalletServiceImpl,
	itemSvc *item.ItemServiceImpl,
	auctionSvc *auction.AuctionServiceImpl,
	orderSvc *order.OrderService,
) http.Handler {
	mux := http.NewServeMux()

	gh := &guildHandler{svc: guildSvc}
	ih := &itemHandler{svc: itemSvc}
	ah := &auctionHandler{svc: auctionSvc}
	oh := &orderHandler{svc: orderSvc}

	// Guild wallet
	mux.HandleFunc("GET /guilds/{id}/wallet", gh.Reserve)

	// Items
	mux.HandleFunc("GET /items", ih.List)
	mux.HandleFunc("GET /items/{id}", ih.Get)

	// Auctions (legendary only)
	mux.HandleFunc("POST /auctions", ah.Start)
	mux.HandleFunc("POST /auctions/{id}/bids", ah.PlaceBid)
	mux.HandleFunc("DELETE /auctions/{auctionID}/bids/{bidID}", ah.CancelBid)

	// Limit orders (common + rare)
	mux.HandleFunc("POST /orders", oh.Create)
	mux.HandleFunc("POST /orders/{id}/buy", oh.Buy)
	mux.HandleFunc("DELETE /orders/{id}", oh.Cancel)

	return mux
}
