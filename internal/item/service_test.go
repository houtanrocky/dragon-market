package item

import (
	"context"
	"testing"
)

type MockItemRepository struct {
	items map[string]*Item
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
	svc := NewItemService(repo)

	i, err := svc.Get(ctx, testInitialItemID)
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
	svc := NewItemService(repo)

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
