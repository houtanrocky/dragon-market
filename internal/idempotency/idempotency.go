package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrKeyConflict  = errors.New("idempotency key reused with different request")
	ErrNotCompleted = errors.New("idempotent operation is not completed")
)

type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(context.Context) error) error
}

type IdempotencyRecord struct {
	Key         string
	Operation   string
	RequestHash string
	StatusCode  *int
	Response    json.RawMessage
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func (r IdempotencyRecord) Completed() bool {
	return r.CompletedAt != nil
}

type IdempotencyRepository interface {
	Claim(
		ctx context.Context,
		key string,
		operation string,
		requestHash string,
	) (record IdempotencyRecord, claimed bool, err error)

	Complete(
		ctx context.Context,
		key string,
		statusCode int,
		response json.RawMessage,
	) error
}
