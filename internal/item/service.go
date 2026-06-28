package item

import (
	"context"
)

type Service struct {
	itemRepository Repository
}

func NewItemService(r Repository) *Service {
	return &Service{itemRepository: r}
}

func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	repo := s.itemRepository
	it, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func (s *Service) ListAvailable(ctx context.Context) ([]*Item, error) {
	repo := s.itemRepository

	av, err := repo.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	return av, nil
}
