package spacecraft

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryCreate(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		t.Fatalf("failed to connect to the database: %v", err)
	}
	defer pool.Close()

	repository := NewRepository(pool)

	name := "TEST-" + time.Now().Format("20060102150405")

	created, err := repository.CreateSpacecraft(ctx, name)
	if err != nil {
		t.Fatalf("failed to create spacecraft: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM spacecraft WHERE id = $1", created.ID)
	}()

	if created.Name != name {
		t.Errorf("expected name %q, got %q", name, created.Name)
	}

	if created.Status != StatusActive {
		t.Errorf("expected status %q, got %q", StatusActive, created.Status)
	}

	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}

	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	if created.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}
