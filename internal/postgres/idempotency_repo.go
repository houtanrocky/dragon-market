package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"market-dragon/internal/idempotency"
)

type IdempotencyRepo struct {
	db *sql.DB
}

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

func (r *IdempotencyRepo) Claim(
	ctx context.Context,
	key string,
	operation string,
	requestHash string,
) (record *idempotency.IdempotencyRecord, claimed bool, err error) {
	q := r.conn(ctx)

	row := q.QueryRowContext(ctx, `INSERT INTO idempotency_records (key, operation, request_hash)
	  VALUES ($1, $2, $3)
	  ON CONFLICT (key) DO NOTHING
	  RETURNING
		  key,
		  operation,
		  request_hash,
		  status_code,
		  response,
		  created_at,
		  completed_at;`, key, operation, requestHash)

	var idem idempotency.IdempotencyRecord
	err = row.Scan(&idem.Key, &idem.Operation, &idem.RequestHash, &idem.StatusCode, &idem.Response, &idem.CreatedAt, &idem.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		exRow := q.QueryRowContext(ctx, `SELECT
		  key,
		  operation,
		  request_hash,
		  status_code,
		  response,
		  created_at,
		  completed_at
	  	FROM idempotency_records
  		WHERE key = $1;`, key)

		var existing idempotency.IdempotencyRecord
		err := exRow.Scan(&existing.Key, &existing.Operation, &existing.RequestHash, &existing.StatusCode, &existing.Response, &existing.CreatedAt, &existing.CompletedAt)
		if err != nil {
			return nil, false, err
		}

		if existing.Operation != operation ||
			existing.RequestHash != requestHash {
			return nil, false, idempotency.ErrKeyConflict
		}

		return &existing, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return &idem, true, nil
}

func (r *IdempotencyRepo) Complete(
	ctx context.Context,
	key string,
	statusCode int,
	response json.RawMessage,
) error {
	q := r.conn(ctx)

	res, err := q.ExecContext(ctx, `
		UPDATE idempotency_records SET
		  status_code = $2,
		  response = $3
		  completed_at = NOW()
		WHERE key = $1
		  AND completed_at IS NULL
	`, key, statusCode, response)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return idempotency.ErrNotCompleted
	}

	return nil
}

func (r *IdempotencyRepo) conn(ctx context.Context) querier {
	if tx := getTx(ctx); tx != nil {
		return tx
	}

	return r.db
}
