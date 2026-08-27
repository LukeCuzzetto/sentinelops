package httpapi

import (
	"log"
)

type Application struct {
	logger               *log.Logger
	database             Database
	spacecraftRepository SpacecraftRepository
}

func NewApplication(logger *log.Logger, database Database, spacecraftRepository SpacecraftRepository) *Application {
	return &Application{
		logger:               logger,
		database:             database,
		spacecraftRepository: spacecraftRepository,
	}
}
