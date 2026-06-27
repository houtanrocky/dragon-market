package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB
var testContainer testcontainers.Container

func LoadMigration(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join("../../migrations", filename))
	if err != nil {
		return "", fmt.Errorf("migration not found: %s: %w", filename, err)
	}
	return string(data), nil
}

// TestMain runs once, sets up container with schema
func TestMain(m *testing.M) {
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
		panic(err)
	}
	testContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		panic(err)
	}
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())

	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	// Load migrations ONCE
	migrations := []string{
		"000001_create_guilds_table.up.sql",
		"000002_add_daily_fields_to_guilds.up.sql",
	}

	for _, migFile := range migrations {
		sql, err := LoadMigration(migFile)
		if err != nil {
			panic(err)
		}
		if _, err := testDB.ExecContext(ctx, sql); err != nil {
			panic(err)
		}
	}

	code := m.Run()

	testDB.Close()
	testContainer.Terminate(ctx)
	os.Exit(code)
}

// Setup for each test - clean state
func setupTest(t *testing.T, ctx context.Context) {
	// Truncate tables before each test
	if _, err := testDB.ExecContext(ctx, "TRUNCATE guilds CASCADE"); err != nil {
		t.Fatal(err)
	}
}

func TestGuildRepository_GetAndUpdate(t *testing.T) {
	const (
		testInitialGuildID = "guild-1"
		testInitialGold    = 200
		testInitialReserve = 100
		testAddedGold      = 10
		testExpectedGold   = testInitialGold + testAddedGold
	)

	ctx := context.Background()
	setupTest(t, ctx)

	repo := NewRepository(testDB)

	// Insert
	_, err := testDB.ExecContext(
		ctx,
		`INSERT INTO guilds (id, gold, reserved) VALUES ($1, $2, $3)`,
		testInitialGuildID, testInitialGold, testInitialReserve,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Test Get
	g, err := repo.Get(ctx, testInitialGuildID)
	if err != nil {
		t.Fatal(err)
	}

	if g.Gold != testInitialGold || g.Reserved != testInitialReserve {
		t.Errorf("unexpected values %+v", g)
	}

	// Test Update
	g.Gold = testExpectedGold
	if err := repo.Update(ctx, g); err != nil {
		t.Fatal(err)
	}

	// Verify update
	updated, err := repo.Get(ctx, testInitialGuildID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Gold != testExpectedGold {
		t.Errorf("expected gold %v, got %v", testExpectedGold, updated.Gold)
	}
}
