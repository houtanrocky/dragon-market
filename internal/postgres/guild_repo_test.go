package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
)

func LoadMigration(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join("../../migrations", filename))
	if err != nil {
		return "", fmt.Errorf("migration not found: %s: %w", filename, err)
	}
	return string(data), nil
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}

	sqlMigration, err := LoadMigration("000001_create_guilds_table.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Run migration/s
	_, err = db.Exec(sqlMigration)
	if err != nil {
		t.Fatal(err)
	}

	return db, func() { db.Close(); container.Terminate(ctx) }
}

func TestGuildRepository_GetAndUpdate(t *testing.T) {
	const (
		testInitialGuildID = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100
		testAddedGold      = 10
		testExpectedGold   = testInitialGold + testAddedGold
	)
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	// insert a guild
	_, err := db.ExecContext(ctx, `INSERT INTO guilds (id, gold, reserved) VALUES ($1, $2, $3)`, testInitialGuildID, testInitialGold, testInitialReserve)
	if err != nil {
		t.Fatal(err)
	}

	// test get
	g, err := repo.Get(ctx, testInitialGuildID)
	if err != nil {
		t.Fatal(err)
	}
	if g.Gold != testInitialGold || g.Reserved != testInitialReserve {
		t.Errorf("unexpected values %+v", g)
	}

	// test update
	g.Gold = testExpectedGold
	err = repo.Update(ctx, g)
	if err != nil {
		t.Fatal(err)
	}

	var gold float64
	db.QueryRowContext(ctx, `SELECT gold FROM guilds WHERE id = $1`, testInitialGuildID).Scan(&gold)
	if gold != testExpectedGold {
		t.Errorf("expected gold %v, got %v", testExpectedGold, gold)
	}
}
