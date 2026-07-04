package item

import "context"

type ItemRepository interface {
	GetByID(ctx context.Context, id string) (*Item, error)
	GetItemForUpdate(ctx context.Context, id string) (*Item, error)
	Update(ctx context.Context, item *Item) error
	ListFree(ctx context.Context) ([]*Item, error)
	MarkListedInAuction(ctx context.Context, id, sellerID string) error
	ReleaseFromAuction(ctx context.Context, id string) error
	TransferFromAuction(ctx context.Context, id, sellerID, winnerID string) error
	TransferFromOrder(ctx context.Context, itemID, sellerID, buyerID string) error
}
