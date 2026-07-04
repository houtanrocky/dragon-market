package order

import (
	"context"
	"fmt"
	"market-dragon/internal/gold"
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
	Spend(ctx context.Context, id string, amount gold.Amount) error
	Earn(ctx context.Context, id string, amount gold.Amount) error
}

type ItemService interface {
	TransferOwnership(ctx context.Context, itemId string, sellerID string, buyerID string) error
	GetItem(ctx context.Context, itemID string) (*item.Item, error)
	GetItemForUpdate(ctx context.Context, itemID string) (*item.Item, error)
	UpdateItem(ctx context.Context, item *item.Item) error
}

type OrderService interface {
	List(ctx context.Context, itemID, sellerID string, price gold.Amount) (*LimitOrder, error)
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

func (s *OrderServiceImpl) List(ctx context.Context, itemID, sellerID string, price gold.Amount) (*LimitOrder, error) {
	if price <= 0 {
		return nil, fmt.Errorf("price must be positive, got %v", price)
	}

	var order *LimitOrder
	err := s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		it, err := s.itemSvc.GetItemForUpdate(ctx, itemID)
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

		it, err := s.itemSvc.GetItemForUpdate(ctx, o.ItemID)
		if err != nil {
			return err
		}

		o, err = s.repo.GetByIDForUpdate(ctx, orderID)
		if err != nil {
			return err
		}

		if o.ItemID != it.ID {
			return fmt.Errorf("order item changed unexpectedly")
		}
		if o.Status != Listed {
			return fmt.Errorf("cannot buy order with status %q", o.Status)
		}
		if o.SellerID == buyerID {
			return fmt.Errorf("cannot buy your own listing")
		}
		if it.OwnerID != o.SellerID {
			return fmt.Errorf("seller no longer owns item")
		}
		if it.Status != item.ListedInOrder {
			return fmt.Errorf("item is not listed in an order")
		}

		o.BuyerID = &buyerID
		o.Status = Sold

		if err = s.walletSvc.Spend(ctx, buyerID, o.Price); err != nil {
			return err
		}
		if err = s.walletSvc.Earn(ctx, o.SellerID, o.Price); err != nil {
			return err
		}
		if err = s.itemSvc.TransferOwnership(ctx, o.ItemID, o.SellerID, buyerID); err != nil {
			return err
		}

		if err := s.repo.Update(ctx, o); err != nil {
			return err
		}

		return s.repo.CancelOtherListed(ctx, o.ItemID, o.ID)
	})
}

func (s *OrderServiceImpl) Cancel(ctx context.Context, orderID, sellerID string) error {
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		o, err := s.repo.GetByID(ctx, orderID)
		if err != nil {
			return err
		}

		_, err = s.itemSvc.GetItemForUpdate(ctx, o.ItemID)
		if err != nil {
			return err
		}

		o, err = s.repo.GetByIDForUpdate(ctx, orderID)
		if err != nil {
			return err
		}

		if o.SellerID != sellerID {
			return fmt.Errorf("cannot cancel an order listed by another seller")
		}
		if o.Status != Listed {
			return fmt.Errorf("cannot cancel order with status %q", o.Status)
		}

		o.Status = Canceled
		if err = s.repo.Update(ctx, o); err != nil {
			return err
		}

		otherListings, err := s.repo.GetOrdersByItemIDAndStatus(ctx, o.ItemID, Listed)
		if err != nil {
			return err
		}

		if len(otherListings) == 0 {
			it, err := s.itemSvc.GetItem(ctx, o.ItemID)
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
