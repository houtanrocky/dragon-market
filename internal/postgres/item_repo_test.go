package postgres

import (
	"context"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"testing"
)

func TestItemRepository_GetByID(t *testing.T) {
	// ----------- Arrange ------------
	const (
		testInitialGuildID = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100
		testAddedGold      = 10
		testExpectedGold   = testInitialGold + testAddedGold
	)
	const (
		testItemInitialID        = "item-1"
		testItemInitialName      = "Sassy Sword"
		testItemInitialType      = item.Common
		testItemInitialOwner     = testInitialGuildID
		testItemInitialAvailable = true
		testItemInitialBasePrice = 1000
	)
	ctx := context.Background()
	if err := TruncateTables(ctx, "items"); err != nil {
		t.Fatal(err)
	}

	repo := NewItemRepository(testDB)
	_, err := testDB.ExecContext(
		ctx,
		`INSERT INTO guilds (id, gold, reserved) VALUES ($1, $2, $3)`,
		testInitialGuildID, testInitialGold, testInitialReserve,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = testDB.ExecContext(ctx,
		`INSERT INTO items (id, name, type, owner_id, available, base_price)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		testItemInitialID, testItemInitialName, testItemInitialType, testItemInitialOwner,
		testItemInitialAvailable, testItemInitialBasePrice)
	if err != nil {
		t.Fatal(err)
	}

	// ------- Act ----------------
	it, err := repo.GetByID(ctx, testItemInitialID)
	if err != nil {
		t.Fatal(err)
	}
	// ------- Assert -------------------
	if it.ID != testItemInitialID || it.Name != testItemInitialName {
		t.Errorf("unexpected item: %v", it)
	}
}

func TestItemRepository_ListAvailable(t *testing.T) {
	// ----- Arrange --------------
	testGuilds := map[string]*guild.Guild{
		"guild-1": {
			ID: "guild-1",
		},
		"guild-2": {
			ID: "guild-2",
		},
	}
	testItems := map[string]*item.Item{
		"item-1": {
			ID:        "item-1",
			Name:      "Sassy Sword",
			Type:      item.Common,
			OwnerID:   "guild-1",
			Available: true,
			BasePrice: 100,
		},
		"item-2": {
			ID:        "item-2",
			Name:      "Brown Sword",
			Type:      item.Common,
			OwnerID:   "guild-2",
			Available: true,
			BasePrice: 200,
		},
	}
	ctx := context.Background()
	if err := TruncateTables(ctx, "items"); err != nil {
		t.Fatal(err)
	}

	repo := NewItemRepository(testDB)

	for _, g := range testGuilds {
		_, err := testDB.ExecContext(
			ctx,
			`INSERT INTO guilds (id, gold, reserved) VALUES ($1, $2, $3)`,
			g.ID, g.Gold, g.Reserved,
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, i := range testItems {
		_, err := testDB.ExecContext(ctx,
			`INSERT INTO items (id, name, type, owner_id, available, base_price)
		VALUES ($1, $2, $3, $4, $5, $6)`,
			i.ID, i.Name, i.Type, i.OwnerID,
			i.Available, i.BasePrice)
		if err != nil {
			t.Fatal(err)
		}
	}

	// ------ Act ---------
	items, err := repo.ListAvailable(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// ------ Assert ------
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	itemMap := make(map[string]*item.Item)
	for _, it := range items {
		itemMap[it.ID] = it
	}

	if _, ok := itemMap["item-1"]; !ok {
		t.Errorf("expected item-1 in results")
	}
	if _, ok := itemMap["item-2"]; !ok {
		t.Errorf("expected item-2 in results")
	}
}

func TestItemRepository_Update(t *testing.T) {
	// ---------- Arrange -----------------------

}
