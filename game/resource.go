package game

import (
	"fmt"
	"math/rand"
	"time"
)

// ResourceType identifies a storable resource.
type ResourceType int

// Resource types.
const (
	Wood ResourceType = iota
)

// ResourceDef describes the behavior of one resource type.
// Each resource is registered in the resourceRegistry via RegisterResource.
type ResourceDef interface {
	Type() ResourceType
	// Harvest is called each tick to harvest the resource from the world into the player's inventory.
	Harvest(env *Env, now time.Time)
	// Regrow is called each tick to advance the resource's world regeneration.
	Regrow(env *Env, rng *rand.Rand, now time.Time)
}

var resourceRegistry = map[ResourceType]ResourceDef{}

// RegisterResource adds a ResourceDef to the global registry.
// Call this from an init() function in an external package (e.g. game/resources).
// Panics on nil or if the resource type is already registered.
func RegisterResource(d ResourceDef) {
	if d == nil {
		panic("RegisterResource: def is nil")
	}
	if _, exists := resourceRegistry[d.Type()]; exists {
		panic(fmt.Sprintf("RegisterResource: ResourceType %d already registered", d.Type()))
	}
	resourceRegistry[d.Type()] = d
}

// IterateResources calls fn once for each registered ResourceDef.
// Iteration order is undefined.
func IterateResources(fn func(ResourceDef)) {
	for _, def := range resourceRegistry {
		fn(def)
	}
}
