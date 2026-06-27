package wallet

import (
	"context"
	"market-dragon/internal/domain/wallet/models"
)

type GuildRepository interface {
	Get(ctx context.Context, id string) (*models.Guild, error)
	Update(ctx context.Context, val *models.Guild) error
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
