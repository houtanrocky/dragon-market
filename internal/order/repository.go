package order

import "context"

type OrderRepository interface {
	Create(ctx context.Context, o *LimitOrder) error
	GetByID(ctx context.Context, id string) (*LimitOrder, error)
	Update(ctx context.Context, o *LimitOrder) error
	GetOrdersByItemIDAndStatus(ctx context.Context, itemID string, status Status) ([]*LimitOrder, error)
}
