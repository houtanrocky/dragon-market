package wallet

import (
	"context"
	"fmt"
	"market-dragon/internal/models"
)

type GuildRepository interface {
	Get(ctx context.Context, id string) (*models.Guild, error)
	Update(ctx context.Context, val *models.Guild) (*models.Guild, error)
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	guildRepository GuildRepository
}

func NewWalletService(r GuildRepository) *Service {
	return &Service{guildRepository: r}
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

func (s *Service) Deduct(id string, amount float64) error {
	repo := s.guildRepository
	g, err := repo.Get(id)
	if err != nil {
		return err
	}

	available := g.Gold - g.Reserved
	if available < amount {
		return fmt.Errorf("insufficient available balance: have: %v need: %v", available, amount)
	}

	g.Gold -= amount
	_, err = repo.Update(g)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Release(id string, amount float64) error {
	repo := s.guildRepository
	g, err := repo.Get(id)
	if err != nil {
		return err
	}

	enoughReserve := g.Reserved >= amount
	if !enoughReserve {
		return fmt.Errorf("insufficient available release: have %v need: %v", g.Reserved, amount)
	}

	g.Reserved -= amount
	_, err = repo.Update(g)
	if err != nil {
		return err
	}

	return nil
}
