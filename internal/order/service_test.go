package order

import (
	"context"
	"errors"
	"fmt"
	"market-dragon/internal/gold"
	"market-dragon/internal/guild"
	"market-dragon/internal/item"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
// --- Mock Order Repo
type MockOrderRepo struct {
	orders map[string]*LimitOrder
}

func (r *MockOrderRepo) GetByIDForUpdate(ctx context.Context, id string) (*LimitOrder, error) {
	return r.GetByID(ctx, id)
}

func (r *MockOrderRepo) CancelOtherListed(ctx context.Context, itemID, exceptOrderID string) error {
	for _, o := range r.orders {
		if o.ItemID == itemID && o.ID != exceptOrderID && o.Status == Listed {
			o.Status = Canceled
		}
	}

	return nil
}

func (r *MockOrderRepo) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (r *MockOrderRepo) GetOrdersByItemIDAndStatus(ctx context.Context, itemID string, status Status) ([]*LimitOrder, error) {

	var orders []*LimitOrder
	for _, order := range r.orders {
		if order.ItemID == itemID && order.Status == status {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (r *MockOrderRepo) Create(ctx context.Context, o *LimitOrder) error {
	r.orders[o.ID] = o
	return nil
}

func (r *MockOrderRepo) GetByID(ctx context.Context, id string) (*LimitOrder, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func (r *MockOrderRepo) Update(ctx context.Context, o *LimitOrder) error {
	r.orders[o.ID] = o
	return nil
}

// --- Mock Wallet Service
type MockWalletService struct {
	guilds map[string]*guild.Guild
}

func (s *MockWalletService) Spend(ctx context.Context, id string, amount gold.Amount) error {
	g, ok := s.guilds[id]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	if g.Gold < amount {
		return fmt.Errorf("insufficient balance")
	}
	if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
		return fmt.Errorf("DailyLimit reached")
	}
	g.Gold -= amount
	return nil
}

func (s *MockWalletService) Earn(ctx context.Context, id string, amount gold.Amount) error {
	g, ok := s.guilds[id]
	if !ok {
		return fmt.Errorf("guild not found")
	}
	g.Gold += amount
	return nil
}

// --- Mock Item Service
type MockItemService struct {
	items map[string]*item.Item
}

func (s *MockItemService) TransferOwnership(ctx context.Context, itemID, sellerID, buyerID string) error {
	i, ok := s.items[itemID]
	if !ok {
		return fmt.Errorf("item %v not found", itemID)
	}
	if i.OwnerID == buyerID {
		return fmt.Errorf("guild %v already owns %v", buyerID, itemID)
	}
	i.OwnerID = buyerID
	return nil
}

func (s *MockItemService) GetItem(ctx context.Context, itemID string) (*item.Item, error) {
	i, ok := s.items[itemID]
	if !ok {
		return nil, fmt.Errorf("item %v not found", itemID)
	}
	return i, nil
}
func (m *MockItemService) GetItemForUpdate(ctx context.Context, id string) (*item.Item, error) {
	return m.GetItem(ctx, id)
}

func (s *MockItemService) UpdateItem(ctx context.Context, item *item.Item) error {
	s.items[item.ID] = item
	return nil
}

//func (r *MockItemRepo) ListFree(ctx context.Context) ([]*item.Item, error) {
//	var result []*item.Item
//	for _, it := range r.items {
//		if it.Status != item.ListedInAuction {
//			result = append(result, it)
//		}
//	}
//	return result, nil
//}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const (
	sellerID = "seller-1"
	buyerID  = "buyer-1"
	itemID   = "item-1"
	item2ID  = "item-2"
	item3ID  = "item-3"
	orderID  = "order-1"
)

func defaultGuilds() map[string]*guild.Guild {
	return map[string]*guild.Guild{
		sellerID: {ID: sellerID, Gold: 0},
		buyerID:  {ID: buyerID, Gold: 10000, DailyLimit: 5000, DailySpent: 0},
	}
}

func defaultItem() map[string]*item.Item {
	return map[string]*item.Item{
		itemID:  {ID: itemID, Name: "Sword", Type: item.Common, OwnerID: sellerID, Status: item.ListedInOrder, BasePrice: 100},
		item2ID: {ID: item2ID, Name: "Knife", Type: item.Common, OwnerID: sellerID, Status: item.Free, BasePrice: 100},
		item3ID: {
			ID:        item3ID,
			Name:      "Unavailable Sword",
			Type:      item.Rare,
			OwnerID:   sellerID,
			Status:    item.ListedInAuction,
			BasePrice: 100,
		},
	}
}

func listedOrder(price gold.Amount) *LimitOrder {
	return &LimitOrder{
		ID:       orderID,
		ItemID:   itemID,
		SellerID: sellerID,
		Price:    price,
		Status:   Listed,
		ListedAt: time.Unix(857174400, 0).UTC(),
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestService_List_Success(t *testing.T) {
	ctx := context.Background()
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}
	oSvc := NewOrderService(r, wSvc, iSvc, r)

	order, err := oSvc.List(ctx, itemID, sellerID, 200)
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != Listed {
		t.Errorf("expected status Listed, got %q", order.Status)
	}
	i, err := iSvc.GetItem(ctx, itemID)
	if err != nil {
		t.Fatal(err)
	}
	if i.Status != item.ListedInOrder {
		t.Errorf("expected status Listed, got %q", order.Status)
	}
}

func TestService_List_InvalidPrice(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	_, err := NewOrderService(oRepo, wSvc, iSvc, r).List(ctx, itemID, sellerID, 0)
	if err == nil {
		t.Error("expected error for zero price, got nil")
	}
}

func TestService_List_NotOwner(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	_, err := NewOrderService(oRepo, wSvc, iSvc, r).List(ctx, itemID, buyerID, 100)
	if err == nil {
		t.Error("expected error when non-owner lists item, got nil")
	}
}

func TestService_List_ItemListedInAuction(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	_, err := NewOrderService(oRepo, wSvc, iSvc, r).List(ctx, item3ID, sellerID, 100)
	if err == nil {
		t.Error("expected error for listing an item listed in auction, got nil")
	}
}

// ---------------------------------------------------------------------------
// Buy
// ---------------------------------------------------------------------------

func TestService_Buy_Success(t *testing.T) {
	ctx := context.Background()
	price := gold.Amount(2000)
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: listedOrder(price)}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	oSvc := NewOrderService(oRepo, wSvc, iSvc, r)

	if err := oSvc.Buy(ctx, orderID, buyerID); err != nil {
		t.Fatal(err)
	}

	o := oRepo.orders[orderID]
	if o.Status != Sold {
		t.Errorf("expected order status Sold, got %q", o.Status)
	}
	if o.BuyerID == nil || *o.BuyerID != buyerID {
		t.Errorf("expected BuyerID %q, got %v", buyerID, o.BuyerID)
	}
	if wSvc.guilds[buyerID].Gold != 10000-o.Price {
		t.Errorf("buyer gold not debited correctly")
	}
	if wSvc.guilds[sellerID].Gold != o.Price {
		t.Errorf("seller gold not credited correctly")
	}
	if iSvc.items[itemID].OwnerID != buyerID {
		t.Error("item ownership not transferred to buyer")
	}
}

func TestService_Buy_InsufficientGold(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, buyerID); err == nil {
		t.Error("expected error for insufficient gold, got nil")
	}
}

func TestService_Buy_DailyLimitReached(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: listedOrder(100)}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}
	wSvc.guilds[buyerID].DailySpent = wSvc.guilds[buyerID].DailyLimit + 100000

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, buyerID); err == nil {
		t.Error("expected error for daily limit, got nil")
	}
}

func TestService_Buy_DailyLimitZeroMeansUnlimited(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: listedOrder(100)}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}
	wSvc.guilds[buyerID].DailyLimit = 0
	wSvc.guilds[buyerID].DailySpent = 999999

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, buyerID); err != nil {
		t.Errorf("expected no error when DailyLimit=0, got %v", err)
	}
}

func TestService_Buy_AlreadySold(t *testing.T) {
	ctx := context.Background()
	o := listedOrder(100)
	o.Status = Sold
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: o}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, buyerID); err == nil {
		t.Error("expected error buying already-sold order, got nil")
	}
}

func TestService_Buy_CanceledOrderNotBuyable(t *testing.T) {
	ctx := context.Background()
	o := listedOrder(100)
	o.Status = Canceled
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: o}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, buyerID); err == nil {
		t.Error("expected error buying canceled order, got nil")
	}
}

func TestService_Buy_CannotBuyOwnListing(t *testing.T) {
	ctx := context.Background()
	o := listedOrder(100)
	o.Status = Canceled
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: o}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	// seller tries to buy their own listing
	if err := NewOrderService(oRepo, wSvc, iSvc, r).Buy(ctx, orderID, sellerID); err == nil {
		t.Error("expected error when seller buys own listing, got nil")
	}
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

func TestService_Cancel_Success(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: listedOrder(100)}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{items: defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Cancel(ctx, orderID, sellerID); err != nil {
		t.Fatal(err)
	}
	if oRepo.orders[orderID].Status != Canceled {
		t.Error("expected order status Canceled")
	}
	if iSvc.items[itemID].Status != item.Free {
		t.Error("expected item to be free again after cancel")
	}
}

func TestService_Cancel_WrongSeller(t *testing.T) {
	ctx := context.Background()
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: listedOrder(100)}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{items: defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Cancel(ctx, orderID, buyerID); err == nil {
		t.Error("expected error when non-seller cancels, got nil")
	}
}

func TestService_Cancel_AlreadySold(t *testing.T) {
	//ctx := context.Background()
	//o := listedOrder(100)
	//o.Status = Sold
	//oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: o}}
	//gRepo := &MockGuildRepo{guilds: defaultGuilds()}
	//iRepo := &MockItemRepo{items: map[string]*item.Item{itemID: defaultItem()}}

	ctx := context.Background()
	o := listedOrder(100)
	o.Status = Sold
	oRepo := &MockOrderRepo{orders: map[string]*LimitOrder{orderID: o}}
	r := &MockOrderRepo{orders: map[string]*LimitOrder{}}

	iSvc := &MockItemService{items: defaultItem()}
	wSvc := &MockWalletService{defaultGuilds()}

	if err := NewOrderService(oRepo, wSvc, iSvc, r).Cancel(ctx, orderID, sellerID); err == nil {
		t.Error("expected error canceling already-sold order, got nil")
	}
}
