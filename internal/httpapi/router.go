package httpapi

import "net/http"

func (app *Application) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/spacecraft", app.spacecraftCollectionHandler)
	mux.HandleFunc("/spacecraft/{id}", app.spacecraftByIDHandler)
	mux.HandleFunc("/ready", app.readinessHandler)
	mux.HandleFunc("/health", app.healthHandler)
	mux.HandleFunc("/", app.notFoundHandler)

	return app.requestLogger(app.recoverPanic(mux))
}
