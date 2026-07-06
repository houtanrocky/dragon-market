package api

import (
	"context"
	"market-dragon/internal/auction"
	"market-dragon/internal/gold"
	"market-dragon/internal/item"
	"market-dragon/internal/order"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type routeAuctionService struct {
	value    *auction.Auction
	startErr error
}

func (s routeAuctionService) StartAuction(context.Context, string, string) (*auction.Auction, error) {
	return nil, s.startErr
}
func (s routeAuctionService) PlaceBid(context.Context, string, string, gold.Amount) (*auction.Bid, error) {
	return nil, nil
}
func (s routeAuctionService) CancelBid(context.Context, string, string, string) error { return nil }
func (s routeAuctionService) EndAuction(context.Context, string) error                { return nil }
func (s routeAuctionService) EndExpiredAuctions(context.Context, int) error           { return nil }
func (s routeAuctionService) GetBid(context.Context, string) (*auction.Bid, error)    { return nil, nil }
func (s routeAuctionService) GetAuction(_ context.Context, id string) (*auction.Auction, error) {
	if s.value != nil && s.value.ID == id {
		return s.value, nil
	}
	return nil, auction.ErrAuctionNotFound
}

type routeOrderService struct {
	canceledID string
	listErr    error
}

func (s *routeOrderService) List(context.Context, string, string, gold.Amount) (*order.LimitOrder, error) {
	return nil, s.listErr
}

func TestRouter_MissingReferencedItemReturns404(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		router http.Handler
	}{
		{
			name: "order", method: http.MethodPost, path: "/orders",
			body:   `{"item_id":"missing","owner_id":"seller","base_price":100}`,
			router: NewRouter(nil, nil, routeAuctionService{}, &routeOrderService{listErr: item.ErrItemNotFound}, nil),
		},
		{
			name: "auction", method: http.MethodPost, path: "/auctions",
			body:   `{"item_id":"missing","seller_id":"seller"}`,
			router: NewRouter(nil, nil, routeAuctionService{startErr: item.ErrItemNotFound}, nil, nil),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			tc.router.ServeHTTP(recorder, req)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
		})
	}
}
func (s *routeOrderService) Buy(context.Context, string, string) error { return nil }
func (s *routeOrderService) Cancel(_ context.Context, id, _ string) error {
	s.canceledID = id
	return nil
}

func TestRouter_HealthAndAuctionRoutes(t *testing.T) {
	a := &auction.Auction{ID: "auction-1", ItemID: "item-1", SellerID: "seller", EndsAt: time.Now(), Status: auction.ActiveAuction}
	router := NewRouter(nil, nil, routeAuctionService{value: a}, nil, nil)

	for _, tc := range []struct {
		path string
		code int
	}{{"/healthz", http.StatusOK}, {"/auctions/auction-1", http.StatusOK}, {"/auctions/missing", http.StatusNotFound}} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if recorder.Code != tc.code {
			t.Errorf("GET %s status=%d, want %d", tc.path, recorder.Code, tc.code)
		}
	}
}

func TestRouter_OrderCancelUsesPathID(t *testing.T) {
	orders := &routeOrderService{}
	router := NewRouter(nil, nil, routeAuctionService{}, orders, nil)
	req := httptest.NewRequest(http.MethodDelete, "/orders/path-order", strings.NewReader(`{"seller_id":"seller"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || orders.canceledID != "path-order" {
		t.Fatalf("status=%d canceled ID=%q", recorder.Code, orders.canceledID)
	}
}
