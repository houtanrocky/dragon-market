package auction

import (
	"context"
	"errors"
	"testing"
	"time"

	"market-dragon/internal/gold"
	"market-dragon/internal/item"
)

const (
	auctionID = "auction-1"
	itemID    = "item-1"
	sellerID  = "seller"
	bidder1   = "bidder-1"
	bidder2   = "bidder-2"
)

var fixedNow = time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

type mockClock struct{ now time.Time }

func (c mockClock) Now() time.Time { return c.now }

type mockTx struct{}

func (mockTx) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockAuctionRepository struct {
	auctions map[string]*Auction
	bids     map[string]*Bid
}

func newMockAuctionRepository() *mockAuctionRepository {
	return &mockAuctionRepository{auctions: map[string]*Auction{}, bids: map[string]*Bid{}}
}

func (r *mockAuctionRepository) CreateAuction(_ context.Context, a *Auction) error {
	r.auctions[a.ID] = a
	return nil
}

func (r *mockAuctionRepository) GetAuctionByID(_ context.Context, id string) (*Auction, error) {
	a := r.auctions[id]
	if a == nil {
		return nil, ErrAuctionNotFound
	}
	return a, nil
}

func (r *mockAuctionRepository) GetActiveAuctionByItemID(_ context.Context, id string) (*Auction, error) {
	for _, a := range r.auctions {
		if a.ItemID == id && a.Status == ActiveAuction {
			return a, nil
		}
	}
	return nil, ErrAuctionNotFound
}

func (r *mockAuctionRepository) ExtendActiveAuction(_ context.Context, id string, endsAt time.Time) error {
	a, err := r.GetAuctionByID(context.Background(), id)
	if err != nil || a.Status != ActiveAuction {
		return ErrAuctionNotActive
	}
	a.EndsAt = endsAt
	return nil
}

func (r *mockAuctionRepository) EndActiveAuction(_ context.Context, id string) error {
	a, err := r.GetAuctionByID(context.Background(), id)
	if err != nil || a.Status != ActiveAuction {
		return ErrAuctionNotActive
	}
	a.Status = EndedAuction
	return nil
}

func (r *mockAuctionRepository) CreateBid(_ context.Context, b *Bid) (*Bid, error) {
	r.bids[b.ID] = b
	return b, nil
}

func (r *mockAuctionRepository) GetBidByID(_ context.Context, id string) (*Bid, error) {
	b := r.bids[id]
	if b == nil {
		return nil, ErrBidNotFound
	}
	return b, nil
}

func (r *mockAuctionRepository) GetTopActiveBid(_ context.Context, auctionID string) (*Bid, error) {
	for _, b := range r.bids {
		if b.AuctionID == auctionID && b.Status == ActiveBid {
			return b, nil
		}
	}
	return nil, ErrBidNotFound
}

func (r *mockAuctionRepository) MarkBidOutbid(_ context.Context, id string) error {
	b, err := r.GetBidByID(context.Background(), id)
	if err != nil || b.Status != ActiveBid {
		return ErrBidNotFound
	}
	b.Status = OutbidBid
	return nil
}

func (r *mockAuctionRepository) CancelOutbidBid(_ context.Context, auctionID, bidID, bidderID string) error {
	b, err := r.GetBidByID(context.Background(), bidID)
	if err != nil || b.AuctionID != auctionID || b.BidderID != bidderID || b.Status != OutbidBid {
		return ErrBidNotCancellable
	}
	b.Status = CancelledBid
	return nil
}

func (r *mockAuctionRepository) MarkBidWinning(_ context.Context, id string) error {
	b, err := r.GetBidByID(context.Background(), id)
	if err != nil || b.Status != ActiveBid {
		return ErrBidNotFound
	}
	b.Status = WinningBid
	return nil
}

type walletCall struct {
	guildID string
	amount  gold.Amount
}

type mockWallet struct {
	reserved []walletCall
	released []walletCall
	deducted []walletCall
	earned   []walletCall
}

func (w *mockWallet) Reserve(_ context.Context, id string, amount gold.Amount) error {
	w.reserved = append(w.reserved, walletCall{id, amount})
	return nil
}
func (w *mockWallet) Release(_ context.Context, id string, amount gold.Amount) error {
	w.released = append(w.released, walletCall{id, amount})
	return nil
}
func (w *mockWallet) Deduct(_ context.Context, id string, amount gold.Amount) error {
	w.deducted = append(w.deducted, walletCall{id, amount})
	return nil
}
func (w *mockWallet) Earn(_ context.Context, id string, amount gold.Amount) error {
	w.earned = append(w.earned, walletCall{id, amount})
	return nil
}

type mockItems struct{ items map[string]*item.Item }

func newMockItems() *mockItems {
	return &mockItems{items: map[string]*item.Item{itemID: {
		ID: itemID, Type: item.Legendary, OwnerID: sellerID, Status: item.Free,
	}}}
}

func (s *mockItems) GetItem(_ context.Context, id string) (*item.Item, error) {
	it := s.items[id]
	if it == nil {
		return nil, item.ErrItemNotFound
	}
	return it, nil
}

func (s *mockItems) MarkListedInAuction(_ context.Context, id, owner string) error {
	it := s.items[id]
	if it == nil || it.OwnerID != owner || it.Status != item.Free || it.Type != item.Legendary {
		return ErrItemNotAvailable
	}
	it.Status = item.ListedInAuction
	return nil
}

func (s *mockItems) TransferFromAuction(_ context.Context, id, owner, winner string) error {
	it := s.items[id]
	if it == nil || it.OwnerID != owner || it.Status != item.ListedInAuction {
		return ErrItemNotAvailable
	}
	it.OwnerID, it.Status = winner, item.Free
	return nil
}

func (s *mockItems) ReleaseFromAuction(_ context.Context, id string) error {
	it := s.items[id]
	if it == nil || it.Status != item.ListedInAuction {
		return ErrItemNotAvailable
	}
	it.Status = item.Free
	return nil
}

func newService(repo *mockAuctionRepository, wallet *mockWallet, items *mockItems, now time.Time) *AuctionServiceImpl {
	return NewAuctionService(repo, wallet, items, mockTx{}, WithClock(mockClock{now: now}))
}

func activeAuction(endsAt time.Time) *Auction {
	return &Auction{ID: auctionID, ItemID: itemID, SellerID: sellerID, EndsAt: endsAt, Status: ActiveAuction}
}

func TestService_StartAuction(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	svc := newService(repo, wallet, items, fixedNow)

	a, err := svc.StartAuction(context.Background(), itemID, sellerID)
	if err != nil {
		t.Fatalf("StartAuction() error = %v", err)
	}
	if a.EndsAt != fixedNow.Add(24*time.Hour) || items.items[itemID].Status != item.ListedInAuction {
		t.Fatal("auction deadline or item status was not initialized")
	}
}

func TestService_PlaceBid_ReplacesTopAndExtendsDeadline(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(4 * time.Minute))
	repo.bids["old"] = &Bid{ID: "old", AuctionID: auctionID, BidderID: bidder1, Amount: 100, Status: ActiveBid}
	svc := newService(repo, wallet, items, fixedNow)

	if _, err := svc.PlaceBid(context.Background(), auctionID, bidder2, 105); err != nil {
		t.Fatalf("PlaceBid() error = %v", err)
	}
	if repo.bids["old"].Status != OutbidBid || len(wallet.released) != 1 || len(wallet.reserved) != 1 {
		t.Fatal("new top bid did not reserve and release funds correctly")
	}
	if repo.auctions[auctionID].EndsAt != fixedNow.Add(9*time.Minute) {
		t.Fatalf("EndsAt = %v, want five-minute extension", repo.auctions[auctionID].EndsAt)
	}
}

func TestService_PlaceBid_RejectsLowIncrement(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(time.Hour))
	repo.bids["old"] = &Bid{ID: "old", AuctionID: auctionID, BidderID: bidder1, Amount: 100, Status: ActiveBid}
	svc := newService(repo, wallet, items, fixedNow)

	_, err := svc.PlaceBid(context.Background(), auctionID, bidder2, 104)
	if !errors.Is(err, ErrBidTooLow) {
		t.Fatalf("PlaceBid() error = %v, want ErrBidTooLow", err)
	}
}

func TestService_CancelBid_CancelsOutbidWithoutReleasingAgain(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(time.Hour))
	repo.bids["old"] = &Bid{ID: "old", AuctionID: auctionID, BidderID: bidder1, Amount: 100, Status: OutbidBid}
	svc := newService(repo, wallet, items, fixedNow)

	if err := svc.CancelBid(context.Background(), auctionID, "old", bidder1); err != nil {
		t.Fatalf("CancelBid() error = %v", err)
	}
	if repo.bids["old"].Status != CancelledBid || len(wallet.released) != 0 {
		t.Fatal("cancel did not preserve reservation invariant")
	}
}

func TestService_EndAuction_SettlesWinner(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	items.items[itemID].Status = item.ListedInAuction
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(-time.Second))
	repo.bids["top"] = &Bid{ID: "top", AuctionID: auctionID, BidderID: bidder1, Amount: 200, Status: ActiveBid}
	svc := newService(repo, wallet, items, fixedNow)

	if err := svc.EndAuction(context.Background(), auctionID); err != nil {
		t.Fatalf("EndAuction() error = %v", err)
	}
	if repo.auctions[auctionID].Status != EndedAuction || repo.bids["top"].Status != WinningBid {
		t.Fatal("auction or winning bid was not finalized")
	}
	if items.items[itemID].OwnerID != bidder1 || len(wallet.deducted) != 1 || len(wallet.earned) != 1 {
		t.Fatal("winner settlement was incomplete")
	}
}

func TestService_EndAuction_WithoutBidsReleasesItem(t *testing.T) {
	repo, wallet, items := newMockAuctionRepository(), &mockWallet{}, newMockItems()
	items.items[itemID].Status = item.ListedInAuction
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(-time.Second))
	svc := newService(repo, wallet, items, fixedNow)

	if err := svc.EndAuction(context.Background(), auctionID); err != nil {
		t.Fatalf("EndAuction() error = %v", err)
	}
	if items.items[itemID].Status != item.Free || repo.auctions[auctionID].Status != EndedAuction {
		t.Fatal("no-bid auction did not release item and end")
	}
}
