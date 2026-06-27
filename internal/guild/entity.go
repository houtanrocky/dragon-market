package guild

type Guild struct {
	ID         string
	Gold       float64
	Reserved   float64
	DailyLimit float64 // max gold committable per day
	DailySpent float64 // gold committed today (purchases + reservations)
}
