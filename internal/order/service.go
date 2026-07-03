package order

import (
	"context"
	"fmt"
	"market-dragon/internal/item"
	"time"

	"github.com/google/uuid"
)

// Transactor begins a transaction and injects it into the context so all
// repositories that use getTx(ctx) participate atomically. It must wrap the
// same *sql.DB as the repositories passed to NewOrderService.
type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type WalletService interface {
	Spend(ctx context.Context, id string, amount float64) error
	Earn(ctx context.Context, id string, amount float64) error
}

type ItemService interface {
	TransferOwnership(ctx context.Context, itemId string, guildId string) error
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
	UpdateItem(ctx context.Context, item *item.Item) error
}

type OrderService interface {
	List(ctx context.Context, itemID, sellerID string, price float64) (*LimitOrder, error)
	Buy(ctx context.Context, orderID, buyerID string) error
	Cancel(ctx context.Context, orderID, sellerID string) error
}

// OrderServiceImpl handles listing and purchasing of Common/Rare items.
type OrderServiceImpl struct {
	repo      OrderRepository
	walletSvc WalletService
	itemSvc   ItemService
	tx        Transactor
}

func NewOrderService(r OrderRepository, wSvc WalletService, iSvc ItemService, t Transactor) *OrderServiceImpl {
	return &OrderServiceImpl{
		repo:      r,
		walletSvc: wSvc,
		itemSvc:   iSvc,
		tx:        t,
	}
}

func (s *OrderServiceImpl) List(ctx context.Context, itemID, sellerID string, price float64) (*LimitOrder, error) {
	if price <= 0 {
		return nil, fmt.Errorf("price must be positive, got %v", price)
	}

	var order *LimitOrder
	err := s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		it, err := s.itemSvc.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if it.OwnerID != sellerID {
			return fmt.Errorf("item %s is not owned by seller %s", itemID, sellerID)
		}
		if it.Status == item.ListedInAuction {
			return fmt.Errorf("item %s is listed in auction", itemID)
		}

		if it.Status == item.Free {
			it.Status = item.ListedInOrder
			if err := s.itemSvc.UpdateItem(ctx, it); err != nil {
				return err
			}
		}

		order = &LimitOrder{
			ID:       uuid.New().String(),
			ItemID:   itemID,
			SellerID: sellerID,
			Price:    price,
			Status:   Listed,
			ListedAt: time.Now(),
		}
		return s.repo.Create(ctx, order)
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderServiceImpl) Buy(ctx context.Context, orderID, buyerID string) error {
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		o, err := s.repo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}
		if o.Status != Listed {
			return fmt.Errorf("cannot buy order with status %q", o.Status)
		}
		if o.SellerID == buyerID {
			return fmt.Errorf("cannot buy your own listing")
		}

		o.BuyerID = &buyerID
		o.Status = Sold

		if err = s.walletSvc.Spend(ctx, buyerID, o.Price); err != nil {
			return err
		}
		if err = s.walletSvc.Earn(ctx, o.SellerID, o.Price); err != nil {
			return err
		}
		if err = s.itemSvc.TransferOwnership(ctx, o.ItemID, buyerID); err != nil {
			return err
		}

		otherListedOrders, err := s.repo.GetOrdersByItemIDAndStatus(ctx, o.ItemID, Listed)
		if err != nil {
			return err
		}

		if err := s.repo.Update(ctx, o); err != nil {
			return err
		}

		for _, order := range otherListedOrders {
			err := s.Cancel(ctx, order.ID, order.SellerID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *OrderServiceImpl) Cancel(ctx context.Context, orderID, sellerID string) error {
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		order, err := s.repo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}
		if order.SellerID != sellerID {
			return fmt.Errorf("cannot cancel an order listed by another seller")
		}
		if order.Status != Listed {
			return fmt.Errorf("cannot cancel order with status %q", order.Status)
		}

		order.Status = Canceled
		if err = s.repo.Update(ctx, order); err != nil {
			return err
		}

		otherListings, err := s.repo.GetOrdersByItemIDAndStatus(ctx, order.ItemID, Listed)
		if err != nil {
			return err
		}

		if len(otherListings) == 0 {
			it, err := s.itemSvc.GetItem(ctx, order.ItemID)
			if err != nil {
				return err
			}
			it.Status = item.Free
			err = s.itemSvc.UpdateItem(ctx, it)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
