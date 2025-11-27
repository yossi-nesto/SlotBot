package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yossigruner/SlotBot/internal/calendar"
	"github.com/yossigruner/SlotBot/internal/config"
	"github.com/yossigruner/SlotBot/internal/slack"
)

func main() {
	// Setup structured logging (JSON)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	calClient, err := calendar.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create calendar client: %w", err)
	}

	slackHandler := slack.NewHandler(calClient)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // Chi's logger is fine for HTTP access logs, or we can replace it
	r.Use(middleware.Recoverer)

	// Slack command endpoints
	r.Route("/slack", func(r chi.Router) {
		r.Use(slack.VerifySignature(cfg.SlackSigningSecret))
		r.Post("/book", slackHandler.HandleBook)
		r.Post("/next", slackHandler.HandleNextSlot)
		r.Post("/bookings", slackHandler.HandleList)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				slog.Error("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		serverStopCtx()
	}()

	slog.Info("Starting server", "port", cfg.Port)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	slog.Info("Server exited properly")
	return nil
}
