package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"market-dragon/internal/auction"
)

type AuctionRepository struct {
	db *sql.DB
}

func NewAuctionRepository(db *sql.DB) *AuctionRepository {
	return &AuctionRepository{db: db}
}

func (r *AuctionRepository) CreateAuction(ctx context.Context, a *auction.Auction) error {
	_, err := r.conn(ctx).ExecContext(ctx, `INSERT INTO auctions
		(id, item_id, seller_id, ends_at, status) VALUES ($1, $2, $3, $4, $5)`,
		a.ID, a.ItemID, a.SellerID, a.EndsAt, a.Status)
	return err
}

func (r *AuctionRepository) GetAuctionByID(ctx context.Context, id string) (*auction.Auction, error) {
	query := `SELECT id, item_id, seller_id, ends_at, status FROM auctions WHERE id = $1`
	if getTx(ctx) != nil {
		query += ` FOR UPDATE`
	}
	var a auction.Auction
	err := r.conn(ctx).QueryRowContext(ctx, query, id).Scan(&a.ID, &a.ItemID, &a.SellerID, &a.EndsAt, &a.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, auction.ErrAuctionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AuctionRepository) GetActiveAuctionByItemID(ctx context.Context, itemID string) (*auction.Auction, error) {
	var a auction.Auction
	err := r.conn(ctx).QueryRowContext(ctx, `SELECT id, item_id, seller_id, ends_at, status
		FROM auctions WHERE item_id = $1 AND status = 'active'`, itemID).
		Scan(&a.ID, &a.ItemID, &a.SellerID, &a.EndsAt, &a.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, auction.ErrAuctionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AuctionRepository) ListExpiredActiveAuctionIDs(ctx context.Context, now time.Time, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.conn(ctx).QueryContext(ctx, `SELECT id FROM auctions
		WHERE status = 'active' AND ends_at <= $1 ORDER BY ends_at LIMIT $2`, now, limit)
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

func (r *AuctionRepository) ExtendActiveAuction(ctx context.Context, id string, endsAt time.Time) error {
	return expectOne(r.conn(ctx).ExecContext(ctx,
		`UPDATE auctions SET ends_at = $2 WHERE id = $1 AND status = 'active'`, id, endsAt))
}

func (r *AuctionRepository) EndActiveAuction(ctx context.Context, id string) error {
	return expectOne(r.conn(ctx).ExecContext(ctx,
		`UPDATE auctions SET status = 'ended' WHERE id = $1 AND status = 'active'`, id))
}

func (r *AuctionRepository) CreateBid(ctx context.Context, b *auction.Bid) (*auction.Bid, error) {
	rows := r.conn(ctx).QueryRowContext(ctx, `INSERT INTO bids
		(id, auction_id, bidder_id, amount, placed_at, status) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, auction_id, bidder_id, amount, placed_at, status`,
		b.ID, b.AuctionID, b.BidderID, b.Amount, b.PlacedAt, b.Status)

	var created auction.Bid
	err := rows.Scan(&created.ID, &created.AuctionID, &created.BidderID, &created.Amount, &created.PlacedAt, &created.Status)
	if err != nil {
		return nil, err
	}

	return &created, err
}

func (r *AuctionRepository) GetBidByID(ctx context.Context, id string) (*auction.Bid, error) {
	var b auction.Bid
	err := r.conn(ctx).QueryRowContext(ctx, `SELECT id, auction_id, bidder_id, amount, placed_at, status
		FROM bids WHERE id = $1`, id).
		Scan(&b.ID, &b.AuctionID, &b.BidderID, &b.Amount, &b.PlacedAt, &b.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, auction.ErrBidNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *AuctionRepository) GetTopActiveBid(ctx context.Context, auctionID string) (*auction.Bid, error) {
	var b auction.Bid
	err := r.conn(ctx).QueryRowContext(ctx, `SELECT id, auction_id, bidder_id, amount, placed_at, status
		FROM bids WHERE auction_id = $1 AND status = 'active'
		ORDER BY amount DESC, placed_at ASC, id ASC LIMIT 1`, auctionID).
		Scan(&b.ID, &b.AuctionID, &b.BidderID, &b.Amount, &b.PlacedAt, &b.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, auction.ErrBidNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *AuctionRepository) MarkBidOutbid(ctx context.Context, id string) error {
	return expectOne(r.conn(ctx).ExecContext(ctx,
		`UPDATE bids SET status = 'outbid' WHERE id = $1 AND status = 'active'`, id))
}

func (r *AuctionRepository) CancelActiveBid(ctx context.Context, auctionID, bidID, bidderID string) error {
	result, err := r.conn(ctx).ExecContext(ctx, `UPDATE bids SET status = 'cancelled'
		WHERE id = $1 AND auction_id = $2 AND bidder_id = $3 AND status = 'active'`,
		bidID, auctionID, bidderID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return auction.ErrBidNotCancellable
	}
	return nil
}

func (r *AuctionRepository) MarkBidWinning(ctx context.Context, id string) error {
	return expectOne(r.conn(ctx).ExecContext(ctx,
		`UPDATE bids SET status = 'winning' WHERE id = $1 AND status = 'active'`, id))
}

func (r *AuctionRepository) conn(ctx context.Context) querier {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

func expectOne(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected one affected row, got %d", rows)
	}
	return nil
}
