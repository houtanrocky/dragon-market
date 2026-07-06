package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type IdempotencyRecord struct {
	Key         string
	Operation   string
	RequestHash string
	StatusCode  *int
	Response    json.RawMessage
	CreatedAt   time.Time
	CompletedAt *time.Time
}

var (
	ErrKeyConflict  = errors.New("idempotency key reused with different request")
	ErrNotCompleted = errors.New("idempotent operation is not completed")
)

type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(context.Context) error) error
}

type IdempotencyService struct {
	repo IdempotencyRepository
	tx   Transactor
}

func NewService(
	repo IdempotencyRepository,
	tx Transactor,
) *IdempotencyService {
	return &IdempotencyService{
		repo: repo,
		tx:   tx,
	}
}

type Result struct {
	StatusCode int
	Response   json.RawMessage
	Replayed   bool
}

type Operation func(ctx context.Context) (statusCode int, response json.RawMessage, err error)

func (s *IdempotencyService) Execute(
	ctx context.Context,
	key string,
	operation string,
	requestHash string,
	run Operation,
) (Result, error) {
	var result Result

	err := s.tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		record, claimed, err := s.repo.Claim(txCtx, key, operation, requestHash)
		if err != nil {
			return err
		}

		if !claimed {
			if record == nil || !record.Completed() || record.StatusCode == nil {
				return ErrNotCompleted
			}

			result = Result{
				StatusCode: *record.StatusCode,
				Response:   record.Response,
				Replayed:   true,
			}
			return nil
		}

		statusCode, response, err := run(txCtx)
		if err != nil {
			return err
		}

		if err := s.repo.Complete(txCtx, key, statusCode, response); err != nil {
			return err
		}

		result = Result{
			StatusCode: statusCode,
			Response:   response,
			Replayed:   false,
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

func (r IdempotencyRecord) Completed() bool {
	return r.CompletedAt != nil
}
