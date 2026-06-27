package wallet

import (
	"errors"
	"market-dragon/internal/models"
	"testing"
)

type MockGuildRepo struct {
	guilds map[string]*models.Guild
}

func (m *MockGuildRepo) Get(id string) (*models.Guild, error) {
	guild, ok := m.guilds[id]
	if !ok {
		return nil, errors.New("guild not found")
	}
	return guild, nil
}

func (m *MockGuildRepo) Update(val *models.Guild) (*models.Guild, error) {
	m.guilds[val.ID] = val

	return val, nil
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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		testGuildID: {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Reserve(testGuildID, testReserveAmount)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g, _ := repo.Get(testGuildID)
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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		testGuildID: {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Reserve(testGuildID, testReserveAmount)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, _ := repo.Get(testGuildID)
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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Deduct("guild-1", testDeductAmount)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	g, err := repo.Get("guild-1")
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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Deduct(testGuildID, testDeductAmount)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, err := repo.Get(testGuildID)
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
		testGuildID              = "guild-1"
		testInitialGold          = 200
		testInitialReserve       = 100
		testInitialAvailableGold = testInitialGold - testInitialReserve
		excessReleaseAmount      = 0
		testReleaseAmount        = testInitialAvailableGold + excessReleaseAmount
		testExpectedGold         = testInitialGold
		testExpectedReserve      = testInitialReserve - testReleaseAmount
	)
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Release(testGuildID, testReleaseAmount)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	g, err := repo.Get(testGuildID)
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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       testGuildID,
			Gold:     testInitialGold,
			Reserved: testInitialReserve,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Release(testGuildID, testReleaseAmount)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, err := repo.Get(testGuildID)
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
