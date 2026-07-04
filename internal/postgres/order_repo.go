package postgres

import (
	"context"
	"database/sql"
	"errors"
	"market-dragon/internal/order"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *order.LimitOrder) error {
	q := r.conn(ctx)
	_, err := q.ExecContext(ctx, `
	INSERT into limit_orders (id, item_id, seller_id, buyer_id, price, status, listed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, o.ID, o.ItemID, o.SellerID, o.BuyerID, o.Price, o.Status, o.ListedAt)

	return err
}

func (r *OrderRepository) GetByIDForUpdate(ctx context.Context, id string) (*order.LimitOrder, error) {
	q := r.conn(ctx)
	row := q.QueryRowContext(ctx, `
	SELECT id, item_id, seller_id, buyer_id, price, status, listed_at FROM limit_orders WHERE id = $1 FOR UPDATE`, id)

	var o order.LimitOrder
	var buyerID sql.NullString
	err := row.Scan(&o.ID, &o.ItemID, &o.SellerID, &buyerID, &o.Price, &o.Status, &o.ListedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, order.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if buyerID.Valid {
		o.BuyerID = &buyerID.String
	}
	return &o, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.LimitOrder, error) {
	q := r.conn(ctx)
	row := q.QueryRowContext(ctx, `
	SELECT id, item_id, seller_id, buyer_id, price, status, listed_at FROM limit_orders WHERE id = $1`, id)

	var o order.LimitOrder
	var buyerID sql.NullString
	err := row.Scan(&o.ID, &o.ItemID, &o.SellerID, &buyerID, &o.Price, &o.Status, &o.ListedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, order.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if buyerID.Valid {
		o.BuyerID = &buyerID.String
	}
	return &o, nil
}

func (r *OrderRepository) GetOrdersByItemIDAndStatus(ctx context.Context, itemID string, status order.Status) ([]*order.LimitOrder, error) {
	q := r.conn(ctx)
	rows, err := q.QueryContext(ctx, `
	SELECT id, item_id, seller_id, buyer_id, price, status, listed_at FROM limit_orders WHERE item_id = $1 AND status = $2`, itemID, status)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)
	var orders []*order.LimitOrder
	for rows.Next() {
		var o order.LimitOrder
		var buyerID sql.NullString
		err := rows.Scan(&o.ID, &o.ItemID, &o.SellerID, &buyerID, &o.Price, &o.Status, &o.ListedAt)
		if err != nil {
			return nil, err
		}
		if buyerID.Valid {
			o.BuyerID = &buyerID.String
		}
		orders = append(orders, &o)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, o *order.LimitOrder) error {
	q := r.conn(ctx)
	_, err := q.ExecContext(ctx, `UPDATE limit_orders
	SET  buyer_id=$1, price=$2, status=$3, listed_at=$4 
	WHERE id = $5`, o.BuyerID, o.Price, o.Status, o.ListedAt, o.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepository) CancelOtherListed(ctx context.Context, itemID, exceptOrderID string) error {
	q := r.conn(ctx)
	_, err := q.ExecContext(ctx, `UPDATE limit_orders
	SET status = 'canceled'
	WHERE item_id = $1
    AND id <> $2	
    AND status = 'listed';`, itemID, exceptOrderID)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepository) conn(ctx context.Context) querier {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return r.db
}
