package wallet

import (
	"context"
	"market-dragon/internal/domain/wallet/models"
)

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

type PostgresRepository struct {
}

func (p PostgresRepository) Get(ctx context.Context, id string) (*models.Guild, error) {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) Update(ctx context.Context, val *models.Guild) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	//TODO implement me
	panic("implement me")
}
