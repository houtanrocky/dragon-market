package idempotency

import (
	"context"
	"encoding/json"
)

type IdempotencyRepository interface {
	Claim(
		ctx context.Context,
		key string,
		operation string,
		requestHash string,
	) (record *IdempotencyRecord, claimed bool, err error)

	Complete(
		ctx context.Context,
		key string,
		statusCode int,
		response json.RawMessage,
	) error
}
