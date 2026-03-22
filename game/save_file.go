package game

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SaveStatusCode identifies the outcome of the most recent save/load/reset operation.
type SaveStatusCode int

const (
	// SaveStatusNone is the default; no operation has completed yet.
	SaveStatusNone SaveStatusCode = iota
	// SaveStatusSaved means the last save succeeded.
	SaveStatusSaved
	// SaveStatusLoaded means the last load succeeded.
	SaveStatusLoaded
	// SaveStatusReset means the game was reset to a new state.
	SaveStatusReset
	// SaveStatusSaveFailed means the last save failed.
	SaveStatusSaveFailed
	// SaveStatusLoadFailed means the last load failed.
	SaveStatusLoadFailed
)

// SaveStatus records the outcome and time of the most recent save/load/reset.
type SaveStatus struct {
	Code  SaveStatusCode
	Err   error
	SetAt time.Time
}

// errSaveNotSupported is returned when save/load is attempted on a platform
// that does not support file I/O (e.g. WASM/browser).
var errSaveNotSupported = errors.New("save/load not supported on this platform")

// savePath returns the path to the single save file.
// On desktop: <os.UserConfigDir>/forester/save.json.
func savePath() (string, error) {
	if runtime.GOOS == "js" {
		return "", errSaveNotSupported
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "forester", "save.json"), nil
}

// saveToFile serializes the game state to JSON and writes it to the save file.
func (g *Game) saveToFile() error {
	path, err := savePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data := g.SaveData()
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Save writes the game state to disk and updates g.Status.
func (g *Game) Save() {
	if err := g.saveToFile(); err != nil {
		g.Status = SaveStatus{Code: SaveStatusSaveFailed, Err: err, SetAt: g.clock.Now()}
		return
	}
	g.Status = SaveStatus{Code: SaveStatusSaved, SetAt: g.clock.Now()}
}

// Load reads the save file and restores the game state in-place, updating g.Status.
// If no save file exists the call is a silent no-op (status is left unchanged).
func (g *Game) Load() {
	path, err := savePath()
	if err != nil {
		// WASM or config-dir unavailable — silent no-op.
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return // no save file yet — silent no-op
		}
		g.Status = SaveStatus{Code: SaveStatusLoadFailed, Err: err, SetAt: g.clock.Now()}
		return
	}
	var data SaveGameData
	if err := json.Unmarshal(b, &data); err != nil {
		g.Status = SaveStatus{Code: SaveStatusLoadFailed, Err: err, SetAt: g.clock.Now()}
		return
	}
	if err := g.LoadSaveData(data); err != nil {
		g.Status = SaveStatus{Code: SaveStatusLoadFailed, Err: err, SetAt: g.clock.Now()}
		return
	}
	g.Status = SaveStatus{Code: SaveStatusLoaded, SetAt: g.clock.Now()}
}

// Reset replaces the game state with a fresh new game and updates g.Status.
func (g *Game) Reset() {
	g.State = newState()
	g.Stores = NewStorageManager()
	g.Villagers = NewVillagerManager()
	g.Status = SaveStatus{Code: SaveStatusReset, SetAt: g.clock.Now()}
}
