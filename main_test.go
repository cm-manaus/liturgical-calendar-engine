package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cm-manaus/liturgical-calendar-engine/api"
	"github.com/cm-manaus/liturgical-calendar-engine/engine"
)

func TestCompositionRootWiring(t *testing.T) {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		dataDir = filepath.Join("..", "data")
	}

	litEngine := engine.NewLiturgicalEngine(dataDir)
	if litEngine == nil {
		t.Fatal("Failed to initialize LiturgicalEngine in main wiring")
	}

	locMgr := engine.NewLocalizationManager(dataDir)
	if locMgr == nil {
		t.Fatal("Failed to initialize LocalizationManager in main wiring")
	}

	handler := api.NewHandler(litEngine, locMgr)
	if handler == nil {
		t.Fatal("Failed to initialize api.Handler")
	}

	router := api.NewRouter(handler)
	if router == nil {
		t.Fatal("Failed to initialize api.Router")
	}
}
