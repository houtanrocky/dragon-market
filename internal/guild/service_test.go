package guild

import (
	"context"
	"errors"
	"testing"
)

type MockGuildRepo struct {
	guilds map[string]*Guild
}

func (m *MockGuildRepo) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *MockGuildRepo) Get(ctx context.Context, id string) (*Guild, error) {
	g, ok := m.guilds[id]
	if !ok {
		return nil, errors.New("g not found")
	}
	return g, nil
}

func (m *MockGuildRepo) Update(ctx context.Context, g *Guild) error {
	m.guilds[g.ID] = g

	return nil
}

func TestWallet_Service_Reserve_Success(t *testing.T) {
	const (
		testGuildID         = "guild-1"
		testInitialGold     = 200
		testInitialReserve  = 100
		excessReserveAmount = 0
		testReserveAmount   = testInitialReserve + excessReserveAmount
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
	svc := NewWalletService(&repo)

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
		testGuildID         = "guild-1"
		testInitialGold     = 200
		testInitialReserve  = 100
		excessReserveAmount = 1
		testReserveAmount   = testInitialReserve + excessReserveAmount
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
	svc := NewWalletService(&repo)

	err := svc.Reserve(ctx, testGuildID, testReserveAmount)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, _ := repo.Get(ctx, testGuildID)
	if g.Reserved != testExpectedReserve {
		t.Errorf("expected reserved: %v, got: %v", testExpectedReserve, g.Reserved)
	}
}

func TestWallet_Service_Deduct_Success(t *testing.T) {
	const (
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		excessDeductAmount       = 0
		testDeductAmount         = testInitialAvailableGold + excessDeductAmount
		testExpectedGold         = testInitialGold - testDeductAmount
		testExpectedReserve      = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo)

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
		testDeductAmount         = testInitialAvailableGold + excessDeductAmount
		testExpectedGold         = testInitialGold
		testExpectedReserve      = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo)

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
	svc := NewWalletService(&repo)

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
		testReleaseAmount        = testInitialAvailableGold + excessReleaseAmount
		testExpectedGold         = testInitialGold
		testExpectedReserve      = testInitialReserve
	)
	repo := MockGuildRepo{guilds: map[string]*Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	ctx := context.Background()
	svc := NewWalletService(&repo)

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
