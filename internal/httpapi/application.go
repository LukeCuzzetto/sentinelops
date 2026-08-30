package httpapi

import (
	"log"
)

type Application struct {
	logger               *log.Logger
	database             Database
	spacecraftRepository SpacecraftRepository
	telemetryRepository  TelemetryRepository
}

func NewApplication(logger *log.Logger, database Database, spacecraftRepository SpacecraftRepository, telemetryRepository TelemetryRepository) *Application {
	return &Application{
		logger:               logger,
		database:             database,
		spacecraftRepository: spacecraftRepository,
		telemetryRepository:  telemetryRepository,
	}
}
