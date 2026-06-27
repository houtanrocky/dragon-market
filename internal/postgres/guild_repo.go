package postgres

import (
	"context"
	"database/sql"
	"market-dragon/internal/guild"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (p *Repository) Get(ctx context.Context, id string) (*guild.Guild, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Repository) Update(ctx context.Context, val *guild.Guild) error {
	//TODO implement me
	panic("implement me")
}

func (p *Repository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	//TODO implement me
	panic("implement me")
}
