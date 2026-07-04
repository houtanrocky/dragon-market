package guild

import "market-dragon/internal/gold"

type Guild struct {
	ID         string
	Gold       gold.Amount
	Reserved   gold.Amount
	DailyLimit gold.Amount // max gold committable per day
	DailySpent gold.Amount // gold committed today (purchases + reservations)
}
