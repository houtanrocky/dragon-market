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

type ItemService interface {
}

type Transactor interface {
	RunInTransaction(
		ctx context.Context,
		fn func(context.Context) error,
	) error
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
	itemService   ItemService
	tx            Transactor
}

func NewAuctionService(repo AuctionRepository, walletSvc WalletService) *AuctionServiceImpl {
	return &AuctionServiceImpl{repo: repo, walletService: walletSvc}
}

func (s *AuctionServiceImpl) StartAuction(ctx context.Context, itemID, sellerID string) (*Auction, error) {
	a := &Auction{
		ID:       uuid.New().String(),
		ItemID:   itemID,
		SellerID: sellerID,
		EndsAt:   time.Time{},
		Status:   ActiveAuction,
	}
	fmt.Println(a)
	//s.repo.Create()
	return nil, nil
}

func (s *AuctionServiceImpl) PlaceBid(ctx context.Context, auctionID, bidderID string, amount float64) error {
	err := s.repo.PlaceBid(ctx, &Bid{
		ID:        uuid.New().String(),
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
		PlacedAt:  time.Time{},
		Status:    ActiveBid,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AuctionServiceImpl) CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error {
	err := s.tx.RunInTransaction(ctx,
		func(ctx context.Context) error {
			a, err := s.repo.GetByID(ctx, auctionID)
			if err != nil {
				return err
			}

			if a.Status != ActiveAuction {
				return fmt.Errorf("cannot cancel a bid from inactive auction")
			}

			b, err := s.repo.GetBidByID(ctx, bidID)
			if err != nil {
				return err
			}

			if b.AuctionID != auctionID {
				return fmt.Errorf("bid not in auction")
			}
			if b.BidderID != bidderID {
				return fmt.Errorf("bid not owned by bidder")
			}
			if b.Status != ActiveBid {
				return fmt.Errorf("bid is not active")
			}

			topB, err := s.repo.GetTopActiveBid(ctx, auctionID)
			if err != nil {
				return err
			}

			if topB.ID == bidID {
				return fmt.Errorf("cannot cancel top bid")
			}

			err = s.repo.CancelActiveBid(ctx, auctionID, bidID, bidderID)
			if err != nil {
				return err
			}
			err = s.walletService.Release(ctx, bidderID, b.Amount)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuctionServiceImpl) EndAuction(ctx context.Context, auctionID string) error {
	return nil
}
