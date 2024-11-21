package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/router"
)

func main() {
	config.New()
	config.App.SetLogger()
	_, err := os.Stat("config.toml")
	if err != nil {
		config.App.Default()
	} else {
		config.App.Load()
	}
	config.App.SetDDragonDataURL()
	config.App.LogConfig()

	r := router.SetupRouter()

	// context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown logic
	// esto no es muy importante, pero es bueno tenerlo
	signalCaught := false
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		if signalCaught {
			config.App.Logger.Warn("Caught second signal, terminating immediately")
			os.Exit(1)
		}
		signalCaught = true
		config.App.Logger.Info("Caught shutdown signal.")
		cancel()
	}()

	// server starts here
	// starts in a go routine so it doesn't block the main thread
	httpServer := http.Server{
		Addr:    config.App.ListenAddr + ":" + strconv.Itoa(config.App.Port),
		Handler: r,
	}

	// lo arrancamos en un go routine para que no bloquee el main thread
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			config.App.Logger.Fatal("Failed to start HTTP server", "err", err)
		}
	}()
	config.App.Logger.Info("HTTP server started", "port", config.App.Port)
	// Block until context is canceled (waiting for the shutdown signal).
	<-ctx.Done()
	// Shutdown logic
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		config.App.Logger.Error("HTTP server failed to shutdown", "err", err)
	}
	config.App.Logger.Info("Server shut down successfully")
}
