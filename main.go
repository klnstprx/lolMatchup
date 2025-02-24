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

	// Load configuration from file if it exists
	if _, err := os.Stat("config.toml"); err == nil {
		if err := cfg.Load("config.toml"); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config.toml: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize logger, cache, HTTP client, etc.
	if err := cfg.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing AppConfig: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Cache.Load(); err != nil {
		cfg.Logger.Warnf("Cache not loaded (might be first run): %v", err)
	}

	// Create the client with the HTTP client and logger
	apiClient := &client.Client{
		HTTPClient: cfg.HTTPClient,
		Logger:     cfg.Logger,
	}

	dataLoader := data.NewDataLoader(cfg, apiClient, cfg.Cache)

	// Create a context for initialization
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	// Initialize data
	if err := dataLoader.Initialize(initCtx); err != nil {
		cfg.Logger.Fatalf("Error during data initialization: %v", err)
	}

	// Set up the router, passing the configuration and the client
	r := router.SetupRouter(cfg, apiClient)

	// Context for graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Graceful shutdown logic
	signalCaught := false
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		if signalCaught {
			cfg.Logger.Warn("Caught second signal, terminating immediately")
			os.Exit(1)
		}
		signalCaught = true
		cfg.Logger.Info("Caught shutdown signal.")
		stop()
	}()

	// Start the HTTP server
	httpServer := http.Server{
		Addr:    cfg.ListenAddr + ":" + strconv.Itoa(cfg.Port),
		Handler: r,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cfg.Logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	cfg.Logger.Infof("HTTP server started on port %d", cfg.Port)

	// Wait for shutdown signal
	<-shutdownCtx.Done()

	// Shutdown HTTP server gracefully
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		cfg.Logger.Errorf("HTTP server shutdown error: %v", err)
	}

	// Save the cache to disk before exiting
	if err := cfg.Cache.Save(); err != nil {
		cfg.Logger.Errorf("Error saving cache during shutdown: %v", err)
	} else {
		cfg.Logger.Info("Cache saved successfully during shutdown.")
	}

	cfg.Logger.Info("Server shut down successfully")
}
