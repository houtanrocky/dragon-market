package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"market-dragon/internal/gold"
	"market-dragon/internal/item"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetByID(ctx context.Context, id string) (*item.Item, error) {
	q := r.itemConn(ctx)
	var row *sql.Row
	row = q.QueryRowContext(ctx, `SELECT id, name, type, owner_id, status, base_price 
		FROM items WHERE id = $1`, id)

	var it item.Item
	err := row.Scan(&it.ID, &it.Name, &it.Type, &it.OwnerID, &it.Status, &it.BasePrice)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, item.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &it, nil
}

func (r *ItemRepository) Update(ctx context.Context, i *item.Item) error {
	q := r.itemConn(ctx)
	_, err := q.ExecContext(ctx, `UPDATE items 
		SET name = $1, type = $2, owner_id = $3, status = $4, base_price = $5
		WHERE id = $6
		`, i.Name, i.Type, i.OwnerID, i.Status, i.BasePrice, i.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) ListFree(ctx context.Context) ([]*item.Item, error) {
	q := r.itemConn(ctx)

	rows, err := q.QueryContext(ctx, `SELECT id, name, type, owner_id, status, base_price 
		FROM items WHERE status = 'free'`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var items []*item.Item
	for rows.Next() {
		var it item.Item
		err := rows.Scan(&it.ID, &it.Name, &it.Type, &it.OwnerID, &it.Status, &it.BasePrice)
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

func (r *ItemRepository) ListItemIDs(ctx context.Context) ([]string, error) {
	rows, err := r.itemConn(ctx).QueryContext(ctx, `SELECT id FROM items ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *ItemRepository) UpdateBasePrice(ctx context.Context, id string, price gold.Amount) error {
	if price <= 0 {
		return fmt.Errorf("base price must be positive")
	}
	result, err := r.itemConn(ctx).ExecContext(ctx,
		`UPDATE items SET base_price = $2 WHERE id = $1`, id, price)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return item.ErrItemNotFound
	}
	return nil
}

func (r *ItemRepository) MarkListedInAuction(ctx context.Context, id, sellerID string) error {
	result, err := r.itemConn(ctx).ExecContext(ctx, `UPDATE items
		SET status = 'listed_in_auction'
		WHERE id = $1 AND owner_id = $2 AND type = 'legendary' AND status = 'free'`, id, sellerID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("item cannot be listed in auction")
	}
	return nil
}

func (r *ItemRepository) ReleaseFromAuction(ctx context.Context, id string) error {
	result, err := r.itemConn(ctx).ExecContext(ctx, `UPDATE items SET status = 'free'
		WHERE id = $1 AND status = 'listed_in_auction'`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("item cannot be released from auction")
	}
	return nil
}

func (r *ItemRepository) TransferFromAuction(ctx context.Context, id, sellerID, winnerID string) error {
	result, err := r.itemConn(ctx).ExecContext(ctx, `UPDATE items
		SET owner_id = $3, status = 'free'
		WHERE id = $1 AND owner_id = $2 AND status = 'listed_in_auction'`, id, sellerID, winnerID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("item cannot be transferred from auction")
	}
	return nil
}

func (r *ItemRepository) GetItemForUpdate(ctx context.Context, itemID string) (*item.Item, error) {
	q := r.itemConn(ctx)
	row := q.QueryRowContext(ctx, `SELECT id, name, type, owner_id, status, base_price 
		FROM items WHERE id = $1 FOR UPDATE`, itemID)

	var it item.Item
	err := row.Scan(&it.ID, &it.Name, &it.Type, &it.OwnerID, &it.Status, &it.BasePrice)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, item.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &it, nil
}

func (r *ItemRepository) TransferFromOrder(
	ctx context.Context,
	itemID, sellerID, buyerID string,
) error {
	result, err := r.itemConn(ctx).ExecContext(ctx, `
                UPDATE items
                SET owner_id = $3, status = 'free'
                WHERE id = $1
                  AND owner_id = $2
                  AND status = 'listed_in_order'`,
		itemID,
		sellerID,
		buyerID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("item cannot be transferred from order")
	}

	return nil
}

func (r *ItemRepository) Create(ctx context.Context, it *item.Item) error {
	q := r.itemConn(ctx)
	execContext, err := q.ExecContext(ctx, `INSERT INTO items
	    (id, name, type, owner_id, base_price, status)
		VALUES($1, $2, $3, $4, $5, $6)`, it.ID, it.Name, it.Type, it.OwnerID, it.BasePrice, it.Status)
	if err != nil {
		return err
	}

	affected, err := execContext.RowsAffected()
	if err != nil {
		return err
	}

	if affected != 1 {
		return fmt.Errorf("expected one inserted item, got %d", affected)
	}

	return nil
}

func (r *ItemRepository) itemConn(ctx context.Context) querier {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return r.db
}
