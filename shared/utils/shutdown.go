package utils

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown handles graceful shutdown of HTTP server
func GracefulShutdown(server *http.Server, timeout time.Duration, onShutdown func()) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-quit

	// Call custom shutdown handler if provided
	if onShutdown != nil {
		onShutdown()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		// Force shutdown if graceful shutdown fails
		server.Close()
	}
}

// SetupSignalHandler sets up a signal handler that returns a context
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		cancel()
	}()

	return ctx
}
