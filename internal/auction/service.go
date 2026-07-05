package auction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"market-dragon/internal/gold"
	"market-dragon/internal/item"
)

type WalletService interface {
	Reserve(ctx context.Context, guildID string, amount gold.Amount) error
	Release(ctx context.Context, guildID string, amount gold.Amount) error
	Deduct(ctx context.Context, guildID string, amount gold.Amount) error
	Earn(ctx context.Context, guildID string, amount gold.Amount) error
}

type ItemService interface {
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
	MarkListedInAuction(ctx context.Context, itemID, sellerID string) error
	TransferFromAuction(ctx context.Context, itemID, sellerID, winnerID string) error
	ReleaseFromAuction(ctx context.Context, itemID string) error
}

type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type Config struct {
	Duration        time.Duration
	ExtensionWindow time.Duration
	Extension       time.Duration
}

func DefaultConfig() Config {
	return Config{24 * time.Hour, 5 * time.Minute, 5 * time.Minute}
}

type Option func(*AuctionServiceImpl)

func WithClock(clock Clock) Option {
	return func(service *AuctionServiceImpl) { service.clock = clock }
}

//func WithConfig(config Config) Option {
//	return func(service *AuctionServiceImpl) { service.config = config }
//}

type AuctionServiceImpl struct {
	repo          AuctionRepository
	walletService WalletService
	itemService   ItemService
	tx            Transactor
	clock         Clock
	config        Config
}

func NewAuctionService(repo AuctionRepository, wallet WalletService, items ItemService, tx Transactor, options ...Option) *AuctionServiceImpl {
	service := &AuctionServiceImpl{
		repo: repo, walletService: wallet, itemService: items, tx: tx,
		clock: realClock{}, config: DefaultConfig(),
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *AuctionServiceImpl) StartAuction(ctx context.Context, itemID, sellerID string) (*Auction, error) {
	var created *Auction
	err := s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		it, err := s.itemService.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if it.Type != item.Legendary {
			return ErrItemNotLegendary
		}
		if it.OwnerID != sellerID {
			return ErrItemNotOwnedBySeller
		}
		if it.Status != item.Free {
			return ErrItemNotAvailable
		}

		active, err := s.repo.GetActiveAuctionByItemID(ctx, itemID)
		if err != nil && !errors.Is(err, ErrAuctionNotFound) {
			return err
		}
		if active != nil {
			return ErrActiveAuctionExists
		}
		if err := s.itemService.MarkListedInAuction(ctx, itemID, sellerID); err != nil {
			return err
		}

		now := s.clock.Now()
		created = &Auction{
			ID: uuid.NewString(), ItemID: itemID, SellerID: sellerID,
			EndsAt: now.Add(s.config.Duration), Status: ActiveAuction,
		}
		return s.repo.CreateAuction(ctx, created)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *AuctionServiceImpl) PlaceBid(ctx context.Context, auctionID, bidderID string, amount gold.Amount) (*Bid, error) {
	if amount <= 0 {
		return nil, ErrInvalidBidAmount
	}

	var bid *Bid

	err := s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		auction, err := s.repo.GetAuctionByID(ctx, auctionID)
		if err != nil {
			return err
		}

		now := s.clock.Now()

		if auction.Status != ActiveAuction {
			return ErrAuctionNotActive
		}
		if !now.Before(auction.EndsAt) {
			return ErrAuctionFinished
		}
		if auction.SellerID == bidderID {
			return ErrSellerCannotBid
		}

		topBid, err := s.repo.GetTopActiveBid(ctx, auctionID)
		if err != nil && !errors.Is(err, ErrBidNotFound) {
			return err
		}

		if topBid != nil && amount < minimumNextBid(topBid.Amount) {
			return fmt.Errorf("minimum bid is %d: %w", minimumNextBid(topBid.Amount), ErrBidTooLow)
		}

		if topBid != nil && topBid.BidderID == bidderID {
			if err := s.walletService.Reserve(ctx, bidderID, amount-topBid.Amount); err != nil {
				return err
			}
		} else {
			if err := s.walletService.Reserve(ctx, bidderID, amount); err != nil {
				return err
			}
			if topBid != nil {
				if err := s.walletService.Release(ctx, topBid.BidderID, topBid.Amount); err != nil {
					return err
				}
			}
		}

		if topBid != nil {
			if err := s.repo.MarkBidOutbid(ctx, topBid.ID); err != nil {
				return err
			}
		}

		// Create new bid
		newBid := &Bid{
			ID:        uuid.NewString(),
			AuctionID: auctionID,
			BidderID:  bidderID,
			Amount:    amount,
			PlacedAt:  now,
			Status:    ActiveBid,
		}

		var err2 error
		bid, err2 = s.repo.CreateBid(ctx, newBid)
		if err2 != nil {
			return err2
		}

		// Extend auction if bid placed near the end
		remaining := auction.EndsAt.Sub(now)
		if remaining > 0 && remaining <= s.config.ExtensionWindow {
			if err := s.repo.ExtendActiveAuction(ctx, auctionID, auction.EndsAt.Add(s.config.Extension)); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return bid, nil
}

func (s *AuctionServiceImpl) CancelBid(ctx context.Context, auctionID, bidID, bidderID string) error {
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		auction, err := s.repo.GetAuctionByID(ctx, auctionID)
		if err != nil {
			return err
		}
		if auction.Status != ActiveAuction || !s.clock.Now().Before(auction.EndsAt) {
			return ErrAuctionNotActive
		}
		bid, err := s.repo.GetBidByID(ctx, bidID)
		if err != nil {
			return err
		}
		if bid.AuctionID != auctionID || bid.BidderID != bidderID || bid.Status != OutbidBid {
			return ErrBidNotCancellable
		}
		return s.repo.CancelOutbidBid(ctx, auctionID, bidID, bidderID)
	})
}

func (s *AuctionServiceImpl) EndAuction(ctx context.Context, auctionID string) error {
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		auction, err := s.repo.GetAuctionByID(ctx, auctionID)
		if err != nil {
			return err
		}
		if auction.Status != ActiveAuction {
			return ErrAuctionNotActive
		}
		if s.clock.Now().Before(auction.EndsAt) {
			return ErrAuctionNotFinished
		}

		top, err := s.repo.GetTopActiveBid(ctx, auctionID)
		if err != nil && !errors.Is(err, ErrBidNotFound) {
			return err
		}
		if top == nil {
			if err := s.itemService.ReleaseFromAuction(ctx, auction.ItemID); err != nil {
				return err
			}
			return s.repo.EndActiveAuction(ctx, auctionID)
		}
		if err := s.walletService.Deduct(ctx, top.BidderID, top.Amount); err != nil {
			return err
		}
		if err := s.walletService.Earn(ctx, auction.SellerID, top.Amount); err != nil {
			return err
		}
		if err := s.itemService.TransferFromAuction(ctx, auction.ItemID, auction.SellerID, top.BidderID); err != nil {
			return err
		}
		if err := s.repo.MarkBidWinning(ctx, top.ID); err != nil {
			return err
		}
		return s.repo.EndActiveAuction(ctx, auctionID)
	})
}

func (s *AuctionServiceImpl) GetBid(ctx context.Context, bidID string) (*Bid, error) {
	b, err := s.GetBid(ctx, bidID)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func minimumNextBid(current gold.Amount) gold.Amount {
	increment := (current + 19) / 20 // ceil(current * 5 / 100)
	return current + increment
}
