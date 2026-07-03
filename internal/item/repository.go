package item

import "context"

type ItemRepository interface {
	GetByID(ctx context.Context, id string) (*Item, error)
	Update(ctx context.Context, item *Item) error
	ListFree(ctx context.Context) ([]*Item, error)
}
