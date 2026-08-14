package spacecraft

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	database *pgxpool.Pool
}

func NewRepository(database *pgxpool.Pool) *Repository {
	return &Repository{database: database}
}

func (repository *Repository) CreateSpacecraft(ctx context.Context, name string) (Spacecraft, error) {
	const query = `
		INSERT INTO spacecraft (name)
		VALUES ($1)
		RETURNING id, name, status, created_at, updated_at
	`

	var spacecraft Spacecraft

	err := repository.database.QueryRow(ctx, query, name).Scan(
		&spacecraft.ID,
		&spacecraft.Name,
		&spacecraft.Status,
		&spacecraft.CreatedAt,
		&spacecraft.UpdatedAt,
	)

	if err != nil {
		return Spacecraft{}, fmt.Errorf("failed to create spacecraft: %w", err)
	}
	return spacecraft, nil
}
