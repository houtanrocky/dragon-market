package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"market-dragon/internal/auction"
)

func TestAuctionRepository_Lifecycle(t *testing.T) {
	ctx := context.Background()
	if err := TruncateTables(ctx, "bids", "auctions", "items", "guilds"); err != nil {
		t.Fatal(err)
	}
	if _, err := testDB.ExecContext(ctx, `INSERT INTO guilds (id, gold, reserved) VALUES
		('seller', 1000, 0), ('bidder', 1000, 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := testDB.ExecContext(ctx, `INSERT INTO items
		(id, name, type, owner_id, status, base_price)
		VALUES ('item-1', 'Soul Reaver', 'legendary', 'seller', 'listed_in_auction', 100)`); err != nil {
		t.Fatal(err)
	}

	repo := NewAuctionRepository(testDB)
	a := &auction.Auction{
		ID: "auction-1", ItemID: "item-1", SellerID: "seller",
		EndsAt: time.Now().Add(time.Hour), Status: auction.ActiveAuction,
	}
	if err := repo.CreateAuction(ctx, a); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetActiveAuctionByItemID(ctx, a.ItemID); err != nil {
		t.Fatal(err)
	}

	b := &auction.Bid{
		ID: "bid-1", AuctionID: a.ID, BidderID: "bidder",
		Amount: 105, PlacedAt: time.Now(), Status: auction.ActiveBid,
	}
	if _, err := repo.CreateBid(ctx, b); err != nil {
		t.Fatal(err)
	}
	if err := repo.CancelActiveBid(ctx, a.ID, b.ID, b.BidderID); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetBidByID(ctx, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != auction.CancelledBid {
		t.Fatalf("bid status = %q, want %q", got.Status, auction.CancelledBid)
	}

	if _, err := repo.GetTopActiveBid(ctx, a.ID); !errors.Is(err, auction.ErrBidNotFound) {
		t.Fatalf("top bid error = %v, want ErrBidNotFound", err)
	}
}
