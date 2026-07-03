package auction

import "time"

type AuctionStatus string
type BidStatus string

const (
	ActiveAuction AuctionStatus = "active"
	EndedAuction  AuctionStatus = "ended"
)

const (
	ActiveBid    BidStatus = "active"
	CancelledBid BidStatus = "cancelled"
)

type Auction struct {
	ID       string
	ItemID   string
	SellerID string
	EndsAt   time.Time
	Status   AuctionStatus
}

type Bid struct {
	ID        string
	AuctionID string
	BidderID  string
	Amount    float64
	PlacedAt  time.Time
	Status    BidStatus // "active", "cancelled", "winning" (optional)
}
