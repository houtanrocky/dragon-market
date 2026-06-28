package order

import (
	"context"
	"market-dragon/internal/guild"
	"testing"
	"time"
)

// ----------- MockOrderRepo -------------
type MockOrderRepo struct {
	orders map[string]*LimitOrder
}

func (r *MockOrderRepo) Create(ctx context.Context, o *LimitOrder) error {
	r.orders[o.ID] = o
	return nil
}

func (r *MockOrderRepo) GetByID(ctx context.Context, id string) (*LimitOrder, error) {
	return r.orders[id], nil
}

func (r *MockOrderRepo) Update(ctx context.Context, o *LimitOrder) error {
	r.orders[o.ID] = o
	return nil
}

func (r *MockOrderRepo) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	return nil
}

// ----------- MockGuildRepo --------------
type MockGuildRepo struct {
	guilds map[string]*guild.Guild
}

func (m *MockGuildRepo) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *MockGuildRepo) Get(ctx context.Context, id string) (*guild.Guild, error) {
	g, ok := m.guilds[id]
	if !ok {
		return nil, guild.ErrGuildNotFound
	}
	return g, nil
}

func (m *MockGuildRepo) Update(ctx context.Context, g *guild.Guild) error {
	m.guilds[g.ID] = g

	return nil
}

// ------------------------------------

func TestService_List(t *testing.T) {
	// Arrange

	// -- ctx
	ctx := context.Background()

	const (
		testGuildID           = "guild-1"
		testInitialGold       = 200
		testInitialReserve    = 100
		testInitialDailyLimit = 200
		testInitialDailySpent = 0
	)

	gRepo := &MockGuildRepo{
		guilds: map[string]*guild.Guild{
			testGuildID: {
				ID:         testGuildID,
				Gold:       testInitialGold,
				Reserved:   testInitialReserve,
				DailyLimit: testInitialDailyLimit,
				DailySpent: testInitialDailySpent,
			},
		},
	}

	// -- Order
	const (
		initialOrderID       = "o-1"
		initialOrderItemID   = "user-1"
		initialOrderSellerID = "seller-1"
		initialOrderBuyerID  = "user-2"
		initialOrderPrice    = 2000
		initialOrderStatus   = Listed
		initialListedEpoch   = 857174400

		initialOrderID2       = "o-10"
		initialOrderItemID2   = "user-10"
		initialOrderSellerID2 = "seller-10"
		initialOrderBuyerID2  = "user-20"
		initialOrderPrice2    = 4000
		initialOrderStatus2   = Listed
		initialListedEpoch2   = 857174422
	)

	oRepo := &MockOrderRepo{
		orders: map[string]*LimitOrder{
			//initialOrderID: {
			//	ID:       initialOrderID,
			//	ItemID:   initialOrderItemID,
			//	SellerID: initialOrderSellerID,
			//	BuyerID:  initialOrderBuyerID,
			//	Price:    initialOrderPrice,
			//	Status:   initialOrderStatus,
			//	ListedAt: time.Unix(initialListedEpoch, 0).UTC(),
			//},
			//initialOrderID2: {
			//	ID:       initialOrderID2,
			//	ItemID:   initialOrderItemID2,
			//	SellerID: initialOrderSellerID2,
			//	BuyerID:  initialOrderBuyerID2,
			//	Price:    initialOrderPrice2,
			//	Status:   initialOrderStatus2,
			//	ListedAt: time.Unix(initialListedEpoch2, 0).UTC(),
			//},
		},
	}

	orders := map[string]*LimitOrder{
		initialOrderID: {
			ID:       initialOrderID,
			ItemID:   initialOrderItemID,
			SellerID: initialOrderSellerID,
			BuyerID:  initialOrderBuyerID,
			Price:    initialOrderPrice,
			Status:   initialOrderStatus,
			ListedAt: time.Unix(initialListedEpoch, 0).UTC(),
		},
		initialOrderID2: {
			ID:       initialOrderID2,
			ItemID:   initialOrderItemID2,
			SellerID: initialOrderSellerID2,
			BuyerID:  initialOrderBuyerID2,
			Price:    initialOrderPrice2,
			Status:   initialOrderStatus2,
			ListedAt: time.Unix(initialListedEpoch2, 0).UTC(),
		},
	}

	svc := NewOrderService(oRepo, gRepo)

	// Act
	for _, o := range orders {
		listedO, err := svc.List(ctx, o.ItemID, o.SellerID, o.Price)
		if err != nil {
			t.Fatal(err)
		}
		// Assert
		if listedO.ItemID != listedO.ItemID {
			t.Errorf("listed ItemID: %s, doesn't match: %s", listedO.ItemID, o.ItemID)
		}

	}
}

func TestService_Buy(t *testing.T) {
	ctx := context.Background()

	const (
		testGuildID           = "guild-1"
		testInitialGold       = 200
		testInitialReserve    = 100
		testInitialDailyLimit = 200
		testInitialDailySpent = 0
	)

	gRepo := &MockGuildRepo{
		guilds: map[string]*guild.Guild{
			testGuildID: {
				ID:         testGuildID,
				Gold:       testInitialGold,
				Reserved:   testInitialReserve,
				DailyLimit: testInitialDailyLimit,
				DailySpent: testInitialDailySpent,
			},
		},
	}

	const (
		initialOrderID       = "o-1"
		initialOrderItemID   = "user-1"
		initialOrderSellerID = "seller-1"
		initialOrderBuyerID  = "user-2"
		initialOrderPrice    = 2000
		initialOrderStatus   = Listed
		initialListedEpoch   = 857174400

		initialOrderID2       = "o-10"
		initialOrderItemID2   = "user-10"
		initialOrderSellerID2 = "seller-10"
		initialOrderBuyerID2  = "user-20"
		initialOrderPrice2    = 4000
		initialOrderStatus2   = Listed
		initialListedEpoch2   = 857174422
	)

	oRepo := &MockOrderRepo{
		orders: map[string]*LimitOrder{
			initialOrderID: {
				ID:       initialOrderID,
				ItemID:   initialOrderItemID,
				SellerID: initialOrderSellerID,
				BuyerID:  initialOrderBuyerID,
				Price:    initialOrderPrice,
				Status:   initialOrderStatus,
				ListedAt: time.Unix(initialListedEpoch, 0).UTC(),
			},
			initialOrderID2: {
				ID:       initialOrderID2,
				ItemID:   initialOrderItemID2,
				SellerID: initialOrderSellerID2,
				BuyerID:  initialOrderBuyerID2,
				Price:    initialOrderPrice2,
				Status:   initialOrderStatus2,
				ListedAt: time.Unix(initialListedEpoch2, 0).UTC(),
			},
		},
	}

	svc := NewOrderService(oRepo, gRepo)

	err := svc.Buy(ctx, initialOrderID, initialOrderBuyerID)
	if err != nil {
		t.Errorf("couldn't buy order id: %s, for buyer id: %s", initialOrderID, initialOrderBuyerID)
	}
}

func TestService_Cancel(t *testing.T) {
	ctx := context.Background()

	const (
		testGuildID           = "guild-1"
		testInitialGold       = 200
		testInitialReserve    = 100
		testInitialDailyLimit = 200
		testInitialDailySpent = 0
	)

	gRepo := &MockGuildRepo{
		guilds: map[string]*guild.Guild{
			testGuildID: {
				ID:         testGuildID,
				Gold:       testInitialGold,
				Reserved:   testInitialReserve,
				DailyLimit: testInitialDailyLimit,
				DailySpent: testInitialDailySpent,
			},
		},
	}

	const (
		initialOrderID       = "o-1"
		initialOrderItemID   = "user-1"
		initialOrderSellerID = "seller-1"
		initialOrderBuyerID  = "user-2"
		initialOrderPrice    = 2000
		initialOrderStatus   = Listed
		initialListedEpoch   = 857174400

		initialOrderID2       = "o-10"
		initialOrderItemID2   = "user-10"
		initialOrderSellerID2 = "seller-10"
		initialOrderBuyerID2  = "user-20"
		initialOrderPrice2    = 4000
		initialOrderStatus2   = Listed
		initialListedEpoch2   = 857174422
	)

	oRepo := &MockOrderRepo{
		orders: map[string]*LimitOrder{
			initialOrderID: {
				ID:       initialOrderID,
				ItemID:   initialOrderItemID,
				SellerID: initialOrderSellerID,
				BuyerID:  initialOrderBuyerID,
				Price:    initialOrderPrice,
				Status:   initialOrderStatus,
				ListedAt: time.Unix(initialListedEpoch, 0).UTC(),
			},
			initialOrderID2: {
				ID:       initialOrderID2,
				ItemID:   initialOrderItemID2,
				SellerID: initialOrderSellerID2,
				BuyerID:  initialOrderBuyerID2,
				Price:    initialOrderPrice2,
				Status:   initialOrderStatus2,
				ListedAt: time.Unix(initialListedEpoch2, 0).UTC(),
			},
		},
	}

	svc := NewOrderService(oRepo, gRepo)

	err := svc.Cancel(ctx, initialOrderID, initialOrderBuyerID)
	if err != nil {
		t.Errorf("couldn't cancel order id: %s, for buyer id: %s", initialOrderID, initialOrderBuyerID)
	}
}
