package postgres

import (
	"context"
	"database/sql"
	"errors"
	"market-dragon/internal/guild"
)

type Repository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GuildExists(ctx context.Context, guildID string) (bool, error) {
	var exists bool
	err := r.guildConn(ctx).QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM guilds WHERE id = $1
		)`, guildID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*guild.Guild, error) {
	q := r.guildConn(ctx)
	row := q.QueryRowContext(ctx, `SELECT id, gold, reserved, daily_limit, daily_spent FROM guilds WHERE id = $1 FOR UPDATE`, id)

	var g guild.Guild
	err := row.Scan(&g.ID, &g.Gold, &g.Reserved, &g.DailyLimit, &g.DailySpent)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, guild.ErrGuildNotFound
	}
	if err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *Repository) Update(ctx context.Context, val *guild.Guild) error {
	q := r.guildConn(ctx)
	_, err := q.ExecContext(ctx,
		`UPDATE guilds SET gold = $1, reserved = $2, daily_limit = $3, daily_spent = $4 
              WHERE id = $5`, val.Gold, val.Reserved, val.DailyLimit, val.DailySpent, val.ID)

	return err
}

func (r *Repository) guildConn(ctx context.Context) querier {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return r.db
}
