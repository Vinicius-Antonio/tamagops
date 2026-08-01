package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tamagops/daemon/pkg/collector"
	"tamagops/daemon/pkg/engine"
	"tamagops/daemon/pkg/storage"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("failed to resolve home directory:", err)
		os.Exit(1)
	}

	dataDir := filepath.Join(homeDir, ".tamagotchi")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		fmt.Println("failed to create data directory:", err)
		os.Exit(1)
	}

	store, err := storage.Open(filepath.Join(dataDir, "data.db"))
	if err != nil {
		fmt.Println("failed to open storage:", err)
		os.Exit(1)
	}
	defer store.Close()

	state, err := store.LoadPetState()
	if err != nil {
		fmt.Println("failed to load pet state:", err)
		os.Exit(1)
	}
	if state == nil {
		fresh := engine.NewPet("Linus")
		state = &fresh
	}

	fmt.Println("Tamagotchi daemon started. Press Ctrl+C to stop.")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		snapshot, err := collector.Collect()
		if err != nil {
			fmt.Println("collector error:", err)
			continue
		}

		next := engine.CalculatePetState(snapshot, *state)
		state = &next

		if err := store.SavePetState(*state); err != nil {
			fmt.Println("storage save error:", err)
		}

		logMessage := engine.BuildLogMessage(snapshot, *state)
		if err := store.LogEvent(logMessage, state.Mood); err != nil {
			fmt.Println("storage log error:", err)
		}

		payload := map[string]interface{}{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"pet":        state,
			"hardware":   snapshot,
			"recent_log": logMessage,
		}

		out, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(out))
	}
}