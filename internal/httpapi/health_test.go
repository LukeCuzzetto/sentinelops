package httpapi

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
)

type stubDatabase struct {
	pingErr error
}

type stubSpacecraftRepository struct{}

func (stubSpacecraftRepository) CreateSpacecraft(ctx context.Context, name string) (spacecraft.Spacecraft, error) {
	return spacecraft.Spacecraft{}, nil
}

func (stubSpacecraftRepository) GetSpacecraftByID(ctx context.Context, id int64) (spacecraft.Spacecraft, error) {
	return spacecraft.Spacecraft{}, nil
}

func (stubSpacecraftRepository) ListSpacecraft(ctx context.Context) ([]spacecraft.Spacecraft, error) {
	return []spacecraft.Spacecraft{}, nil
}

func (db stubDatabase) Ping(ctx context.Context) error {
	return db.pingErr
}

func newTestApplication(database Database) *Application {
	logger := log.New(&bytes.Buffer{}, "", 0)

	return NewApplication(logger, database, stubSpacecraftRepository{}, telemetryRepositoryStub{})

}

func TestHealthHandler(t *testing.T) {
	app := newTestApplication(stubDatabase{})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	expectedBody := "{\"status\":\"ok\"}\n"

	if recorder.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, recorder.Body.String())
	}
}

func TestReadinnessHandlerWhenDatabaseIsReady(t *testing.T) {
	app := newTestApplication(stubDatabase{})

	request := httptest.NewRequest(http.MethodGet, "/ready", nil)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

}

func TestReadinnessHandlerWhenDatabaseIsUnavailable(t *testing.T) {
	app := newTestApplication(stubDatabase{pingErr: errors.New("database unavailable")})

	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status code %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestHealthHandlerRejectsWrongMethod(t *testing.T) {
	app := newTestApplication(stubDatabase{})

	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}

	if recorder.Header().Get("Allow") != http.MethodGet {
		t.Errorf("expected Allow header %q, got %q", http.MethodGet, recorder.Header().Get("Allow"))
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	app := newTestApplication(stubDatabase{})

	request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}
