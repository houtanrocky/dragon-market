package guild

type Guild struct {
	ID         string
	Gold       int64
	Reserved   int64
	DailyLimit int64 // max gold committable per day
	DailySpent int64 // gold committed today (purchases + reservations)
}
