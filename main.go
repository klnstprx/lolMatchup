package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/data"
	"github.com/klnstprx/lolMatchup/router"
)

func main() {
	// Initialize configuration
	cfg := config.New()

	// Load configuration from file if available
	const configPath = "config.toml"
	if _, err := os.Stat(configPath); err == nil {
		if err := cfg.Load(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", configPath, err)
			os.Exit(1)
		}
	}

	// Initialize objects (logger, cache, etc.)
	if err := cfg.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing AppConfig: %v\n", err)
		os.Exit(1)
	}

	// Load cache from disk if present
	if err := cfg.Cache.Load(); err != nil {
		cfg.Logger.Warnf("Cache not loaded (possibly first run): %v", err)
	}

	// Create the Riot API client
	apiClient := &client.Client{
		HTTPClient: cfg.HTTPClient,
		Logger:     cfg.Logger,
	}

	// Prepare data loader, fetch or refresh patch info
	dataLoader := data.NewDataLoader(cfg, apiClient, cfg.Cache)

	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	if err := dataLoader.Initialize(initCtx); err != nil {
		cfg.Logger.Fatalf("Error during data initialization: %v", err)
	}

	// Set up router
	r := router.SetupRouter(cfg, apiClient)

	// Handle graceful shutdown signals
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-sigChan
		cfg.Logger.Info("Received signal", "signal", s)
		stop()
	}()

	httpServer := &http.Server{
		Addr:    cfg.ListenAddr + ":" + strconv.Itoa(cfg.Port),
		Handler: r,
	}

	// Start serving in a goroutine
	go func() {
		cfg.Logger.Infof("HTTP server starting on port %d", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cfg.Logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownCtx.Done()
	cfg.Logger.Info("Shutdown signal received; stopping HTTP server")

	// Graceful shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		cfg.Logger.Errorf("HTTP server shutdown error: %v", err)
	}

	// Save state (cache) before exiting
	if err := cfg.Cache.Save(); err != nil {
		cfg.Logger.Errorf("Error saving cache during shutdown: %v", err)
	} else {
		cfg.Logger.Info("Cache saved successfully on shutdown.")
	}

	cfg.Logger.Info("Server shut down gracefully")
}
