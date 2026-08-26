package spacecraft

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("spacecraft not found")

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

func (repository *Repository) GetSpacecraftByID(ctx context.Context, id int64) (Spacecraft, error) {
	const query = `
		SELECT id, name, status, created_at, updated_at
		FROM spacecraft
		WHERE id = $1
	`

	var spacecraft Spacecraft

	err := repository.database.QueryRow(ctx, query, id).Scan(
		&spacecraft.ID,
		&spacecraft.Name,
		&spacecraft.Status,
		&spacecraft.CreatedAt,
		&spacecraft.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Spacecraft{}, fmt.Errorf(
			"%w: id %d",
			ErrNotFound,
			id,
		)
	}

	if err != nil {
		return Spacecraft{}, fmt.Errorf("failed to get spacecraft by id: %w", err)
	}

	return spacecraft, nil
}

func (repository *Repository) ListSpacecraft(ctx context.Context) ([]Spacecraft, error) {
	const query = `SELECT  id, name, status, created_at, updated_at FROM spacecraft ORDER BY id`

	rows, err := repository.database.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to list spacecraft: %w", err)
	}
	defer rows.Close()

	spacecraftList := make([]Spacecraft, 0)

	for rows.Next() {
		var spacecraft Spacecraft
		err := rows.Scan(
			&spacecraft.ID,
			&spacecraft.Name,
			&spacecraft.Status,
			&spacecraft.CreatedAt,
			&spacecraft.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan spacecraft: %w", err)
		}

		spacecraftList = append(spacecraftList, spacecraft)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", rows.Err())
	}

	return spacecraftList, nil

}
