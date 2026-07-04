package auction

import (
	"time"

	"market-dragon/internal/gold"
)

type AuctionStatus string
type BidStatus string

const (
	ActiveAuction AuctionStatus = "active"
	EndedAuction  AuctionStatus = "ended"
)

const (
	ActiveBid    BidStatus = "active"
	OutbidBid    BidStatus = "outbid"
	CancelledBid BidStatus = "cancelled"
	WinningBid   BidStatus = "winning"
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
	Amount    gold.Amount
	PlacedAt  time.Time
	Status    BidStatus
}
