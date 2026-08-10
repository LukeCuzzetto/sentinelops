package httpapi

import (
	"context"
	"net/http"
	"time"
)

const databaseReadinessTimeout = 2 * time.Second

func (app *Application) healthHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		_ = writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	_ = writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (app *Application) readinessHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		_ = writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databaseReadinessTimeout)
	defer cancel()

	if err := app.database.Ping(ctx); err != nil {

		app.logger.Printf("database ping failed: %v", err)

		_ = writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "service not available"})

		return
	}

	_ = writeJSON(w, http.StatusOK, readinessResponse{Status: "ok", Database: "ok"})
}

func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusNotFound, errorResponse{Error: "route not found"})
}
