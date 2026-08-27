package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LukeCuzzetto/sentinelops/internal/config"
	"github.com/LukeCuzzetto/sentinelops/internal/database"
	"github.com/LukeCuzzetto/sentinelops/internal/httpapi"
	"github.com/LukeCuzzetto/sentinelops/internal/spacecraft"
)

const (
	databaseConnectionTimeout = 5 * time.Second
	serverReadHeaderTimeout   = 5 * time.Second
	serverReadTimeout         = 10 * time.Second
	serverWriteTimeout        = 30 * time.Second
	serverIdleTimeout         = 60 * time.Second
	serverShutdownTimeout     = 10 * time.Second
)

func newHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if err := run(logger); err != nil {
		logger.Printf("application error: %v", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)

	}

	databaseContext, cancelDatabase := context.WithTimeout(
		context.Background(),
		databaseConnectionTimeout,
	)
	defer cancelDatabase()

	databasePool, err := database.Open(
		databaseContext,
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	defer func() {
		databasePool.Close()
		logger.Println("database connection pool closed")
	}()
	logger.Println("database connection pool established")

	spacecraftRepository := spacecraft.NewRepository(databasePool)

	app := httpapi.NewApplication(logger, databasePool, spacecraftRepository)

	router := app.Router()

	server := newHTTPServer(cfg.Address, router)

	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	logger.Printf("SentinelOps API listening on %s", cfg.Address)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("run server: %w", err)

		}
		return nil

	case <-shutdownSignal.Done():
		logger.Println("shutdown signal received")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		serverShutdownTimeout,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Println("SentinelOps API stopped gracefully")

	return nil

}
