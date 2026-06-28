package item

type Type string

const (
	Common    Type = "common"
	Rare      Type = "rare"
	Legendary Type = "legendary"
)

type Item struct {
	ID        string
	Name      string
	Type      Type
	OwnerID   string // guild ID
	Available bool
	BasePrice float64
}
