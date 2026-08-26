package spacecraft

import (
	"context"
	"errors"
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

func TestRepositoryGetSpacecraftByID(t *testing.T) {

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

	name := "GET_TEST-" + time.Now().Format("20060102150405")

	created, err := repository.CreateSpacecraft(ctx, name)
	if err != nil {
		t.Fatalf("failed to create spacecraft: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM spacecraft WHERE id = $1", created.ID)
	}()

	found, err := repository.GetSpacecraftByID(ctx, created.ID)

	if err != nil {
		t.Fatalf("failed to get spacecraft by ID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}

	if found.Name != created.Name {
		t.Errorf("expected name %q, got %q", created.Name, found.Name)
	}
}

func TestRepositoryGetSpacecraftByIDNotFound(t *testing.T) {
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

	_, err = repository.GetSpacecraftByID(ctx, 9223372036854775807)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepositoryListSpacecrafts(t *testing.T) {
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

	timestamp := time.Now().Format("20060102150405")

	first, err := repository.CreateSpacecraft(ctx, "LIST-A-"+timestamp)
	if err != nil {
		t.Fatalf("failed to create first spacecraft: %v", err)
	}

	second, err := repository.CreateSpacecraft(ctx, "LIST-B-"+timestamp)
	if err != nil {
		t.Fatalf("failed to create second spacecraft: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM spacecraft WHERE id IN ($1, $2)", first.ID, second.ID)
	}()

	spacecrafts, err := repository.ListSpacecraft(ctx)
	if err != nil {
		t.Fatalf("failed to list spacecrafts: %v", err)
	}

	var foundFirst, foundSecond bool

	for _, spacecraft := range spacecrafts {
		if spacecraft.ID == first.ID {
			foundFirst = true
		}
		if spacecraft.ID == second.ID {
			foundSecond = true
		}
	}

	if !foundFirst {
		t.Errorf("first spacecraft with ID %d not found in list", first.ID)
	}

	if !foundSecond {
		t.Errorf("second spacecraft with ID %d not found in list", second.ID)
	}
}
