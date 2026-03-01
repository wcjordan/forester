// Package gametest provides shared StructureDef stubs for tests across the
// game module (game/geom, etc.). It lives under game/internal so it is
// accessible to all packages rooted at forester/game but not to callers
// outside that subtree.
package gametest

import (
	"time"

	"forester/game/core"
	"forester/game/geom"
)

// StructureType constants mirror the values defined in game/structures.
// Defined here as string literals so this package does not import game/structures
// (which imports game, which would create a cycle for package game tests).
const (
	FoundationLogStorage core.StructureType = "foundation_log_storage"
	LogStorage           core.StructureType = "log_storage"
	FoundationHouse      core.StructureType = "foundation_house"
	House                core.StructureType = "house"
)

// LogStorageDef is a minimal game.StructureDef stub that mimics a 4×4 log
// storage. ShouldSpawn always returns false; initial spawning is owned by
// story beats.
type LogStorageDef struct{}

// FoundationType implements game.StructureDef.
func (LogStorageDef) FoundationType() core.StructureType { return FoundationLogStorage }

// BuiltType implements game.StructureDef.
func (LogStorageDef) BuiltType() core.StructureType { return LogStorage }

// Footprint implements game.StructureDef.
func (LogStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildCost implements game.StructureDef.
func (LogStorageDef) BuildCost() int { return 20 }

// ShouldSpawn implements game.StructureDef.
func (LogStorageDef) ShouldSpawn(_ core.StructureEnv) bool { return false }

// OnPlayerInteraction implements game.StructureDef.
func (LogStorageDef) OnPlayerInteraction(_ core.StructureEnv, _ geom.Point, _ time.Time) {}

// OnBuilt implements game.StructureDef.
func (LogStorageDef) OnBuilt(_ core.StructureEnv, _ geom.Point) {}

// WallDef is a minimal game.StructureDef stub that places a solid blocking
// rectangle of arbitrary size. Useful for pathfinding and routing obstacle tests.
type WallDef struct{ Width, Height int }

// FoundationType implements game.StructureDef.
func (d WallDef) FoundationType() core.StructureType { return LogStorage }

// BuiltType implements game.StructureDef.
func (d WallDef) BuiltType() core.StructureType { return LogStorage }

// Footprint implements game.StructureDef.
func (d WallDef) Footprint() (w, h int) { return d.Width, d.Height }

// BuildCost implements game.StructureDef.
func (d WallDef) BuildCost() int { return 0 }

// ShouldSpawn implements game.StructureDef.
func (d WallDef) ShouldSpawn(_ core.StructureEnv) bool { return false }

// OnPlayerInteraction implements game.StructureDef.
func (d WallDef) OnPlayerInteraction(_ core.StructureEnv, _ geom.Point, _ time.Time) {}

// OnBuilt implements game.StructureDef.
func (d WallDef) OnBuilt(_ core.StructureEnv, _ geom.Point) {}
