package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB        *sql.DB
	testContainer testcontainers.Container
)

func LoadMigration(filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join("../../migrations/up", filename))
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
		"000003_create_items_table.up.sql",
		"000004_create_limit_orders_table.up.sql",
		"000005_add_status_to_items_table.up.sql",
		"000006_create_auctions_table.up.sql",
		"000007_create_bids_table.up.sql",
		"000008_restore_items_available.up.sql",
	}

	for _, migFile := range migrations {
		sqlStr, err := LoadMigration(migFile)
		if err != nil {
			panic(err)
		}
		if _, err := testDB.ExecContext(ctx, sqlStr); err != nil {
			panic(err)
		}
	}

	code := m.Run()

	testDB.Close()
	testContainer.Terminate(ctx)
	os.Exit(code)
}

func TruncateTables(ctx context.Context, tables ...string) error {
	for _, t := range tables {
		if _, err := testDB.ExecContext(ctx, "TRUNCATE "+t+" CASCADE"); err != nil {
			return err
		}
	}
	return nil
}
