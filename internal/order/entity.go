package order

import "time"

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
	BuyerID  string
	Price    float64
	Status   Status
	ListedAt time.Time
}
