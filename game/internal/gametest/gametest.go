// Package gametest provides shared StructureDef stubs for tests across the
// game module (game/geom, etc.). It lives under game/internal so it is
// accessible to all packages rooted at forester/game but not to callers
// outside that subtree.
//
// Note: game package internal tests (package game) cannot import this package
// because gametest imports game, which would create a cycle. Those tests keep
// local equivalents in story_test.go.
package gametest

import (
	"time"

	"forester/game"
	"forester/game/geom"
	"forester/game/structures"
)

// LogStorageDef is a minimal game.StructureDef stub that mimics a 4×4 log
// storage. ShouldSpawn always returns false; initial spawning is owned by
// story beats.
type LogStorageDef struct{}

// FoundationType implements game.StructureDef.
func (LogStorageDef) FoundationType() game.StructureType { return structures.FoundationLogStorage }

// BuiltType implements game.StructureDef.
func (LogStorageDef) BuiltType() game.StructureType { return structures.LogStorage }

// Footprint implements game.StructureDef.
func (LogStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildCost implements game.StructureDef.
func (LogStorageDef) BuildCost() int { return 20 }

// ShouldSpawn implements game.StructureDef.
func (LogStorageDef) ShouldSpawn(_ *game.Env) bool { return false }

// OnPlayerInteraction implements game.StructureDef.
func (LogStorageDef) OnPlayerInteraction(_ *game.Env, _ geom.Point, _ time.Time) {}

// OnBuilt implements game.StructureDef.
func (LogStorageDef) OnBuilt(_ *game.Env, _ geom.Point) {}

// WallDef is a minimal game.StructureDef stub that places a solid blocking
// rectangle of arbitrary size. Useful for pathfinding and routing obstacle tests.
type WallDef struct{ Width, Height int }

// FoundationType implements game.StructureDef.
func (d WallDef) FoundationType() game.StructureType { return structures.LogStorage }

// BuiltType implements game.StructureDef.
func (d WallDef) BuiltType() game.StructureType { return structures.LogStorage }

// Footprint implements game.StructureDef.
func (d WallDef) Footprint() (w, h int) { return d.Width, d.Height }

// BuildCost implements game.StructureDef.
func (d WallDef) BuildCost() int { return 0 }

// ShouldSpawn implements game.StructureDef.
func (d WallDef) ShouldSpawn(_ *game.Env) bool { return false }

// OnPlayerInteraction implements game.StructureDef.
func (d WallDef) OnPlayerInteraction(_ *game.Env, _ geom.Point, _ time.Time) {}

// OnBuilt implements game.StructureDef.
func (d WallDef) OnBuilt(_ *game.Env, _ geom.Point) {}
