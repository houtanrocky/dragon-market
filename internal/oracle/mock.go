package oracle

import "context"

type MockOracle struct {
	Prices map[string]float64
}

func (m *MockOracle) GetBasePrice(_ context.Context, itemID string) (float64, error) {
	return m.Prices[itemID], nil
}
