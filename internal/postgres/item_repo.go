package postgres

import (
	"context"
	"database/sql"
	"market-dragon/internal/item"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetByID(ctx context.Context, id string) (*item.Item, error) {
	panic("implement me")
}

func (r *ItemRepository) Update(ctx context.Context, i *item.Item) error {
	panic("implement me")
}

func (r *ItemRepository) ListAvailable(ctx context.Context) ([]*item.Item, error) {
	panic("implement me")
}
