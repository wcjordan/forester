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

// Sprite sheet images loaded at init time.
var (
	GrassTile   *ebiten.Image
	Treetop     *ebiten.Image
	Trunk       *ebiten.Image
	Grass       *ebiten.Image
	Dirt        *ebiten.Image
	House       *ebiten.Image
	Barrel      *ebiten.Image
	Player      *ebiten.Image
	Villager    *ebiten.Image
	TroddenPath *ebiten.Image
	Road        *ebiten.Image
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

	TroddenPath = ebiten.NewImage(32, 32)
	TroddenPath.Fill(color.RGBA{R: 0xC8, G: 0xA0, B: 0x60, A: 0xFF}) // sandy tan

	Road = ebiten.NewImage(32, 32)
	Road.Fill(color.RGBA{R: 0x90, G: 0x78, B: 0x60, A: 0xFF}) // gray-brown packed dirt

	Treetop = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/treetop.png")
	Trunk = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/trunk.png")
	Grass = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/grass.png")
	Dirt = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/dirt.png")
	House = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/house.png")
	Barrel = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/barrel.png")

	Player = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier.png")
	Villager = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier_altcolor.png")
}
