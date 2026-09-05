package httpapi

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
	"github.com/LukeCuzzetto/sentinelops/internal/telemetry"
)

type spacecraftRepositoryStub struct {
	createResult spacecraft.Spacecraft
	createErr    error

	getResult spacecraft.Spacecraft
	getErr    error

	listResult []spacecraft.Spacecraft
	listErr    error
}

type telemetryRepositoryStub struct {
}

func (stub telemetryRepositoryStub) CreateSample(ctx context.Context, params telemetry.CreateSampleParams) (telemetry.Sample, error) {
	return telemetry.Sample{}, nil
}

func (stub telemetryRepositoryStub) ListSamplesBySpacecraftID(ctx context.Context, spacecraftID int64) ([]telemetry.Sample, error) {
	return []telemetry.Sample{}, nil
}

func (stub spacecraftRepositoryStub) CreateSpacecraft(ctx context.Context, name string) (spacecraft.Spacecraft, error) {
	return stub.createResult, stub.createErr
}

func (stub spacecraftRepositoryStub) GetSpacecraftByID(ctx context.Context, id int64) (spacecraft.Spacecraft, error) {
	return stub.getResult, stub.getErr
}

func (stub spacecraftRepositoryStub) ListSpacecraft(ctx context.Context) ([]spacecraft.Spacecraft, error) {
	return stub.listResult, stub.listErr
}

func newSpacecraftTestApplication(repository SpacecraftRepository) *Application {
	logger := log.New(
		&bytes.Buffer{},
		"",
		0,
	)

	return NewApplication(logger, stubDatabase{}, repository, telemetryRepositoryStub{})
}

func TestCreateSpacecraftHandler(t *testing.T) {
	now := time.Now()

	repository := spacecraftRepositoryStub{
		createResult: spacecraft.Spacecraft{
			ID:        1,
			Name:      "PATHFINDER-1",
			Status:    spacecraft.StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	app := newSpacecraftTestApplication(
		repository,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/spacecraft",
		strings.NewReader(
			`{"name":"PATHFINDER-1"}`,
		),
	)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusCreated {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"name":"PATHFINDER-1"`,
	) {
		t.Errorf(
			"unexpected response body: %s",
			recorder.Body.String(),
		)
	}
}

func TestCreateSpacecraftHandlerRequiresName(
	t *testing.T,
) {
	app := newSpacecraftTestApplication(
		spacecraftRepositoryStub{},
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/spacecraft",
		strings.NewReader(
			`{"name":"   "}`,
		),
	)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"name is required",
	) {
		t.Errorf(
			"unexpected response body: %s",
			recorder.Body.String(),
		)
	}
}

func TestListSpacecraftHandler(t *testing.T) {
	repository := spacecraftRepositoryStub{
		listResult: []spacecraft.Spacecraft{
			{
				ID:     1,
				Name:   "PATHFINDER-1",
				Status: spacecraft.StatusActive,
			},
			{
				ID:     2,
				Name:   "SENTINEL-2",
				Status: spacecraft.StatusActive,
			},
		},
	}

	app := newSpacecraftTestApplication(
		repository,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/spacecraft",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	body := recorder.Body.String()

	if !strings.Contains(
		body,
		"PATHFINDER-1",
	) {
		t.Errorf(
			"expected PATHFINDER-1 in response",
		)
	}

	if !strings.Contains(
		body,
		"SENTINEL-2",
	) {
		t.Errorf(
			"expected SENTINEL-2 in response",
		)
	}
}

func TestGetSpacecraftByIDHandler(t *testing.T) {
	repository := spacecraftRepositoryStub{
		getResult: spacecraft.Spacecraft{
			ID:     7,
			Name:   "PATHFINDER-7",
			Status: spacecraft.StatusActive,
		},
	}

	app := newSpacecraftTestApplication(
		repository,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/spacecraft/7",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"PATHFINDER-7",
	) {
		t.Errorf(
			"unexpected response body: %s",
			recorder.Body.String(),
		)
	}
}

func TestGetSpacecraftByIDHandlerNotFound(
	t *testing.T,
) {
	repository := spacecraftRepositoryStub{
		getErr: spacecraft.ErrNotFound,
	}

	app := newSpacecraftTestApplication(
		repository,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/spacecraft/999",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}
