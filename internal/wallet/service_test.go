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
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       "guild-1",
			Gold:     200,
			Reserved: 10,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Reserve("guild-1", 100)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g, _ := repo.Get("guild-1")
	if g.Reserved != 110 {
		t.Errorf("expected reserved: %v, got: %v", 110, g.Reserved)
	}
}

func TestWallet_Service_Reserve_Insufficient(t *testing.T) {
	repo := MockGuildRepo{guilds: map[string]*models.Guild{
		"guild-1": {
			ID:       "guild-1",
			Gold:     200,
			Reserved: 10,
		},
	}}

	svc := NewWalletService(&repo)

	err := svc.Reserve("guild-1", 200)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	g, _ := repo.Get("guild-1")
	if g.Reserved != 10 {
		t.Errorf("expected reserved: %v, got: %v", 10, g.Reserved)
	}
}
