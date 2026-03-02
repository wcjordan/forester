//go:build js

package main

// shouldRunTUI always returns false for WASM builds; bubbletea is not available.
func shouldRunTUI() bool { return false }

// runTUI is a no-op for WASM builds.
func runTUI() {}
