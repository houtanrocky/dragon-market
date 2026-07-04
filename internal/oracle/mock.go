package oracle

import (
	"context"
	"market-dragon/internal/gold"
)

type MockOracle struct {
	Prices map[string]gold.Amount
}

func (m *MockOracle) GetBasePrice(_ context.Context, itemID string) (gold.Amount, error) {
	return m.Prices[itemID], nil
}
