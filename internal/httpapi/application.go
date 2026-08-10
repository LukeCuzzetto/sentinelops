package httpapi

import (
	"log"
)

type Application struct {
	logger   *log.Logger
	database Database
}

func NewApplication(logger *log.Logger, database Database) *Application {
	return &Application{
		logger:   logger,
		database: database,
	}
}
