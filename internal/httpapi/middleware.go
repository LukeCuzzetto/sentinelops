package httpapi

import (
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (recorder *responseRecorder) WriteHeader(code int) {
	if recorder.statusCode != 0 {
		return
	}

	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}

func (recorder *responseRecorder) Write(b []byte) (int, error) {
	if recorder.statusCode == 0 {
		recorder.WriteHeader(http.StatusOK)
	}

	return recorder.ResponseWriter.Write(b)
}

func (app *Application) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		app.logger.Printf(
			"request method=%s path=%s status=%d duration=%s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			time.Since(start),
		)
	})
}

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				w.Header().Set("Connection", "close")

				app.logger.Printf("panic recovered: %v", recovered)

				_ = writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
