package guild

import (
	"context"
	"errors"
	"fmt"
	"market-dragon/internal/gold"
)

type Transactor interface {
	RunInTransaction(
		ctx context.Context,
		fn func(context.Context) error,
	) error
}

type WalletServiceImpl struct {
	guildRepository GuildRepository
	tx              Transactor
}

func NewWalletService(r GuildRepository, tx Transactor) *WalletServiceImpl {
	return &WalletServiceImpl{guildRepository: r, tx: tx}
}

func (s *WalletServiceImpl) Reserve(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
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
			return fmt.Errorf("insufficient reserve: %v, amount: %v", available, amount)
		}
		if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
			return fmt.Errorf("daily spend limit reached, spent: %v, limit: %v", g.DailySpent, g.DailyLimit)
		}
		g.DailySpent += amount
		g.Reserved += amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletServiceImpl) Deduct(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		repo := s.guildRepository
		g, err := repo.Get(ctx, id)
		if errors.Is(err, ErrGuildNotFound) {
			return fmt.Errorf("guild %s does not exist: %w", id, err)
		}
		if err != nil {
			return err
		}

		if amount > g.Reserved {
			return fmt.Errorf("insufficient reserve")
		}
		if amount > g.Gold {
			return fmt.Errorf("insufficient gold")
		}
		g.Reserved -= amount
		g.Gold -= amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletServiceImpl) Release(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
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

		if g.DailySpent > amount {
			g.DailySpent -= amount
		} else {
			g.DailySpent = 0
		}
		g.Reserved -= amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletServiceImpl) Earn(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
		repo := s.guildRepository
		g, err := repo.Get(ctx, id)
		if errors.Is(err, ErrGuildNotFound) {
			return fmt.Errorf("guild %s does not exist: %w", id, err)
		}
		if err != nil {
			return err
		}

		g.Gold += amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletServiceImpl) Spend(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.tx.RunInTransaction(ctx, func(ctx context.Context) error {
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
		if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
			return fmt.Errorf("daily spend limit reached, spent: %v, limit: %v", g.DailySpent, g.DailyLimit)
		}

		g.DailySpent += amount
		g.Gold -= amount
		return repo.Update(ctx, g)
	})
}

func (s *WalletServiceImpl) GetGuild(ctx context.Context, id string) (*Guild, error) {
	g, err := s.guildRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return g, nil
}
