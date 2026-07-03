package auction

import "context"

type AuctionRepository interface {
	Create(ctx context.Context, a *Auction) error
	GetByID(ctx context.Context, id string) (*Auction, error)
	GetActiveByItemID(ctx context.Context, itemID string) (*Auction, error)
	Update(ctx context.Context, a *Auction) error

	PlaceBid(ctx context.Context, b *Bid) error
	GetBidsByAuction(ctx context.Context, auctionID string) ([]*Bid, error)
	GetTopActiveBid(
		ctx context.Context,
		auctionID string,
	) (*Bid, error)
	CancelActiveBid(
		ctx context.Context,
		auctionID string,
		bidID string,
		bidderID string,
	) error
	GetBidByID(ctx context.Context, bidID string) (*Bid, error)
}
