package item

import "context"

type Repository interface {
	GetByID(ctx context.Context, id string) (*Item, error)
	Update(ctx context.Context, item *Item) error
	ListAvailable(ctx context.Context) ([]*Item, error)
}
