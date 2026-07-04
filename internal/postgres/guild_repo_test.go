package postgres

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Setup for each test - clean state
func setupTest(t *testing.T, ctx context.Context) {
	// Truncate tables before each test
	if _, err := testDB.ExecContext(ctx, "TRUNCATE guilds CASCADE"); err != nil {
		t.Fatal(err)
	}
}

func TestGuildRepository_GetAndUpdate(t *testing.T) {
	const (
		testInitialGuildID = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100
		testAddedGold      = 10
		testExpectedGold   = testInitialGold + testAddedGold
	)

	ctx := context.Background()
	setupTest(t, ctx)

	repo := NewRepository(testDB)

	// Insert
	_, err := testDB.ExecContext(
		ctx,
		`INSERT INTO guilds (id, gold, reserved) VALUES ($1, $2, $3)`,
		testInitialGuildID, testInitialGold, testInitialReserve,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Test GetItem
	g, err := repo.Get(ctx, testInitialGuildID)
	if err != nil {
		t.Fatal(err)
	}

	if g.Gold != testInitialGold || g.Reserved != testInitialReserve {
		t.Errorf("unexpected values %+v", g)
	}

	// Test Update
	g.Gold = testExpectedGold
	if err := repo.Update(ctx, g); err != nil {
		t.Fatal(err)
	}

	// Verify update
	updated, err := repo.Get(ctx, testInitialGuildID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Gold != testExpectedGold {
		t.Errorf("expected gold %v, got %v", testExpectedGold, updated.Gold)
	}
}
