package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// serve runs the HTTP server until ctx is cancelled (SIGINT/SIGTERM), then
// shuts down gracefully.
func (app *App) serve(ctx context.Context) error {
	s := &http.Server{
		Addr:              app.cfg.ParsePort(),
		Handler:           app.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("halora-be listening on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.Shutdown(shutCtx)
}
