package httpapi

import "net/http"

func (app *Application) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/ready", app.readinessHandler)
	mux.HandleFunc("/health", app.healthHandler)
	mux.HandleFunc("/", app.notFoundHandler)

	return app.requestLogger(app.recoverPanic(mux))
}
