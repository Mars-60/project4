package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mars-60/project4/backend/configs"
	"github.com/Mars-60/project4/backend/internal/logger"
	"go.uber.org/zap"
)

func StartServer() error {

	app, err := NewApp(context.Background())
	if err != nil {
		return err
	}

	router := NewRouter(app)

	address := fmt.Sprintf(
		"%s:%s",
		configs.App.Server.Host,
		configs.App.Server.Port,
	)

	logger.Log.Info(
		"HTTP server started",
		zap.String("address", address),
	)

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  configs.App.Server.ReadTimeout,
		WriteTimeout: configs.App.Server.WriteTimeout,
		IdleTimeout:  configs.App.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case sig := <-shutdownCh:
		logger.Log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), configs.App.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			return fmt.Errorf("database shutdown failed: %w", err)
		}
	}

	logger.Log.Info("HTTP server stopped")
	return nil

}
