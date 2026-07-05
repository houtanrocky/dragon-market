package auction

import (
	"context"
	"time"
)

// AuctionRepository persists auctions and bids. GetAuctionByID must lock the
// auction row when ctx contains a transaction, serializing state changes.
type AuctionRepository interface {
	CreateAuction(ctx context.Context, auction *Auction) error
	GetAuctionByID(ctx context.Context, auctionID string) (*Auction, error)
	GetActiveAuctionByItemID(ctx context.Context, itemID string) (*Auction, error)
	ExtendActiveAuction(ctx context.Context, auctionID string, endsAt time.Time) error
	EndActiveAuction(ctx context.Context, auctionID string) error

	CreateBid(ctx context.Context, bid *Bid) (*Bid, error)
	GetBidByID(ctx context.Context, bidID string) (*Bid, error)
	GetTopActiveBid(ctx context.Context, auctionID string) (*Bid, error)
	MarkBidOutbid(ctx context.Context, bidID string) error
	CancelOutbidBid(ctx context.Context, auctionID, bidID, bidderID string) error
	MarkBidWinning(ctx context.Context, bidID string) error
}
