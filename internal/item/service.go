package item

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"market-dragon/internal/gold"
)

//type ItemService interface {
//	TransferOwnership(ctx context.Context, itemId string, sellerID string, buyerID string) error
//	UpdateItem(ctx context.Context, item *Item) error
//}

type ItemServiceImpl struct {
	itemRepository ItemRepository
	ownerChecker   OwnerChecker
}

type OwnerChecker interface {
	GuildExists(ctx context.Context, guildID string) (bool, error)
}

func NewItemService(r ItemRepository, ownerChecker OwnerChecker) *ItemServiceImpl {
	return &ItemServiceImpl{
		itemRepository: r,
		ownerChecker:   ownerChecker,
	}
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

func (s *ItemServiceImpl) Create(ctx context.Context, name string, typ Type, ownerID string, basePrice gold.Amount) (*Item, error) {
	name = strings.TrimSpace(name)
	ownerID = strings.TrimSpace(ownerID)
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidItem)
	}
	if !typ.IsValid() {
		return nil, fmt.Errorf("%w: invalid item type: %q", ErrInvalidItem, typ)
	}
	if ownerID == "" {
		return nil, fmt.Errorf("%w: owner ID cannot be empty", ErrInvalidItem)
	}
	if basePrice <= 0 {
		return nil, fmt.Errorf("%w: base price must be positive", ErrInvalidItem)
	}

	exists, err := s.ownerChecker.GuildExists(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("check owner guild: %w", err)
	}
	if !exists {
		return nil, ErrOwnerNotFound
	}

	it := &Item{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      typ,
		OwnerID:   ownerID,
		Status:    Free,
		BasePrice: basePrice,
	}
	if err := s.itemRepository.Create(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}
