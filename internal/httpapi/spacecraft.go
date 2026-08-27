package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
)

type SpacecraftRepository interface {
	CreateSpacecraft(ctx context.Context, name string) (spacecraft.Spacecraft, error)

	GetSpacecraftByID(ctx context.Context, id int64) (spacecraft.Spacecraft, error)

	ListSpacecraft(ctx context.Context) ([]spacecraft.Spacecraft, error)
}

type createSpacecraftRequest struct {
	Name string `json:"name"`
}

func (app *Application) spacecraftCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.createSpacecraftHandler(w, r)
	case http.MethodGet:
		app.listSpacecraftHandler(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")

		_ = writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (app *Application) spacecraftByIDHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set(
			"Allow",
			http.MethodGet,
		)

		_ = writeJSON(
			w,
			http.StatusMethodNotAllowed,
			errorResponse{
				Error: "method not allowerd",
			},
		)
		return

	}

	idValue := r.PathValue("id")

	id, err := strconv.ParseInt(
		idValue, 10, 64,
	)

	if err != nil || id < 1 {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid spacecraft id",
			},
		)

		return
	}

	found, err := app.spacecraftRepository.GetSpacecraftByID(r.Context(), id)

	if errors.Is(err, spacecraft.ErrNotFound) {
		_ = writeJSON(
			w,
			http.StatusNotFound,
			errorResponse{
				Error: "spacecraft not found",
			},
		)

		return

	}

	if err != nil {
		app.logger.Printf("get spacecraft failed: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)

		return
	}

	_ = writeJSON(
		w,
		http.StatusOK,
		found,
	)
}

func (app *Application) createSpacecraftHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input createSpacecraftRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid request payload",
			},
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		_ = writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "name is required",
			},
		)
		return
	}

	created, err := app.spacecraftRepository.CreateSpacecraft(r.Context(), input.Name)

	if err != nil {
		app.logger.Printf("error creating spacecraft: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	_ = writeJSON(w, http.StatusCreated, created)
}

func (app *Application) listSpacecraftHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	spacecraftList, err := app.spacecraftRepository.ListSpacecraft(r.Context())

	if err != nil {
		app.logger.Printf("error listing spacecraft: %v", err)

		_ = writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	_ = writeJSON(
		w,
		http.StatusOK,
		spacecraftList,
	)
}
