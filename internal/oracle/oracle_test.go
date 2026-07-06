package oracle

import (
	"context"
	"errors"
	"market-dragon/internal/gold"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHTTPOracle_ValidatesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prices/soul reaver" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"base_price":1250}`))
	}))
	defer server.Close()

	client, err := NewHTTPOracle(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	price, err := client.GetBasePrice(context.Background(), "soul reaver")
	if err != nil || price != 1250 {
		t.Fatalf("price=%d err=%v", price, err)
	}
}

type sequenceOracle struct {
	mu     sync.Mutex
	values []gold.Amount
	errs   []error
	call   int
}

func (o *sequenceOracle) GetBasePrice(context.Context, string) (gold.Amount, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	i := o.call
	o.call++
	return o.values[i], o.errs[i]
}

func TestResilientOracle_KeepsLastValidPrice(t *testing.T) {
	upstream := &sequenceOracle{
		values: []gold.Amount{100, 0, 0},
		errs:   []error{nil, nil, errors.New("unavailable")},
	}
	oracle := NewResilientOracle(upstream)
	for i := 0; i < 3; i++ {
		price, err := oracle.GetBasePrice(context.Background(), "item-1")
		if err != nil || price != 100 {
			t.Fatalf("call %d: price=%d err=%v", i, price, err)
		}
	}
}

type memoryStore struct {
	ids    []string
	prices map[string]gold.Amount
	mu     sync.Mutex
}

func (s *memoryStore) ListItemIDs(context.Context) ([]string, error) { return s.ids, nil }
func (s *memoryStore) UpdateBasePrice(_ context.Context, id string, price gold.Amount) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prices[id] = price
	return nil
}

func TestUpdater_RefreshesImmediatelyAndPeriodically(t *testing.T) {
	store := &memoryStore{ids: []string{"item-1"}, prices: make(map[string]gold.Amount)}
	source := &MockOracle{Prices: map[string]gold.Amount{"item-1": 250}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go NewUpdater(source, store, 5*time.Millisecond, nil).Run(ctx)

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		store.mu.Lock()
		price := store.prices["item-1"]
		store.mu.Unlock()
		if price == 250 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("price was not refreshed")
}
