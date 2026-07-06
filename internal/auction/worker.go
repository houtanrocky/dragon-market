package auction

import (
	"context"
	"log/slog"
	"time"
)

// RunSettlementWorker periodically settles expired auctions. Database row
// locks and status checks make concurrent worker instances safe.
func RunSettlementWorker(ctx context.Context, service *AuctionServiceImpl, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := service.EndExpiredAuctions(ctx, 100); err != nil && ctx.Err() == nil {
			slog.Error("settle expired auctions", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
