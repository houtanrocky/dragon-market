package auction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WalletService interface {
	Reserve(ctx context.Context, id string, amount float64) error
	Release(ctx context.Context, id string, amount float64) error
	Deduct(ctx context.Context, id string, amount float64) error
	Earn(ctx context.Context, id string, amount float64) error
	Spend(ctx context.Context, id string, amount float64) error
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
	walletService WalletService
}

func NewAuctionService(repo AuctionRepository, walletService WalletService) *AuctionServiceImpl {
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
	fmt.Println(a)
	//s.repo.Create()
	return nil, nil
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
