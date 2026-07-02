package item

import (
	"context"
)

type ItemService struct {
	itemRepository Repository
}

func NewItemService(r Repository) *ItemService {
	return &ItemService{itemRepository: r}
}

func (s *ItemService) Get(ctx context.Context, id string) (*Item, error) {
	repo := s.itemRepository
	it, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func (s *ItemService) ListAvailable(ctx context.Context) ([]*Item, error) {
	repo := s.itemRepository

	av, err := repo.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	return av, nil
}
