package guild

import (
	"context"
	"errors"
	"fmt"
)

type WalletService struct {
	guildRepository Repository
}

func NewWalletService(r Repository) *WalletService {
	return &WalletService{guildRepository: r}
}

func (s *WalletService) Reserve(ctx context.Context, id string, amount float64) error {
	return s.guildRepository.RunInTransaction(ctx, func(ctx context.Context) error {
		repo := s.guildRepository
		g, err := repo.Get(ctx, id)
		if errors.Is(err, ErrGuildNotFound) {
			return fmt.Errorf("guild %s does not exist: %w", id, err)
		}
		if err != nil {
			return err
		}

		enough := g.Gold-g.Reserved >= amount
		if !enough {
			return fmt.Errorf("insufficient reserve: %v, amount: %v", g.Reserved, amount)
		}
		g.Reserved += amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletService) Deduct(ctx context.Context, id string, amount float64) error {
	return s.guildRepository.RunInTransaction(ctx, func(ctx context.Context) error {
		repo := s.guildRepository
		g, err := repo.Get(ctx, id)
		if errors.Is(err, ErrGuildNotFound) {
			return fmt.Errorf("guild %s does not exist: %w", id, err)
		}
		if err != nil {
			return err
		}

		available := g.Gold - g.Reserved
		if available < amount {
			return fmt.Errorf("insufficient available balance: have: %v need: %v", available, amount)
		}

		g.Gold -= amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletService) Release(ctx context.Context, id string, amount float64) error {
	return s.guildRepository.RunInTransaction(ctx, func(ctx context.Context) error {
		repo := s.guildRepository
		g, err := repo.Get(ctx, id)
		if errors.Is(err, ErrGuildNotFound) {
			return fmt.Errorf("guild %s does not exist: %w", id, err)
		}
		if err != nil {
			return err
		}

		enoughReserve := g.Reserved >= amount
		if !enoughReserve {
			return fmt.Errorf("insufficient available release: have %v need: %v", g.Reserved, amount)
		}

		g.Reserved -= amount
		return repo.Update(ctx, g)
	})
}
