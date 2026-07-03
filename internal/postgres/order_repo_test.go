package postgres

import (
	"context"
	"market-dragon/internal/order"
	"testing"
	"time"
)

func TestOrderRepository_CreateGetUpdate(t *testing.T) {
	ctx := context.Background()
	if err := TruncateTables(ctx, "limit_orders", "items", "guilds"); err != nil {
		t.Fatal(err)
	}

	// insert guilds (seller/buyer), and an item
	sellerID := "guild-1"
	buyerID := "guild-2"
	itemID := "item-1"

	// insert guilds
	_, err := testDB.ExecContext(ctx, `INSERT INTO guilds (id, gold, reserved) VALUES ($1, 100, 0), ($2, 200, 0)`, sellerID, buyerID)
	if err != nil {
		t.Fatal(err)
	}

	// Insert item (status = 'free')
	_, err = testDB.ExecContext(ctx, `INSERT INTO items
    (id, name, type, owner_id, status, base_price) 
	VALUES ($1, 'Sword', 'common', $2, 'free', 100)`, itemID, sellerID)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewOrderRepository(testDB)

	// Test create
	now := time.Now().UTC()
	o := order.LimitOrder{
		ID:       "order-1",
		ItemID:   itemID,
		SellerID: sellerID,
		Price:    200.0,
		Status:   order.Listed,
		ListedAt: now,
	}
	if err := repo.Create(ctx, &o); err != nil {
		t.Fatal(err)
	}

	// Test GetByID
	got, err := repo.GetByID(ctx, "order-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != o.ID || got.ItemID != o.ItemID || got.SellerID != o.SellerID || got.Price != o.Price || got.Status != o.Status {
		t.Errorf("GetByID returned unexpected order: %+v", got)
	}
	if got.BuyerID != nil {
		t.Error("expected BuyerID to be nil, got", *got.BuyerID)
	}
	// compare times ( tolerate 2 secs)
	if got.ListedAt.Sub(now) > 2*time.Second {
		t.Errorf("listed_at mismatch: got %v, expected ~%v", got.ListedAt, now)
	}

	// Test Update set buyer and status to sold
	buyer := buyerID
	o.BuyerID = &buyer
	o.Status = order.Sold
	if err := repo.Update(ctx, &o); err != nil {
		t.Fatal(err)
	}
	updated, err := repo.GetByID(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.BuyerID == nil || *updated.BuyerID != buyerID {
		t.Errorf("buyer ID not updated correctly: got %v, expected %v", updated.BuyerID, buyerID)
	}
	if updated.Status != order.Sold {
		t.Errorf("status not updated to sold: got %v", updated.Status)
	}
}
