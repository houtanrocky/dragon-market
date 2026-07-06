package guild

import (
	"context"
	"errors"
	"market-dragon/internal/gold"
	"testing"
	"time"
)

func TestWallet_Service_Reserve_ResetsSpentOnNewDay(t *testing.T) {
	today := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	repo := MockGuildRepo{guilds: map[string]*Guild{"guild-1": {
		ID: "guild-1", Gold: 1000, DailyLimit: 200, DailySpent: 200, SpentOn: today.AddDate(0, 0, -1),
	}}}
	svc := NewWalletService(&repo, &MockTransactor{})
	svc.now = func() time.Time { return today }

	if err := svc.Reserve(context.Background(), "guild-1", 100); err != nil {
		t.Fatal(err)
	}
	g := repo.guilds["guild-1"]
	if g.DailySpent != 100 || !g.SpentOn.Equal(today.Truncate(24*time.Hour)) {
		t.Fatalf("daily spent=%d spent on=%v", g.DailySpent, g.SpentOn)
	}
}

type MockGuildRepo struct {
	guilds map[string]*Guild
}

func (m *MockGuildRepo) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *MockGuildRepo) Get(ctx context.Context, id string) (*Guild, error) {
	g, ok := m.guilds[id]
	if !ok {
		return nil, ErrGuildNotFound
	}
	return g, nil
}

func (m *MockGuildRepo) Update(ctx context.Context, g *Guild) error {
	m.guilds[g.ID] = g

	return nil
}

func (m *MockGuildRepo) Create(_ context.Context, g *Guild) error {
	if _, exists := m.guilds[g.ID]; exists {
		return ErrGuildAlreadyExists
	}
	m.guilds[g.ID] = g
	return nil
}

func TestWalletService_CreateGuild(t *testing.T) {
	repo := &MockGuildRepo{guilds: make(map[string]*Guild)}
	svc := NewWalletService(repo, &MockTransactor{})
	g, err := svc.CreateGuild(context.Background(), " guild-1 ", 1000, 500)
	if err != nil {
		t.Fatal(err)
	}
	if g.ID != "guild-1" || g.Gold != 1000 || g.DailyLimit != 500 || g.SpentOn.IsZero() {
		t.Fatalf("unexpected guild: %+v", g)
	}
	if _, err := svc.CreateGuild(context.Background(), "guild-1", 1000, 500); !errors.Is(err, ErrGuildAlreadyExists) {
		t.Fatalf("duplicate error=%v", err)
	}
}

func TestWalletService_CreateGuild_RejectsInvalidValues(t *testing.T) {
	svc := NewWalletService(&MockGuildRepo{guilds: make(map[string]*Guild)}, &MockTransactor{})
	for _, input := range []struct {
		id          string
		gold, limit gold.Amount
	}{{"", 1, 1}, {"guild", -1, 1}, {"guild", 1, -1}} {
		if _, err := svc.CreateGuild(context.Background(), input.id, input.gold, input.limit); !errors.Is(err, ErrInvalidGuild) {
			t.Fatalf("CreateGuild(%q,%d,%d) error=%v", input.id, input.gold, input.limit, err)
		}
	}
}

type MockTransactor struct{}

func (MockTransactor) RunInTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return fn(ctx)
}

// ---------------------- reserve tests -------------
func TestWallet_Service_Reserve_Success(t *testing.T) {
	const (
		testGuildID        = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100

		testReserveAmount   = 100
		testExpectedReserve = testInitialReserve + testReserveAmount
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		testGuildID: {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g, _ := repo.Get(ctx, testGuildID)
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Reserve_Insufficient(t *testing.T) {
	const (
		testGuildID        = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100
		exessiveReserve    = 1

		testReserveAmount   = testInitialGold - testInitialReserve + exessiveReserve
		testExpectedReserve = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		testGuildID: {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, _ := repo.Get(ctx, testGuildID)
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Reserve_DailyLimitReached(t *testing.T) {
	const (
		testGuildID           = "guild-1"
		testInitialGold       = 200
		testInitialReserve    = 100
		testInitialDailyLimit = 100
		testInitialDailySpent = testInitialDailyLimit

		testReserveAmount      = 100
		testExpectedReserve    = testInitialReserve
		testExpectedDailySpent = testInitialDailySpent
	)

	repo := MockGuildRepo{guilds: map[string]*Guild{
		testGuildID: {
			ID:         testGuildID,
			Gold:       testInitialGold,
			Reserved:   testInitialReserve,
			DailyLimit: testInitialDailyLimit,
			DailySpent: testInitialDailySpent,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)
	if err == nil {
		t.Error("expected error got nil")
	}

	g, _ := repo.Get(ctx, testGuildID)
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
	if g.DailySpent != testExpectedDailySpent {
		t.Errorf("expected daily spent: %v, got %v", testExpectedDailySpent, g.DailySpent)
	}
}

func TestWallet_Service_Reserve_DailySpentUpdated(t *testing.T) {
	const (
		testGuildID           = "guild-1"
		testInitialGold       = 200
		testInitialReserve    = 100
		testInitialDailyLimit = 100
		testInitialDailySpent = 0

		testReserveAmount      = 100
		testExpectedReserve    = testInitialReserve + testReserveAmount
		testExpectedDailySpent = testInitialDailySpent + testReserveAmount
	)

	repo := MockGuildRepo{guilds: map[string]*Guild{
		testGuildID: {
			ID:         testGuildID,
			Gold:       testInitialGold,
			Reserved:   testInitialReserve,
			DailyLimit: testInitialDailyLimit,
			DailySpent: testInitialDailySpent,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)
	if err != nil {
		t.Error(err)
	}

	g, _ := repo.Get(ctx, testGuildID)
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
	if g.DailySpent != testExpectedDailySpent {
		t.Errorf("expected daily spent: %v, got %v", testExpectedDailySpent, g.DailySpent)
	}
}

func TestWallet_Service_Reserve_ZeroLimitMeansUnlimited(t *testing.T) {
	const (
		testGuildID           = "guild-1"
		testInitialGold       = 1_000_000_000_000
		testInitialReserve    = 0
		testInitialDailyLimit = 0
		testInitialDailySpent = 100

		testReserveAmount      = 1_000_000_000_000
		testExpectedDailySpent = testInitialDailySpent + testReserveAmount
	)

	repo := MockGuildRepo{guilds: map[string]*Guild{
		testGuildID: {
			ID:         testGuildID,
			Gold:       testInitialGold,
			Reserved:   testInitialReserve,
			DailyLimit: testInitialDailyLimit,
			DailySpent: testInitialDailySpent,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)
	if err != nil {
		t.Error(err)
	}

	g, _ := repo.Get(ctx, testGuildID)
	if g.DailySpent != testExpectedDailySpent {
		t.Errorf("expected daily spent: %v, got %v", testExpectedDailySpent, g.DailySpent)
	}
}

// ---------------------- deduct tests -------------
func TestWallet_Service_Deduct_Success(t *testing.T) {
	const (
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		excessDeductAmount       = 0

		testDeductAmount    = testInitialAvailableGold + excessDeductAmount
		testExpectedGold    = testInitialGold - testDeductAmount
		testExpectedReserve = testInitialReserve - testDeductAmount
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Deduct(ctx, testGuildID, testDeductAmount)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	g, err := repo.Get(ctx, testGuildID)
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	if g.Gold != testExpectedGold {
		t.Errorf("expected gold: %v, got: %v", testExpectedGold, g.Gold)
	}
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Deduct_Fail(t *testing.T) {
	const (
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		excessDeductAmount       = 1

		testDeductAmount    = testInitialAvailableGold + excessDeductAmount
		testExpectedGold    = testInitialGold
		testExpectedReserve = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Deduct(ctx, testGuildID, testDeductAmount)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, err := repo.Get(ctx, testGuildID)
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	if g.Gold != testExpectedGold {
		t.Errorf("expected gold: %v, got: %v", testExpectedGold, g.Gold)
	}
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

// ---------------------- release tests -------------
func TestWallet_Service_Release_Success(t *testing.T) {
	const (
		testGuildID         = "guild-1"
		testInitialGold     = 200
		testInitialReserve  = 100
		excessReleaseAmount = 0

		testReleaseAmount   = testInitialReserve + excessReleaseAmount
		testExpectedGold    = testInitialGold
		testExpectedReserve = testInitialReserve - testReleaseAmount
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Release(ctx, testGuildID, testReleaseAmount)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	g, err := repo.Get(ctx, testGuildID)
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	if g.Gold != testExpectedGold {
		t.Errorf("expected gold: %v, got: %v", testExpectedGold, g.Gold)
	}
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Release_Insufficient(t *testing.T) {
	const (
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		excessReleaseAmount      = 1

		testReleaseAmount   = testInitialAvailableGold + excessReleaseAmount
		testExpectedGold    = testInitialGold
		testExpectedReserve = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Release(ctx, testGuildID, testReleaseAmount)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, err := repo.Get(ctx, testGuildID)
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	if g.Gold != testExpectedGold {
		t.Errorf("expected gold: %v, got: %v", testExpectedGold, g.Gold)
	}
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Release_DailySpentDecremented(t *testing.T) {
	const (
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		testInitialDailySpent    = 100
		testInitialDailyLimit    = 100

		testReleaseAmount      = testInitialAvailableGold
		testExpectedDailySpent = testInitialDailySpent - testReleaseAmount
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:         testGuildID,
			Gold:       testInitialGold,
			Reserved:   testInitialReserve,
			DailyLimit: testInitialDailyLimit,
			DailySpent: testInitialDailySpent,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo, &MockTransactor{})

	err := svc.Release(ctx, testGuildID, testReleaseAmount)
	if err != nil {
		t.Error(err)
	}

	g, err := repo.Get(ctx, testGuildID)
	if err != nil {
		t.Errorf("failed to get guild: %v", err)
	}

	if g.DailySpent != testExpectedDailySpent {
		t.Errorf("expected daily spent: %v, got: %v", testExpectedDailySpent, g.DailySpent)
	}
}
