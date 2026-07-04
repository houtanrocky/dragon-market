package item

import (
	"context"
	"fmt"
)

//type ItemService interface {
//	TransferOwnership(ctx context.Context, itemId string, sellerID string, buyerID string) error
//	UpdateItem(ctx context.Context, item *Item) error
//}

type ItemServiceImpl struct {
	itemRepository ItemRepository
}

func NewItemService(r ItemRepository) *ItemServiceImpl {
	return &ItemServiceImpl{itemRepository: r}
}

func (s *ItemServiceImpl) GetItem(ctx context.Context, id string) (*Item, error) {
	repo := s.itemRepository
	it, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func (s *ItemServiceImpl) GetItemForUpdate(
	ctx context.Context,
	itemID string,
) (*Item, error) {
	return s.itemRepository.GetItemForUpdate(ctx, itemID)
}

func (s *ItemServiceImpl) ListFree(ctx context.Context) ([]*Item, error) {
	repo := s.itemRepository

	av, err := repo.ListFree(ctx)
	if err != nil {
		return nil, err
	}

	return av, nil
}

func (s *ItemServiceImpl) ReleaseFromAuction(ctx context.Context, id string) error {
	return s.itemRepository.ReleaseFromAuction(ctx, id)
}

func (s *ItemServiceImpl) MarkListedInAuction(ctx context.Context, id, sellerID string) error {
	return s.itemRepository.MarkListedInAuction(ctx, id, sellerID)
}

func (s *ItemServiceImpl) TransferFromAuction(ctx context.Context, itemID, sellerID, winnerID string) error {
	if winnerID == "" {
		return fmt.Errorf("winner ID is required")
	}
	return s.itemRepository.TransferFromAuction(ctx, itemID, sellerID, winnerID)
}

func (s *ItemServiceImpl) UpdateItem(ctx context.Context, item *Item) error {
	return s.itemRepository.Update(ctx, item)
}

func (s *ItemServiceImpl) TransferOwnership(
	ctx context.Context,
	itemID, sellerID, buyerID string,
) error {
	if buyerID == "" {
		return fmt.Errorf("buyer ID is required")
	}

	return s.itemRepository.TransferFromOrder(
		ctx,
		itemID,
		sellerID,
		buyerID,
	)
}
