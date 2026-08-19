package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"tamagops/daemon/pkg/collector"
	"tamagops/daemon/pkg/engine"
	"tamagops/daemon/pkg/storage"
)

const tickInterval = 2 * time.Second

const eventPetUpdate = "pet:update"

type DashboardPayload struct {
	Timestamp string                   `json:"timestamp"`
	Pet       engine.PetState          `json:"pet"`
	Hardware  collector.SystemSnapshot `json:"hardware"`
	RecentLog string                   `json:"recent_log"`
}

type App struct {
	ctx context.Context

	store *storage.Storage

	mu     sync.Mutex
	last   DashboardPayload
	stopCh chan struct{}

	// register the last mood and time
	lastLoggedMood string
	lastLoggedAt   time.Time
}

func NewApp() *App {
	return &App{
		stopCh: make(chan struct{}),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	homeDir, err := os.UserHomeDir()
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "failed to resolve home directory: %v", err)
		return
	}

	dataDir := filepath.Join(homeDir, ".tamagotchi")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		wailsRuntime.LogErrorf(a.ctx, "failed to create data directory: %v", err)
		return
	}

	store, err := storage.Open(filepath.Join(dataDir, "data.db"))
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "failed to open storage: %v", err)
		return
	}
	a.store = store

	state, err := a.store.LoadPetState()
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "failed to load pet state: %v", err)
		return
	}
	if state == nil {
		fresh := engine.NewPet("Linus")
		state = &fresh
	}

	a.mu.Lock()
	a.last = DashboardPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Pet:       *state,
		RecentLog: "Aguardando a primeira leitura de hardware...",
	}
	a.mu.Unlock()

	go a.loop(*state)
}

func (a *App) shutdown(ctx context.Context) {
	close(a.stopCh)
	if a.store != nil {
		a.store.Close()
	}
}

func (a *App) loop(state engine.PetState) {
	ticker := time.NewTicker(tickInterval)
	const logRepeatInterval = 5 * time.Minute
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return

		case <-ticker.C:
			snapshot, err := collector.Collect()
			if err != nil {
				wailsRuntime.LogErrorf(a.ctx, "collector error: %v", err)
				continue
			}

			next := engine.CalculatePetState(snapshot, state)
			state = next

			if err := a.store.SavePetState(state); err != nil {
				wailsRuntime.LogErrorf(a.ctx, "storage save error: %v", err)
			}

			logMessage := engine.BuildLogMessage(snapshot, state)

			displayLog := ""
			moodChanged := state.Mood != a.lastLoggedMood
			timeToRepeat := time.Since(a.lastLoggedAt) >= logRepeatInterval
			if moodChanged || timeToRepeat {
				if err := a.store.LogEvent(logMessage, state.Mood); err != nil {
					wailsRuntime.LogErrorf(a.ctx, "storage log error: %v", err)
				}
				a.lastLoggedMood = state.Mood
				a.lastLoggedAt = time.Now()
				displayLog = logMessage
			}

			payload := DashboardPayload{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Pet:       state,
				Hardware:  snapshot,
				RecentLog: displayLog,
			}

			a.mu.Lock()
			a.last = payload
			a.mu.Unlock()

			wailsRuntime.EventsEmit(a.ctx, eventPetUpdate, payload)
		}
	}
}

func (a *App) GetSnapshot() DashboardPayload {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.last
}
