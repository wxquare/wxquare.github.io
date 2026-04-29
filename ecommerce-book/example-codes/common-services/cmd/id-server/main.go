package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"common-services/internal/bootstrap"
)

func main() {
	ctx := context.Background()
	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatalf("bootstrap id service: %v", err)
	}
	defer app.Close(context.Background())

	errCh := make(chan error, 1)
	go func() {
		log.Printf("id-service listening on %s", app.Addr)
		errCh <- app.Server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := app.Server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown failed: %v", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("id-service stopped: %v", err)
		}
	}
}
