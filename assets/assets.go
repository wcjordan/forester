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
	GrassTile  *ebiten.Image
	Treetop    *ebiten.Image
	Trunk      *ebiten.Image
	Grass      *ebiten.Image
	Dirt       *ebiten.Image
	House      *ebiten.Image
	Barrel     *ebiten.Image
	MaleWalk   *ebiten.Image
	FemaleWalk *ebiten.Image
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

	Treetop = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/treetop.png")
	Trunk = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/trunk.png")
	Grass = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/grass.png")
	Dirt = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/dirt.png")
	House = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/house.png")
	Barrel = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/barrel.png")

	MaleWalk = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/male_walkcycle.png")
	FemaleWalk = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/female_walkcycle.png")
}
