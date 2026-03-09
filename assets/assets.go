package assets

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/png"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

//go:embed sprites/lpc_base_assets/tiles
var tilesFS embed.FS

//go:embed sprites/lpc_base_assets/sprites/people
var peopleFS embed.FS

//go:embed sprites/lpc-trees
var lpcTreesFS embed.FS

//go:embed sprites/lpc-terrains
var lpcTerrainsFS embed.FS

//go:embed sprites/player-spritesheet.png
var playerSheetData []byte

// Sprite sheet images loaded at init time.
var (
	GrassTile *ebiten.Image
	Dirt      *ebiten.Image
	House     *ebiten.Image
	Barrel    *ebiten.Image
	Player    *ebiten.Image
	Villager  *ebiten.Image
	// TreesGreen is the LPC Trees Mega-Pack green variant sheet (1024×1024).
	TreesGreen *ebiten.Image
	// TerrainSheet is the lpc-terrains v7 sheet (1024×2048, 32×32 tiles, 32 columns).
	TerrainSheet *ebiten.Image
	// PlayerSheet is the full Universal LPC character spritesheet for the player.
	// Contains walk, slash, and thrust animation rows at 64×64 px per frame.
	PlayerSheet *ebiten.Image
)

func loadFromFS(fs embed.FS, path string) *ebiten.Image {
	data, err := fs.ReadFile(path)
	if err != nil {
		panic("assets: cannot read " + path + ": " + err.Error())
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic("assets: cannot decode " + path + ": " + err.Error())
	}
	return ebiten.NewImageFromImage(img)
}

func init() {
	GrassTile = ebiten.NewImage(32, 32)
	GrassTile.Fill(color.RGBA{R: 0x7E, G: 0xC8, B: 0x50, A: 0xFF})

	Dirt = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/dirt.png")
	House = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/house.png")
	Barrel = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/barrel.png")

	Player = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier.png")
	Villager = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier_altcolor.png")
	TreesGreen = loadFromFS(lpcTreesFS, "sprites/lpc-trees/trees-green.png")
	TerrainSheet = loadFromFS(lpcTerrainsFS, "sprites/lpc-terrains/terrain-v7.png")

	img, _, err := image.Decode(bytes.NewReader(playerSheetData))
	if err != nil {
		panic("assets: cannot decode player-spritesheet.png: " + err.Error())
	}
	PlayerSheet = ebiten.NewImageFromImage(img)
}
