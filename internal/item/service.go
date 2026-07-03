package item

import (
	"context"
)

type ItemService interface {
	Get(ctx context.Context, id string) (*Item, error)
	ListFree(ctx context.Context) ([]*Item, error)
}

type ItemServiceImpl struct {
	itemRepository ItemRepository
}

func NewItemService(r ItemRepository) *ItemServiceImpl {
	return &ItemServiceImpl{itemRepository: r}
}

func (s *ItemServiceImpl) Get(ctx context.Context, id string) (*Item, error) {
	repo := s.itemRepository
	it, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func (s *ItemServiceImpl) ListFree(ctx context.Context) ([]*Item, error) {
	repo := s.itemRepository

	av, err := repo.ListFree(ctx)
	if err != nil {
		return nil, err
	}

	return av, nil
}
