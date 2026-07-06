package oracle

import (
	"context"
	"market-dragon/internal/gold"
)

type MockOracle struct {
	Prices map[string]gold.Amount
	Err    error
}

func (m *MockOracle) GetBasePrice(_ context.Context, itemID string) (gold.Amount, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return m.Prices[itemID], nil
}
