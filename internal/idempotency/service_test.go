package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	record      *IdempotencyRecord
	claimed     bool
	claimErr    error
	completeErr error
	completeKey string
	statusCode  int
	response    json.RawMessage
}

func (r *fakeRepository) Claim(context.Context, string, string, string) (*IdempotencyRecord, bool, error) {
	return r.record, r.claimed, r.claimErr
}

func (r *fakeRepository) Complete(_ context.Context, key string, statusCode int, response json.RawMessage) error {
	r.completeKey = key
	r.statusCode = statusCode
	r.response = response
	return r.completeErr
}

type fakeTransactor struct {
	ctx context.Context
}

func (t fakeTransactor) RunInTransaction(_ context.Context, fn func(context.Context) error) error {
	return fn(t.ctx)
}

func TestIdempotencyService_Execute_NewRequest(t *testing.T) {
	txCtx := context.WithValue(context.Background(), struct{ name string }{"tx"}, "present")
	repo := &fakeRepository{record: &IdempotencyRecord{Key: "key-1"}, claimed: true}
	service := NewService(repo, fakeTransactor{ctx: txCtx})
	wantResponse := json.RawMessage(`{"id":"bid-1"}`)
	runCalls := 0

	result, err := service.Execute(context.Background(), "key-1", "auction.place_bid", "hash-1", func(ctx context.Context) (int, json.RawMessage, error) {
		runCalls++
		if ctx != txCtx {
			t.Fatal("operation did not receive transactional context")
		}
		return 201, wantResponse, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if runCalls != 1 {
		t.Fatalf("operation calls = %d, want 1", runCalls)
	}
	if result.Replayed || result.StatusCode != 201 || string(result.Response) != string(wantResponse) {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.completeKey != "key-1" || repo.statusCode != 201 || string(repo.response) != string(wantResponse) {
		t.Fatalf("unexpected completion: key=%q status=%d response=%s", repo.completeKey, repo.statusCode, repo.response)
	}
}

func TestIdempotencyService_Execute_ReplaysCompletedRequest(t *testing.T) {
	statusCode := 201
	completedAt := time.Now()
	wantResponse := json.RawMessage(`{"id":"bid-1"}`)
	repo := &fakeRepository{
		record: &IdempotencyRecord{
			StatusCode:  &statusCode,
			Response:    wantResponse,
			CompletedAt: &completedAt,
		},
		claimed: false,
	}
	service := NewService(repo, fakeTransactor{ctx: context.Background()})

	result, err := service.Execute(context.Background(), "key-1", "auction.place_bid", "hash-1", func(context.Context) (int, json.RawMessage, error) {
		t.Fatal("operation ran for a completed retry")
		return 0, nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Replayed || result.StatusCode != 201 || string(result.Response) != string(wantResponse) {
		t.Fatalf("unexpected replay result: %+v", result)
	}
	if repo.completeKey != "" {
		t.Fatal("replayed request was completed again")
	}
}

func TestIdempotencyService_Execute_IncompleteRetry(t *testing.T) {
	repo := &fakeRepository{record: &IdempotencyRecord{}, claimed: false}
	service := NewService(repo, fakeTransactor{ctx: context.Background()})

	_, err := service.Execute(context.Background(), "key-1", "auction.place_bid", "hash-1", func(context.Context) (int, json.RawMessage, error) {
		t.Fatal("operation ran for an incomplete retry")
		return 0, nil, nil
	})
	if !errors.Is(err, ErrNotCompleted) {
		t.Fatalf("Execute() error = %v, want ErrNotCompleted", err)
	}
}

func TestIdempotencyService_Execute_OperationFailureDoesNotComplete(t *testing.T) {
	repo := &fakeRepository{record: &IdempotencyRecord{Key: "key-1"}, claimed: true}
	service := NewService(repo, fakeTransactor{ctx: context.Background()})
	wantErr := errors.New("operation failed")

	_, err := service.Execute(context.Background(), "key-1", "auction.place_bid", "hash-1", func(context.Context) (int, json.RawMessage, error) {
		return 0, nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want operation error", err)
	}
	if repo.completeKey != "" {
		t.Fatal("failed operation was completed")
	}
}

func TestIdempotencyService_Execute_PropagatesRepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		repo *fakeRepository
	}{
		{name: "claim", repo: &fakeRepository{claimErr: ErrKeyConflict}},
		{name: "complete", repo: &fakeRepository{record: &IdempotencyRecord{}, claimed: true, completeErr: ErrNotCompleted}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.repo, fakeTransactor{ctx: context.Background()})
			_, err := service.Execute(context.Background(), "key-1", "auction.place_bid", "hash-1", func(context.Context) (int, json.RawMessage, error) {
				return 201, json.RawMessage(`{}`), nil
			})
			if err == nil {
				t.Fatal("Execute() error = nil, want repository error")
			}
		})
	}
}
