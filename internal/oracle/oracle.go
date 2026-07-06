package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"market-dragon/internal/gold"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// PriceOracle fetches live base prices from the external Oracle Price Service.
// Responses may be stale, zero, or negative, callers must validate before use.
type PriceOracle interface {
	GetBasePrice(ctx context.Context, itemID string) (gold.Amount, error)
}

var (
	ErrInvalidPrice = errors.New("oracle returned an invalid price")
	ErrNoValidPrice = errors.New("no valid oracle price is available")
)

// HTTPOracle implements the external boundary. The service is expected to
// expose GET {baseURL}/prices/{itemID} returning {"base_price": 123}.
type HTTPOracle struct {
	baseURL string
	client  *http.Client
}

func NewHTTPOracle(baseURL string, client *http.Client) (*HTTPOracle, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid oracle URL %q", baseURL)
	}
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &HTTPOracle{baseURL: baseURL, client: client}, nil
}

func (o *HTTPOracle) GetBasePrice(ctx context.Context, itemID string) (gold.Amount, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		o.baseURL+"/prices/"+url.PathEscape(itemID), nil)
	if err != nil {
		return 0, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request oracle: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return 0, fmt.Errorf("oracle returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		BasePrice gold.Amount `json:"base_price"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return 0, fmt.Errorf("decode oracle response: %w", err)
	}
	if payload.BasePrice <= 0 {
		return 0, fmt.Errorf("%w: %d", ErrInvalidPrice, payload.BasePrice)
	}
	return payload.BasePrice, nil
}

// ResilientOracle retains the last valid value for each item. Temporary
// upstream errors and invalid values therefore do not replace a known price.
type ResilientOracle struct {
	upstream PriceOracle
	mu       sync.RWMutex
	prices   map[string]gold.Amount
}

func NewResilientOracle(upstream PriceOracle) *ResilientOracle {
	return &ResilientOracle{upstream: upstream, prices: make(map[string]gold.Amount)}
}

func (o *ResilientOracle) GetBasePrice(ctx context.Context, itemID string) (gold.Amount, error) {
	price, err := o.upstream.GetBasePrice(ctx, itemID)
	if err == nil && price > 0 {
		o.mu.Lock()
		o.prices[itemID] = price
		o.mu.Unlock()
		return price, nil
	}
	if err == nil {
		err = fmt.Errorf("%w: %d", ErrInvalidPrice, price)
	}
	o.mu.RLock()
	cached, ok := o.prices[itemID]
	o.mu.RUnlock()
	if ok {
		return cached, nil
	}
	return 0, fmt.Errorf("%w for item %q: %v", ErrNoValidPrice, itemID, err)
}

type PriceStore interface {
	ListItemIDs(ctx context.Context) ([]string, error)
	UpdateBasePrice(ctx context.Context, itemID string, price gold.Amount) error
}

// Updater refreshes every known item immediately and then at the configured
// interval. A failure for one item does not prevent other items from updating.
type Updater struct {
	oracle   PriceOracle
	store    PriceStore
	interval time.Duration
	onError  func(error)
}

func NewUpdater(source PriceOracle, store PriceStore, interval time.Duration, onError func(error)) *Updater {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if onError == nil {
		onError = func(error) {}
	}
	return &Updater{oracle: source, store: store, interval: interval, onError: onError}
}

func (u *Updater) Run(ctx context.Context) {
	u.refresh(ctx)
	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.refresh(ctx)
		}
	}
}

func (u *Updater) refresh(ctx context.Context) {
	ids, err := u.store.ListItemIDs(ctx)
	if err != nil {
		u.onError(fmt.Errorf("list items for oracle refresh: %w", err))
		return
	}
	for _, id := range ids {
		price, err := u.oracle.GetBasePrice(ctx, id)
		if err != nil {
			u.onError(fmt.Errorf("refresh price for %s: %w", id, err))
			continue
		}
		if price <= 0 {
			u.onError(fmt.Errorf("refresh price for %s: %w: %d", id, ErrInvalidPrice, price))
			continue
		}
		if err := u.store.UpdateBasePrice(ctx, id, price); err != nil {
			u.onError(fmt.Errorf("store price for %s: %w", id, err))
		}
	}
}
