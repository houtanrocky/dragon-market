package auction

import (
	"context"
	"market-dragon/internal/guild"
	"time"

	"github.com/google/uuid"
)

type AuctionService interface {
}

// AuctionServiceImpl enforces:
// - one active auction per legendary item
// - min 5% bid increment over current top bid
// - 5-minute window extension on last-minute bids
// - reserve on bid, release on outbid, deduct on win
// - guild cannot bid on its own item
// - cancel bid only if not the current top bidder
type AuctionServiceImpl struct {
	repo          AuctionRepository
	walletService guild.WalletService
}

func NewAuctionService(repo AuctionRepository, walletService guild.WalletService) *AuctionServiceImpl {
	return &AuctionServiceImpl{repo: repo, walletService: walletService}
}

func (s *AuctionServiceImpl) StartAuction(ctx context.Context, itemID, sellerID string) (*Auction, error) {
	a := &Auction{
		ID:       uuid.New().String(),
		ItemID:   itemID,
		SellerID: sellerID,
		EndsAt:   time.Time{},
		Status:   Active,
	}
	s.repo.Create()
}

func (s *AuctionServiceImpl) PlaceBid(ctx context.Context, auctionID, bidderID string, amount float64) (*Bid, error) {
	panic("implement me")
}

func (s *AuctionServiceImpl) CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error {
	panic("implement me")
}

func (s *AuctionServiceImpl) EndAuction(ctx context.Context, auctionID string) error {
	panic("implement me")
}
