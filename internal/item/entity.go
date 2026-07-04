package item

import "market-dragon/internal/gold"

type Type string

const (
	Common    Type = "common"
	Rare      Type = "rare"
	Legendary Type = "legendary"
)

type Status string

const (
	Free            Status = "free"
	ListedInOrder   Status = "listed_in_order"
	ListedInAuction Status = "listed_in_auction"
)

type Item struct {
	ID        string
	Name      string
	Type      Type
	OwnerID   string // guild ID
	Status    Status
	BasePrice gold.Amount
}
