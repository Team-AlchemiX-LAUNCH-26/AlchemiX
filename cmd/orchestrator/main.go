package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/orchestrator"
)

func main() {
	configPath := env("CONFIG_PATH", env("UNIVERSE_CONFIG_PATH", "configs/universe-config.json"))
	addr := env("ORCHESTRATOR_ADDR", ":8080")
	publicURL := env("ORCHESTRATOR_PUBLIC_URL", "http://localhost:8080")
	threshold, _ := strconv.Atoi(env("HEALTH_FAILURE_THRESHOLD", "2"))
	if threshold < 1 {
		threshold = 2
	}
	interval, err := time.ParseDuration(env("HEALTH_INTERVAL", "2s"))
	if err != nil {
		interval = 2 * time.Second
	}
	service, err := orchestrator.New(configPath, publicURL, threshold)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go service.StartHealthMonitor(ctx, interval)
	server := &http.Server{Addr: addr, Handler: service.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("orchestrator listening on %s for universe %s", addr, service.Config.Metadata.SystemName)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
