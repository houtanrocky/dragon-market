package auction

import (
	"context"
	"fmt"
	"market-dragon/internal/guild"
	"testing"
)

type MockAuctionRepo struct {
	auctions map[string]*Auction
	bids     map[string]*Bid
}

func (r MockAuctionRepo) Create(ctx context.Context, a *Auction) error {
	r.auctions[a.ItemID] = a
	return nil
}

func (r MockAuctionRepo) GetByID(ctx context.Context, id string) (*Auction, error) {
	a, ok := r.auctions[id]
	if !ok {
		return nil, fmt.Errorf("auction with id: %v not found", id)
	}
	return a, nil
}

func (r MockAuctionRepo) GetActiveByItemID(ctx context.Context, id string) (*Auction, error) {
	a, ok := r.auctions[id]
	if !ok {
		return nil, fmt.Errorf("auction with id %v not found", id)
	}
	if a.Status != Active {
		return nil, fmt.Errorf("auction with id %v is not %v", id, Active)
	}

	return a, nil
}

func (r MockAuctionRepo) Update(ctx context.Context, a *Auction) error {
	_, ok := r.auctions[a.ID]
	if !ok {
		return fmt.Errorf("there's no user with id %v", a.ID)
	}
	r.auctions[a.ID] = a
	return nil
}

func (r MockAuctionRepo) PlaceBid(ctx context.Context, b *Bid) error {
	_, ok := r.auctions[b.AuctionID]
	if !ok {
		return fmt.Errorf("auction with id: %v, not found for bid: %v", b.AuctionID, b.ID)
	}

	r.bids[b.ID] = b
	return nil
}

func (r MockAuctionRepo) GetTopBid(ctx context.Context, auctionID string) (*Bid, error) {
	var topB *Bid
	var topBidAmount float64 = -1
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID && bid.Amount > topBidAmount {
			topBidAmount = bid.Amount
			topB = bid
		}
	}
	if topBidAmount == -1 {
		return nil, fmt.Errorf("no bid found")
	}
	return topB, nil
}

func (r MockAuctionRepo) GetBidByAuction(ctx context.Context, auctionID string) ([]*Bid, error) {
	var bids []*Bid
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			bids = append(bids, bid)
		}
	}
	return bids, nil
}

type MockWalletService struct {
}

func NewMockWalletService() *MockWalletService {
	guild.WalletService()
}

func TestService_StartAuction(t *testing.T) {
	ctx := context.Background()
	r := &MockAuctionRepo{}
	wSvc := NewWalletService()
	aSvc := NewAuctionService(r, wSvc)
}

func TestService_PlaceBid(t *testing.T) {

}

func TestService_CancelBid(t *testing.T) {

}

func TestService_EndAuction(t *testing.T) {

}
