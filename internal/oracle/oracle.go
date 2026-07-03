package oracle

import "context"

// PriceOracle fetches live base prices from the external Oracle Price Service.
// Responses may be stale, zero, or negative — callers must validate before use.
type PriceOracle interface {
	GetBasePrice(ctx context.Context, itemID string) (float64, error)
}
