package main

import (
	"log"

	"github.com/jsopn/vrc-lyrics/internal/app"
	"github.com/jsopn/vrc-lyrics/internal/config"
)

func main() {
	cfg, err := config.ParseConfig("config.toml")
	if err != nil {
		log.Fatalf("Failed to parse config: %q", err.Error())
		return
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("An error occured while running the app: %q", err)
		return
	}
}
