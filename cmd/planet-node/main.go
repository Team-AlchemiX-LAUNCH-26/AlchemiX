package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/launch26/relic-ring-protocol/internal/config"
	"github.com/launch26/relic-ring-protocol/internal/planet"
)

func main() {
	planetFlag := flag.String("planet", "", "planet ID from the universe configuration")
	portFlag := flag.Int("port", 0, "HTTP port")
	configFlag := flag.String("config", "", "path to universe-config.json")
	flag.Parse()

	planetID := firstNonEmpty(*planetFlag, os.Getenv("PLANET_ID"))
	if planetID == "" {
		log.Fatal("planet ID is required through --planet or PLANET_ID")
	}
	configPath := firstNonEmpty(*configFlag, os.Getenv("CONFIG_PATH"), os.Getenv("UNIVERSE_CONFIG_PATH"), "configs/universe-config.json")
	port := *portFlag
	if port == 0 {
		if value, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
			port = value
		}
	}
	if port == 0 {
		port = 8101
	}

	cfg, err := config.LoadUniverseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	p, err := config.FindPlanetByID(cfg, planetID)
	if err != nil {
		log.Fatal(err)
	}
	service := planet.NewService(cfg, p)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("planet %s listening on %s (codex=%d towers=%d)", p.ID, addr, p.Codex, p.ActiveTowers)
	server := &http.Server{Addr: addr, Handler: service.Handler(), ReadHeaderTimeout: 5e9}
	log.Fatal(server.ListenAndServe())
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
