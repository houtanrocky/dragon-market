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
	var row *sql.Row
	// TODO: excessive if/else must be refactored
	if tx := getTx(ctx); tx != nil {
		row = tx.QueryRowContext(ctx, `SELECT id, name, type, owner_id, available, base_price 
		FROM items WHERE id = $1`, id)
	} else {
		row = r.db.QueryRowContext(ctx, `SELECT id, name, type, owner_id, available, base_price 
		FROM items WHERE id = $1`, id)
	}

	var it item.Item
	err := row.Scan(&it.ID, &it.Name, &it.Type, &it.OwnerID, &it.Available, &it.BasePrice)
	if err != nil {
		return nil, err
	}

	return &it, nil
}

func (r *ItemRepository) Update(ctx context.Context, i *item.Item) error {
	var err error
	if tx := getTx(ctx); tx != nil {
		_, err = tx.ExecContext(ctx, `UPDATE items 
		SET name = $1, type = $2, owner_id = $3, available = $4, base_price = $5
		WHERE id = $6
		`, i.Name, i.Type, i.OwnerID, i.Available, i.BasePrice, i.ID)
	} else {
		_, err = r.db.ExecContext(ctx, `UPDATE items 
		SET name = $1, type = $2, owner_id = $3, available = $4, base_price = $5
		WHERE id = $6
		`, i.Name, i.Type, i.OwnerID, i.Available, i.BasePrice, i.ID)
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) ListAvailable(ctx context.Context) ([]*item.Item, error) {
	var rows *sql.Rows
	var err error

	if tx := getTx(ctx); tx != nil {
		rows, err = tx.QueryContext(ctx, `SELECT id, name, type, owner_id, available, base_price 
		FROM items WHERE available = true`)
	} else {
		rows, err = r.db.QueryContext(ctx, `SELECT id, name, type, owner_id, available, base_price 
		FROM items WHERE available = true`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*item.Item
	for rows.Next() {
		var it item.Item
		err := rows.Scan(&it.ID, &it.Name, &it.Type, &it.OwnerID, &it.Available, &it.BasePrice)
		if err != nil {
			return nil, err
		}
		items = append(items, &it)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
