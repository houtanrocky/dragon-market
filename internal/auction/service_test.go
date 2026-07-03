package auction

import (
	"context"
	"fmt"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"testing"
	"time"
)

type MockAuctionRepository struct {
	auctions map[string]*Auction
	bids     map[string]*Bid
}

func (r *MockAuctionRepository) Create(ctx context.Context, a *Auction) error {
	r.auctions[a.ID] = a
	return nil
}

func (r *MockAuctionRepository) GetByID(ctx context.Context, id string) (*Auction, error) {
	a, ok := r.auctions[id]
	if !ok {
		return nil, fmt.Errorf("auction with id: %v not found", id)
	}
	return a, nil
}

func (r *MockAuctionRepository) GetActiveByItemID(ctx context.Context, id string) (*Auction, error) {
	var auction *Auction
	for _, a := range r.auctions {
		if a.ItemID == id {
			auction = a
		}
	}
	if auction == nil {
		return nil, fmt.Errorf("auction not found")
	}
	if auction.Status != ActiveAuction {
		return nil, fmt.Errorf("auction is not %v", ActiveAuction)
	}

	return auction, nil
}

func (r *MockAuctionRepository) Update(ctx context.Context, a *Auction) error {
	_, ok := r.auctions[a.ID]
	if !ok {
		return fmt.Errorf("there's no user with id %v", a.ID)
	}
	r.auctions[a.ID] = a
	return nil
}

func (r *MockAuctionRepository) PlaceBid(ctx context.Context, b *Bid) error {
	_, ok := r.auctions[b.AuctionID]
	if !ok {
		return fmt.Errorf("auction with id: %v, not found for bid: %v", b.AuctionID, b.ID)
	}

	r.bids[b.ID] = b
	return nil
}

func (r *MockAuctionRepository) GetTopBid(ctx context.Context, auctionID string) (*Bid, error) {
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

func (r *MockAuctionRepository) GetBidsByAuction(ctx context.Context, auctionID string) ([]*Bid, error) {
	var bids []*Bid
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			bids = append(bids, bid)
		}
	}
	return bids, nil
}

func (r *MockAuctionRepository) CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error {
	b, ok := r.bids[bidID]
	if !ok {
		return fmt.Errorf("bid with id %v doesn't exist", bidID)
	}
	if b.AuctionID != auctionID || b.BidderID != bidID || b.BidderID != bidderID || b.Status != ActiveBid {
		return fmt.Errorf("unexpected bid: %v", b)
	}
	b.Status = CancelledBid
	r.bids[bidID] = b
	return nil
}

// --- Mock Wallet Service
type MockWalletService struct {
	guilds map[string]*guild.Guild
}

func (s *MockWalletService) Spend(ctx context.Context, id string, amount float64) error {
	g, ok := s.guilds[id]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	if g.Gold < amount {
		return fmt.Errorf("insufficient balance")
	}
	if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
		return fmt.Errorf("DailyLimit reached")
	}
	g.Gold -= amount
	return nil
}

func (s *MockWalletService) Earn(ctx context.Context, id string, amount float64) error {
	g, ok := s.guilds[id]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	g.Gold += amount
	return nil
}

func NewMockWalletService() *MockWalletService {
	return &MockWalletService{}
}

func (s *MockWalletService) Reserve(ctx context.Context, id string, amount float64) error {
	return nil
}
func (s *MockWalletService) Release(ctx context.Context, id string, amount float64) error {
	return nil
}
func (s *MockWalletService) Deduct(ctx context.Context, id string, amount float64) error {
	return nil
}

type MockItemRepository struct {
	items map[string]*item.Item
}

func (r *MockItemRepository) GetByID(ctx context.Context, id string) (*item.Item, error) {
	i, ok := r.items[id]
	if !ok {
		return nil, item.ErrItemNotFound
	}
	return i, nil
}

func (r *MockItemRepository) Update(ctx context.Context, i *item.Item) error {
	r.items[i.ID] = i
	return nil
}

func (r *MockItemRepository) ListFree(ctx context.Context) ([]*item.Item, error) {
	var result []*item.Item
	for _, it := range r.items {
		if it.Status == item.Free {
			result = append(result, it)
		}
	}
	return result, nil
}

const (
	AuctionID  = "auction-1"
	AuctionID2 = "auction-2"
	ItemID     = "item-1"
	ItemID2    = "item-2"
	ItemName   = "Legendary Sword"
	WalletID   = "guid-1"
	WalletID2  = "guild-2"

	BidID   = "bid-1"
	BidID2  = "bid-2"
	Amount  = 100
	Amount2 = 105
)

func defaultAuction() map[string]*Auction {
	return map[string]*Auction{
		AuctionID: {
			ID:       AuctionID,
			ItemID:   ItemID,
			SellerID: WalletID,
			EndsAt:   time.Now().Add(24 * time.Hour),
			Status:   ActiveAuction,
		},
	}
}

func defaultBid() map[string]*Bid {
	return map[string]*Bid{
		BidID: {
			ID:        BidID,
			AuctionID: AuctionID,
			BidderID:  WalletID2,
			Amount:    Amount,
			PlacedAt:  time.Now(),
		},
		BidID2: {
			ID:        BidID2,
			AuctionID: AuctionID,
			BidderID:  WalletID2,
			Amount:    Amount2,
			PlacedAt:  time.Now(),
		},
	}
}

func defaultItem() map[string]*item.Item {
	return map[string]*item.Item{
		ItemID: {
			ID:        ItemID,
			Name:      ItemName,
			Type:      item.Legendary,
			OwnerID:   WalletID,
			Status:    item.Free,
			BasePrice: 100,
		},
	}
}

func TestService_StartAuction(t *testing.T) {
	ctx := context.Background()
	aR := &MockAuctionRepository{
		auctions: defaultAuction(),
		bids:     defaultBid(),
	}
	wSvc := NewMockWalletService()
	aSvc := NewAuctionService(aR, wSvc)

	a, err := aSvc.StartAuction(ctx, ItemID, WalletID)
	if err != nil {
		t.Fatal(err)
	}

	createdAuction, err := aSvc.repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}

	if createdAuction.Status != ActiveAuction {
		t.Errorf("expected status active, received: %v", createdAuction.Status)
	}
}

func TestService_PlaceBid(t *testing.T) {
	ctx := context.Background()
	aR := &MockAuctionRepository{
		auctions: defaultAuction(),
		bids:     defaultBid(),
	}
	wSvc := NewMockWalletService()
	aSvc := NewAuctionService(aR, wSvc)

	err := aSvc.PlaceBid(ctx, AuctionID, WalletID2, 200)
	if err != nil {
		t.Fatal(err)
	}

	var saved *Bid
	for _, bid := range aR.bids {
		if bid.BidderID == WalletID2 && bid.Amount == 200 {
			saved = bid
			break
		}
	}

	if saved == nil {
		t.Fatal("expected bid to be saved")
	}

	if saved.AuctionID != AuctionID {
		t.Errorf("AuctionID = %q, want %q", saved.AuctionID, AuctionID)
	}
}

func TestService_CancelBid(t *testing.T) {
	ctx := context.Background()
	aR := &MockAuctionRepository{
		auctions: defaultAuction(),
		bids:     defaultBid(),
	}
	wSvc := NewMockWalletService()
	aSvc := NewAuctionService(aR, wSvc)

	err := aSvc.CancelBid(ctx, AuctionID, BidID, WalletID2)
	if err != nil {
		t.Fatal(err)
	}

	bids, err := aSvc.repo.GetBidsByAuction(ctx, AuctionID)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, b := range bids {
		if b.ID == BidID {
			if b.Status != CancelledBid {
				t.Errorf("expected status Cancelled, got %v", b.Status)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("bid %s not found", BidID)
	}
}

func TestService_EndAuction(t *testing.T) {
	ctx := context.Background()
	aR := &MockAuctionRepository{
		auctions: defaultAuction(),
		bids:     defaultBid(),
	}
	wSvc := NewMockWalletService()
	aSvc := NewAuctionService(aR, wSvc)

	err := aSvc.EndAuction(ctx, AuctionID)
	if err != nil {
		t.Fatal(err)
	}

	a, err := aSvc.repo.GetByID(ctx, AuctionID)
	if err != nil {
		t.Fatal()
	}

	if a.Status != EndedAuction {
		t.Errorf("expected status Ended, got %v", a.Status)
	}
}
