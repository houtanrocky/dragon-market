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

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

type fakeTx struct{}

func (fakeTx) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeAuctionRepository struct {
	auctions map[string]*Auction
	bids     map[string]*Bid
}

func newFakeAuctionRepository() *fakeAuctionRepository {
	return &fakeAuctionRepository{auctions: map[string]*Auction{}, bids: map[string]*Bid{}}
}

func (r *fakeAuctionRepository) CreateAuction(_ context.Context, a *Auction) error {
	r.auctions[a.ID] = a
	return nil
}

func (r *fakeAuctionRepository) GetAuctionByID(_ context.Context, id string) (*Auction, error) {
	a := r.auctions[id]
	if a == nil {
		return nil, ErrAuctionNotFound
	}
	return a, nil
}

func (r *fakeAuctionRepository) GetActiveAuctionByItemID(_ context.Context, id string) (*Auction, error) {
	for _, a := range r.auctions {
		if a.ItemID == id && a.Status == ActiveAuction {
			return a, nil
		}
	}
	return nil, ErrAuctionNotFound
}

func (r *fakeAuctionRepository) ExtendActiveAuction(_ context.Context, id string, endsAt time.Time) error {
	a, err := r.GetAuctionByID(context.Background(), id)
	if err != nil || a.Status != ActiveAuction {
		return ErrAuctionNotActive
	}
	a.EndsAt = endsAt
	return nil
}

func (r *fakeAuctionRepository) EndActiveAuction(_ context.Context, id string) error {
	a, err := r.GetAuctionByID(context.Background(), id)
	if err != nil || a.Status != ActiveAuction {
		return ErrAuctionNotActive
	}
	a.Status = EndedAuction
	return nil
}

func (r *fakeAuctionRepository) CreateBid(_ context.Context, b *Bid) error {
	r.bids[b.ID] = b
	return nil
}

func (r *fakeAuctionRepository) GetBidByID(_ context.Context, id string) (*Bid, error) {
	b := r.bids[id]
	if b == nil {
		return nil, ErrBidNotFound
	}
	return b, nil
}

func (r *fakeAuctionRepository) GetTopActiveBid(_ context.Context, auctionID string) (*Bid, error) {
	for _, b := range r.bids {
		if b.AuctionID == auctionID && b.Status == ActiveBid {
			return b, nil
		}
	}
	return nil, ErrBidNotFound
}

func (r *fakeAuctionRepository) MarkBidOutbid(_ context.Context, id string) error {
	b, err := r.GetBidByID(context.Background(), id)
	if err != nil || b.Status != ActiveBid {
		return ErrBidNotFound
	}
	b.Status = OutbidBid
	return nil
}

func (r *fakeAuctionRepository) CancelOutbidBid(_ context.Context, auctionID, bidID, bidderID string) error {
	b, err := r.GetBidByID(context.Background(), bidID)
	if err != nil || b.AuctionID != auctionID || b.BidderID != bidderID || b.Status != OutbidBid {
		return ErrBidNotCancellable
	}
	b.Status = CancelledBid
	return nil
}

func (r *fakeAuctionRepository) MarkBidWinning(_ context.Context, id string) error {
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

type fakeWallet struct {
	reserved []walletCall
	released []walletCall
	deducted []walletCall
	earned   []walletCall
}

func (w *fakeWallet) Reserve(_ context.Context, id string, amount gold.Amount) error {
	w.reserved = append(w.reserved, walletCall{id, amount})
	return nil
}
func (w *fakeWallet) Release(_ context.Context, id string, amount gold.Amount) error {
	w.released = append(w.released, walletCall{id, amount})
	return nil
}
func (w *fakeWallet) Deduct(_ context.Context, id string, amount gold.Amount) error {
	w.deducted = append(w.deducted, walletCall{id, amount})
	return nil
}
func (w *fakeWallet) Earn(_ context.Context, id string, amount gold.Amount) error {
	w.earned = append(w.earned, walletCall{id, amount})
	return nil
}

type fakeItems struct{ items map[string]*item.Item }

func newFakeItems() *fakeItems {
	return &fakeItems{items: map[string]*item.Item{itemID: {
		ID: itemID, Type: item.Legendary, OwnerID: sellerID, Status: item.Free,
	}}}
}

func (s *fakeItems) Get(_ context.Context, id string) (*item.Item, error) {
	it := s.items[id]
	if it == nil {
		return nil, item.ErrItemNotFound
	}
	return it, nil
}

func (s *fakeItems) MarkListedInAuction(_ context.Context, id, owner string) error {
	it := s.items[id]
	if it == nil || it.OwnerID != owner || it.Status != item.Free || it.Type != item.Legendary {
		return ErrItemNotAvailable
	}
	it.Status = item.ListedInAuction
	return nil
}

func (s *fakeItems) TransferFromAuction(_ context.Context, id, owner, winner string) error {
	it := s.items[id]
	if it == nil || it.OwnerID != owner || it.Status != item.ListedInAuction {
		return ErrItemNotAvailable
	}
	it.OwnerID, it.Status = winner, item.Free
	return nil
}

func (s *fakeItems) ReleaseFromAuction(_ context.Context, id string) error {
	it := s.items[id]
	if it == nil || it.Status != item.ListedInAuction {
		return ErrItemNotAvailable
	}
	it.Status = item.Free
	return nil
}

func newService(repo *fakeAuctionRepository, wallet *fakeWallet, items *fakeItems, now time.Time) *AuctionServiceImpl {
	return NewAuctionService(repo, wallet, items, fakeTx{}, WithClock(fakeClock{now: now}))
}

func activeAuction(endsAt time.Time) *Auction {
	return &Auction{ID: auctionID, ItemID: itemID, SellerID: sellerID, EndsAt: endsAt, Status: ActiveAuction}
}

func TestService_StartAuction(t *testing.T) {
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
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
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(4 * time.Minute))
	repo.bids["old"] = &Bid{ID: "old", AuctionID: auctionID, BidderID: bidder1, Amount: 100, Status: ActiveBid}
	svc := newService(repo, wallet, items, fixedNow)

	if err := svc.PlaceBid(context.Background(), auctionID, bidder2, 105); err != nil {
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
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
	repo.auctions[auctionID] = activeAuction(fixedNow.Add(time.Hour))
	repo.bids["old"] = &Bid{ID: "old", AuctionID: auctionID, BidderID: bidder1, Amount: 100, Status: ActiveBid}
	svc := newService(repo, wallet, items, fixedNow)

	err := svc.PlaceBid(context.Background(), auctionID, bidder2, 104)
	if !errors.Is(err, ErrBidTooLow) {
		t.Fatalf("PlaceBid() error = %v, want ErrBidTooLow", err)
	}
}

func TestService_CancelBid_CancelsOutbidWithoutReleasingAgain(t *testing.T) {
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
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
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
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
	repo, wallet, items := newFakeAuctionRepository(), &fakeWallet{}, newFakeItems()
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
