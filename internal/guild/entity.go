package guild

import (
	"market-dragon/internal/gold"
	"time"
)

type Guild struct {
	ID         string
	Gold       gold.Amount
	Reserved   gold.Amount
	DailyLimit gold.Amount // max gold committable per day
	DailySpent gold.Amount // gold committed today (purchases + reservations)
	SpentOn    time.Time   // UTC calendar day represented by DailySpent
}
