package main

import (
	"bufio"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cm-manaus/liturgical-calendar-engine/api"
	"github.com/cm-manaus/liturgical-calendar-engine/engine"
)

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	// Initialize structured JSON logging (Google/Enterprise standard)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	loadDotEnv(".env")

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	slog.Info("initializing_engine", "data_dir", dataDir)
	litEngine := engine.NewLiturgicalEngine(dataDir)
	locMgr := engine.NewLocalizationManager(dataDir)

	// Clean Dependency Injection (SoC: Zero global variables)
	handler := api.NewHandler(litEngine, locMgr)
	router := api.NewRouter(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// P0: Hardened HTTP Server with explicit connection timeouts (Anti-Slowloris)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB header limit
	}

	// P0: Graceful Shutdown listener
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("server_listening", "port", port, "version", "2.1.0")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_listen_error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting_down_server_gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server_shutdown_failed", "error", err)
	} else {
		slog.Info("server_stopped_cleanly")
	}
}
