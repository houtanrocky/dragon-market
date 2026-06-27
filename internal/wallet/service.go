package wallet

import (
	"fmt"
	"market-dragon/internal/models"
)

type GuildRepository interface {
	Get(id string) (*models.Guild, error)
	Update(val *models.Guild) (*models.Guild, error)
}

type Service struct {
	guildRepository GuildRepository
}

func (s *Service) Reserve(id string, amount float64) error {
	repo := s.guildRepository
	g, err := repo.Get(id)
	if err != nil {
		return err
	}

	enough := g.Gold-g.Reserved >= amount
	if !enough {
		return fmt.Errorf("insufficient reserve: %v, amount: %v", g.Reserved, amount)
	}
	g.Reserved += amount
	_, err = repo.Update(g)
	if err != nil {
		return err
	}

	return nil
}

func NewWalletService(r GuildRepository) *Service {
	return &Service{guildRepository: r}
}
