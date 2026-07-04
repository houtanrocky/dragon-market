package order

import (
	"market-dragon/internal/gold"
	"time"
)

type Status string

const (
	Listed   Status = "listed"
	Sold     Status = "sold"
	Canceled Status = "canceled"
)

// LimitOrder is used for Common and Rare items only.
type LimitOrder struct {
	ID       string
	ItemID   string
	SellerID string
	BuyerID  *string // nil until sold
	Price    gold.Amount
	Status   Status
	ListedAt time.Time
}
