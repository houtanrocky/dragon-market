package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"market-dragon/internal/idempotency"
)

func TestIdempotencyRepository_Claim_NewAndMatchingKey(t *testing.T) {
	ctx := context.Background()
	truncateIdempotencyRecords(t, ctx)
	repo := NewIdempotencyRepo(testDB)

	first, claimed, err := repo.Claim(ctx, "key-1", "auction.place_bid", "hash-1")
	if err != nil {
		t.Fatal(err)
	}
	if !claimed {
		t.Fatal("first claim reported claimed = false")
	}
	if first == nil {
		t.Fatal("first claim returned a nil record")
	}
	if first.Key != "key-1" || first.Operation != "auction.place_bid" || first.RequestHash != "hash-1" {
		t.Fatalf("unexpected first record: %+v", first)
	}
	if first.Completed() {
		t.Fatal("newly claimed record is completed")
	}

	second, claimed, err := repo.Claim(ctx, "key-1", "auction.place_bid", "hash-1")
	if err != nil {
		t.Fatal(err)
	}
	if claimed {
		t.Fatal("matching retry reported claimed = true")
	}
	if second == nil || second.Key != first.Key || second.RequestHash != first.RequestHash {
		t.Fatalf("unexpected retry record: %+v", second)
	}
}

func TestIdempotencyRepository_Claim_Conflict(t *testing.T) {
	ctx := context.Background()
	truncateIdempotencyRecords(t, ctx)
	repo := NewIdempotencyRepo(testDB)

	if _, _, err := repo.Claim(ctx, "key-2", "auction.place_bid", "hash-1"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		operation string
		hash      string
	}{
		{name: "different hash", operation: "auction.place_bid", hash: "hash-2"},
		{name: "different operation", operation: "order.buy", hash: "hash-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, claimed, err := repo.Claim(ctx, "key-2", tt.operation, tt.hash)
			if !errors.Is(err, idempotency.ErrKeyConflict) {
				t.Fatalf("Claim() error = %v, want ErrKeyConflict", err)
			}
			if claimed {
				t.Fatal("conflicting request reported claimed = true")
			}
		})
	}
}

func TestIdempotencyRepository_Complete_ReplaysOriginalResult(t *testing.T) {
	ctx := context.Background()
	truncateIdempotencyRecords(t, ctx)
	repo := NewIdempotencyRepo(testDB)

	if _, _, err := repo.Claim(ctx, "key-3", "auction.place_bid", "hash-3"); err != nil {
		t.Fatal(err)
	}

	wantResponse := json.RawMessage(`{"id":"bid-1"}`)
	if err := repo.Complete(ctx, "key-3", 201, wantResponse); err != nil {
		t.Fatal(err)
	}

	record, claimed, err := repo.Claim(ctx, "key-3", "auction.place_bid", "hash-3")
	if err != nil {
		t.Fatal(err)
	}
	if claimed {
		t.Fatal("completed retry reported claimed = true")
	}
	if record == nil || !record.Completed() || record.StatusCode == nil {
		t.Fatalf("completed record is incomplete: %+v", record)
	}
	if *record.StatusCode != 201 {
		t.Fatalf("status code = %d, want 201", *record.StatusCode)
	}
	assertJSONEqual(t, record.Response, wantResponse)

	if err := repo.Complete(ctx, "key-3", 200, json.RawMessage(`{"id":"other"}`)); !errors.Is(err, idempotency.ErrNotCompleted) {
		t.Fatalf("second Complete() error = %v, want ErrNotCompleted", err)
	}

	record, _, err = repo.Claim(ctx, "key-3", "auction.place_bid", "hash-3")
	if err != nil {
		t.Fatal(err)
	}
	if *record.StatusCode != 201 {
		t.Fatalf("second completion overwrote original result: %+v", record)
	}
	assertJSONEqual(t, record.Response, wantResponse)
}

func TestIdempotencyRepository_Claim_RollbackMakesKeyClaimable(t *testing.T) {
	ctx := context.Background()
	truncateIdempotencyRecords(t, ctx)
	repo := NewIdempotencyRepo(testDB)
	tx := NewTransactor(testDB)
	rollbackErr := errors.New("force rollback")

	err := tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		_, claimed, err := repo.Claim(txCtx, "key-4", "auction.place_bid", "hash-4")
		if err != nil {
			return err
		}
		if !claimed {
			t.Fatal("initial transactional claim reported claimed = false")
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("RunInTransaction() error = %v, want rollback error", err)
	}

	_, claimed, err := repo.Claim(ctx, "key-4", "auction.place_bid", "hash-4")
	if err != nil {
		t.Fatal(err)
	}
	if !claimed {
		t.Fatal("rolled-back key was not claimable")
	}
}

func truncateIdempotencyRecords(t *testing.T, ctx context.Context) {
	t.Helper()
	if err := TruncateTables(ctx, "idempotency_records"); err != nil {
		t.Fatal(err)
	}
}

func assertJSONEqual(t *testing.T, got, want json.RawMessage) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("invalid actual JSON %q: %v", got, err)
	}
	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("invalid expected JSON %q: %v", want, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("JSON response = %s, want %s", got, want)
	}
}
