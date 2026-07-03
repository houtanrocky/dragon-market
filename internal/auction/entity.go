package auction

import "time"

type Status string

const (
	Active Status = "active"
	Ended  Status = "ended"
)

type Auction struct {
	ID       string
	ItemID   string
	SellerID string
	EndsAt   time.Time
	Status   Status
}

type Bid struct {
	ID        string
	AuctionID string
	BidderID  string
	Amount    float64
	PlacedAt  time.Time
}
