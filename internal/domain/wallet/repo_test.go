package wallet

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go"
)

//go:embed migrations/*.sql
var migration embed.FS

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
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
	dsn := fmt.Sprintf("posgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}

	sqlBytes, err := migration.ReadFile("migrations/000001_create_guilds_table.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Run migration/s
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		t.Fatal(err)
	}

	return db, func() { db.Close(); container.Terminate(ctx) }
}

//func TestGuildRepository_GetAndUpdate(t *testing.T) {
//	db, cleanup := setupTestDB(t)
//	defer cleanup()
//
//	repo := NewPostgresRepository(&db)
//	fmt.Println(repo)
//}
