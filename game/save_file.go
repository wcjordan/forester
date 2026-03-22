package game

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

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

// SaveToFile serializes the game state to JSON and writes it to the save file.
func (g *Game) SaveToFile() error {
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

// LoadFromFile reads the save file and returns a restored Game.
// Returns an error if the file does not exist or cannot be parsed.
func LoadFromFile() (*Game, error) {
	path, err := savePath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data SaveGameData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	g := &Game{
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
		clock: RealClock{},
	}
	if err := g.LoadSaveData(data); err != nil {
		return nil, err
	}
	return g, nil
}
