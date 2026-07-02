package guild

import (
	"context"
)

type GuildRepository interface {
	Get(ctx context.Context, id string) (*Guild, error)
	Update(ctx context.Context, val *Guild) error
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
