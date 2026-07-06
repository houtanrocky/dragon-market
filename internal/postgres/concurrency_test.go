package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"

	"market-dragon/internal/auction"
	"market-dragon/internal/gold"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"market-dragon/internal/order"
)

func TestConcurrentBuy_OnlyOneBuyerGetsItem(t *testing.T) {
	ctx := context.Background()
	if err := TruncateTables(ctx, "wallet_transactions", "bids", "auctions", "limit_orders", "items", "guilds"); err != nil {
		t.Fatal(err)
	}
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO guilds (id, gold, reserved, daily_limit) VALUES
		('seller', 0, 0, 0), ('buyer-1', 1000, 0, 0), ('buyer-2', 1000, 0, 0);
		INSERT INTO items (id, name, type, owner_id, status, base_price)
		VALUES ('item-1', 'wand', 'rare', 'seller', 'listed_in_order', 100);
		INSERT INTO limit_orders (id, item_id, seller_id, price, status)
		VALUES ('order-1', 'item-1', 'seller', 100, 'listed');`)
	if err != nil {
		t.Fatal(err)
	}

	tx := NewTransactor(testDB)
	guildRepo := NewWalletRepository(testDB)
	itemRepo := NewItemRepository(testDB)
	svc := order.NewOrderService(NewOrderRepository(testDB), guild.NewWalletService(guildRepo, tx), item.NewItemService(itemRepo, guildRepo), tx)

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, buyer := range []string{"buyer-1", "buyer-2"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			errs <- svc.Buy(ctx, "order-1", id)
		}(buyer)
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if !errors.Is(err, order.ErrOrderAlreadySold) {
			t.Fatalf("unexpected buy error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful purchases=%d, want 1", successes)
	}
	var owner string
	if err := testDB.QueryRowContext(ctx, `SELECT owner_id FROM items WHERE id='item-1'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	var auditCount int
	if err := testDB.QueryRowContext(ctx, `SELECT count(*) FROM wallet_transactions`).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if (owner != "buyer-1" && owner != "buyer-2") || auditCount != 2 {
		t.Fatalf("owner=%s audit records=%d", owner, auditCount)
	}
}

func TestConcurrentBids_LeavesSingleHighestReservation(t *testing.T) {
	ctx := context.Background()
	if err := TruncateTables(ctx, "wallet_transactions", "bids", "auctions", "limit_orders", "items", "guilds"); err != nil {
		t.Fatal(err)
	}
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO guilds (id, gold, reserved, daily_limit) VALUES
		('seller', 0, 0, 0), ('bidder-1', 1000, 0, 0), ('bidder-2', 1000, 0, 0);
		INSERT INTO items (id, name, type, owner_id, status, base_price)
		VALUES ('legend-1', 'Soul Reaver', 'legendary', 'seller', 'listed_in_auction', 100);
		INSERT INTO auctions (id, item_id, seller_id, ends_at, status)
		VALUES ('auction-1', 'legend-1', 'seller', NOW() + INTERVAL '1 hour', 'active');`)
	if err != nil {
		t.Fatal(err)
	}
	tx := NewTransactor(testDB)
	guildRepo := NewWalletRepository(testDB)
	svc := auction.NewAuctionService(NewAuctionRepository(testDB), guild.NewWalletService(guildRepo, tx), item.NewItemService(NewItemRepository(testDB), guildRepo), tx)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, input := range []struct {
		bidder string
		amount gold.Amount
	}{{"bidder-1", 100}, {"bidder-2", 110}} {
		wg.Add(1)
		go func(bidder string, amount gold.Amount) {
			defer wg.Done()
			<-start
			_, err := svc.PlaceBid(ctx, "auction-1", bidder, amount)
			if err != nil && !errors.Is(err, auction.ErrBidTooLow) {
				t.Errorf("PlaceBid: %v", err)
			}
		}(input.bidder, input.amount)
	}
	close(start)
	wg.Wait()

	var active, amount, totalReserved int64
	if err := testDB.QueryRowContext(ctx, `SELECT count(*), max(amount) FROM bids WHERE status='active'`).Scan(&active, &amount); err != nil {
		t.Fatal(err)
	}
	if err := testDB.QueryRowContext(ctx, `SELECT sum(reserved) FROM guilds`).Scan(&totalReserved); err != nil {
		t.Fatal(err)
	}
	if active != 1 || amount != 110 || totalReserved != 110 {
		t.Fatalf("active=%d amount=%d reserved=%d", active, amount, totalReserved)
	}
}

func TestConcurrentAuctionStart_OnlyOneSucceeds(t *testing.T) {
	ctx := context.Background()
	if err := TruncateTables(ctx, "wallet_transactions", "bids", "auctions", "limit_orders", "items", "guilds"); err != nil {
		t.Fatal(err)
	}
	_, err := testDB.ExecContext(ctx, `
		INSERT INTO guilds (id, gold, reserved) VALUES ('seller', 0, 0);
		INSERT INTO items (id, name, type, owner_id, status, base_price)
		VALUES ('legend-1', 'Soul Reaver', 'legendary', 'seller', 'free', 100);`)
	if err != nil {
		t.Fatal(err)
	}
	tx := NewTransactor(testDB)
	guildRepo := NewWalletRepository(testDB)
	svc := auction.NewAuctionService(NewAuctionRepository(testDB), guild.NewWalletService(guildRepo, tx), item.NewItemService(NewItemRepository(testDB), guildRepo), tx)

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := svc.StartAuction(ctx, "legend-1", "seller")
			results <- err
		}()
	}
	close(start)
	successes := 0
	for range 2 {
		if err := <-results; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful auction starts=%d, want 1", successes)
	}
	var count int
	if err := testDB.QueryRowContext(ctx, `SELECT count(*) FROM auctions WHERE status='active'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("active auctions=%d", count)
	}
}
