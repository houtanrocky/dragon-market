package item

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type MockItemRepository struct {
	items map[string]*Item
}

type MockOwnerChecker struct {
	exists bool
	err    error
}

func (f MockOwnerChecker) GuildExists(context.Context, string) (bool, error) {
	return f.exists, f.err
}

func (r MockItemRepository) Create(ctx context.Context, it *Item) error {
	r.items[it.ID] = it
	return nil
}

func (r MockItemRepository) GetItemForUpdate(ctx context.Context, id string) (*Item, error) {
	return r.GetByID(ctx, id)
}

func (r MockItemRepository) TransferFromOrder(ctx context.Context, itemID, sellerID, buyerID string) error {
	if r.items[itemID].OwnerID != sellerID {
		return fmt.Errorf("cannot sell item not owned")
	}
	if buyerID == "" {
		return fmt.Errorf("buyer empty")
	}
	r.items[itemID].OwnerID = buyerID

	return nil
}

func (r MockItemRepository) GetByID(ctx context.Context, id string) (*Item, error) {
	i, ok := r.items[id]
	if !ok {
		return nil, ErrItemNotFound
	}
	return i, nil
}

func (r MockItemRepository) Update(ctx context.Context, i *Item) error {
	r.items[i.ID] = i
	return nil
}

func (r MockItemRepository) ListFree(ctx context.Context) ([]*Item, error) {
	var result []*Item
	for _, it := range r.items {
		if it.Status == Free {
			result = append(result, it)
		}
	}
	return result, nil
}

func (r MockItemRepository) MarkListedInAuction(ctx context.Context, id, sellerID string) error {
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	item.Status = ListedInAuction
	return nil
}

func (r MockItemRepository) ReleaseFromAuction(ctx context.Context, id string) error {
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	item.Status = Free
	return nil
}

func (r MockItemRepository) TransferFromAuction(ctx context.Context, id, sellerID, winnerID string) error {
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	item.OwnerID, item.Status = winnerID, Free
	return nil
}

// ---------------------------- Tests ----------
func TestItemService_Create(t *testing.T) {
	const (
		testInitialItemID    = "item-1"
		testInitialItemName  = "Sussy Sword"
		testInitialItemType  = Common
		testInitialOwnerId   = "guild-1"
		testInitialStatus    = Free
		testInitialBasePrice = 1000
	)

	repo := &MockItemRepository{items: map[string]*Item{
		testInitialItemID: {
			ID:        testInitialItemID,
			Name:      testInitialItemName,
			Type:      testInitialItemType,
			OwnerID:   testInitialOwnerId,
			Status:    testInitialStatus,
			BasePrice: testInitialBasePrice,
		},
	}}

	ctx := context.Background()
	svc := NewItemService(repo, MockOwnerChecker{exists: true})

	i, err := svc.Create(ctx, testInitialItemName, testInitialItemType, testInitialOwnerId, testInitialBasePrice)
	if err != nil {
		t.Error(err)
	}
	if i.Name != testInitialItemName {
		t.Errorf("unexpected item %v", i)
	}
}

func TestItemService_Create_OwnerNotFound(t *testing.T) {
	repo := &MockItemRepository{items: make(map[string]*Item)}
	svc := NewItemService(repo, MockOwnerChecker{exists: false})

	created, err := svc.Create(context.Background(), "Sword", Common, "missing-guild", 100)

	if !errors.Is(err, ErrOwnerNotFound) {
		t.Fatalf("expected ErrOwnerNotFound, got %v", err)
	}
	if created != nil {
		t.Fatalf("expected no item, got %v", created)
	}
}

func TestItemService_Create_OwnerLookupFailure(t *testing.T) {
	repo := &MockItemRepository{items: make(map[string]*Item)}
	lookupErr := errors.New("database unavailable")
	svc := NewItemService(repo, MockOwnerChecker{err: lookupErr})

	created, err := svc.Create(context.Background(), "Sword", Common, "guild-1", 100)

	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error to be preserved, got %v", err)
	}
	if errors.Is(err, ErrOwnerNotFound) {
		t.Fatalf("lookup failure must not be reported as owner not found")
	}
	if created != nil {
		t.Fatalf("expected no item, got %v", created)
	}
}

func TestService_Get_Success(t *testing.T) {
	const (
		testInitialItemID    = "item-1"
		testInitialItemName  = "Sussy Sword"
		testInitialItemType  = Common
		testInitialOwnerId   = "guild-1"
		testInitialStatus    = Free
		testInitialBasePrice = 1000
	)
	repo := &MockItemRepository{items: map[string]*Item{
		testInitialItemID: {
			ID:        testInitialItemID,
			Name:      testInitialItemName,
			Type:      testInitialItemType,
			OwnerID:   testInitialOwnerId,
			Status:    testInitialStatus,
			BasePrice: testInitialBasePrice,
		},
	}}

	ctx := context.Background()
	svc := NewItemService(repo, MockOwnerChecker{exists: true})

	i, err := svc.GetItem(ctx, testInitialItemID)
	if err != nil {
		t.Error(err)
	}
	if i.ID != testInitialItemID || i.Name != testInitialItemName {
		t.Errorf("unexpected item %v", i)
	}
}

func TestService_ListFree(t *testing.T) {
	// -------- Arrange ---------------------------
	initialItems := map[string]*Item{
		"item-1": {
			ID:     "item-1",
			Status: Free,
		},
		"item-2": {
			ID:     "item-2",
			Status: Free,
		},
	}

	repo := &MockItemRepository{items: initialItems}
	ctx := context.Background()
	svc := NewItemService(repo, MockOwnerChecker{exists: true})

	// -------- Act ------------
	items, err := svc.ListFree(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// -------- Assert ----------
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %v", len(items))
	}
	for _, it := range items {
		if it.Status != Free {
			t.Errorf("returned item %v is not free", it.ID)
		}
	}

}
