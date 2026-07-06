package guild

import (
	"context"
	"errors"
	"fmt"
	"market-dragon/internal/gold"
	"strings"
	"time"
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
	now             func() time.Time
}

type transactionRecorder interface {
	RecordWalletTransaction(ctx context.Context, guildID, operation string, amount gold.Amount, state *Guild) error
}

type guildCreator interface {
	Create(ctx context.Context, guild *Guild) error
}

func (s *WalletServiceImpl) CreateGuild(ctx context.Context, id string, initialGold, dailyLimit gold.Amount) (*Guild, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: ID is required", ErrInvalidGuild)
	}
	if initialGold < 0 {
		return nil, fmt.Errorf("%w: initial gold cannot be negative", ErrInvalidGuild)
	}
	if dailyLimit < 0 {
		return nil, fmt.Errorf("%w: daily limit cannot be negative", ErrInvalidGuild)
	}
	creator, ok := s.guildRepository.(guildCreator)
	if !ok {
		return nil, errors.New("guild repository does not support creation")
	}
	g := &Guild{
		ID: id, Gold: initialGold, DailyLimit: dailyLimit,
		SpentOn: s.now().UTC().Truncate(24 * time.Hour),
	}
	if err := creator.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func NewWalletService(r GuildRepository, tx Transactor) *WalletServiceImpl {
	return &WalletServiceImpl{guildRepository: r, tx: tx, now: time.Now}
}

func (s *WalletServiceImpl) resetDailySpend(g *Guild) {
	today := s.now().UTC().Truncate(24 * time.Hour)
	if g.SpentOn.IsZero() {
		g.SpentOn = today
		return
	}
	if g.SpentOn.UTC().Truncate(24*time.Hour) != today {
		g.DailySpent = 0
		g.SpentOn = today
	}
}

func (s *WalletServiceImpl) save(ctx context.Context, g *Guild, operation string, amount gold.Amount) error {
	if err := s.guildRepository.Update(ctx, g); err != nil {
		return err
	}
	if recorder, ok := s.guildRepository.(transactionRecorder); ok {
		return recorder.RecordWalletTransaction(ctx, g.ID, operation, amount, g)
	}
	return nil
}

func (s *WalletServiceImpl) Reserve(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return ErrInvalidAmount
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
		s.resetDailySpend(g)

		available := g.Gold - g.Reserved
		if available < amount {
			return fmt.Errorf("%w: available %v, requested %v", ErrInsufficientBalance, available, amount)
		}
		if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
			return fmt.Errorf("%w: spent %v, limit %v", ErrDailyLimitReached, g.DailySpent, g.DailyLimit)
		}
		g.DailySpent += amount
		g.Reserved += amount
		return s.save(ctx, g, "reserve", amount)
	})
}

func (s *WalletServiceImpl) Deduct(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return ErrInvalidAmount
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
		s.resetDailySpend(g)

		if amount > g.Reserved {
			return ErrInsufficientReserve
		}
		if amount > g.Gold {
			return ErrInsufficientBalance
		}
		g.Reserved -= amount
		g.Gold -= amount
		return s.save(ctx, g, "deduct", amount)
	})
}

func (s *WalletServiceImpl) Release(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return ErrInvalidAmount
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
		s.resetDailySpend(g)

		enoughReserve := g.Reserved >= amount
		if !enoughReserve {
			return fmt.Errorf("%w: have %v, need %v", ErrInsufficientReserve, g.Reserved, amount)
		}

		if g.DailySpent > amount {
			g.DailySpent -= amount
		} else {
			g.DailySpent = 0
		}
		g.Reserved -= amount
		return s.save(ctx, g, "release", amount)
	})
}

func (s *WalletServiceImpl) Earn(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return ErrInvalidAmount
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
		s.resetDailySpend(g)

		g.Gold += amount
		return s.save(ctx, g, "earn", amount)
	})
}

func (s *WalletServiceImpl) Spend(ctx context.Context, id string, amount gold.Amount) error {
	if amount <= 0 {
		return ErrInvalidAmount
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
		s.resetDailySpend(g)

		available := g.Gold - g.Reserved
		if available < amount {
			return fmt.Errorf("%w: have %v, need %v", ErrInsufficientBalance, available, amount)
		}
		if g.DailyLimit > 0 && g.DailySpent+amount > g.DailyLimit {
			return fmt.Errorf("%w: spent %v, limit %v", ErrDailyLimitReached, g.DailySpent, g.DailyLimit)
		}

		g.DailySpent += amount
		g.Gold -= amount
		return s.save(ctx, g, "spend", amount)
	})
}

func (s *WalletServiceImpl) GetGuild(ctx context.Context, id string) (*Guild, error) {
	g, err := s.guildRepository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return g, nil
}
