package main

import (
	"PessiTorrent/internal/config"
	"PessiTorrent/internal/logger"
	"flag"
)

func main() {
	cfg, err := config.NewConfig(config.DefaultConfigPath)
	if err != nil {
		logger.Error("Failed to load config: %s", err)
	}

	port := cfg.Tracker.Port

	flag.UintVar(&port, "p", port, "Port to listen on")
	flag.Parse()

	tracker := NewTracker(uint16(port))
	tracker.Start()
}
