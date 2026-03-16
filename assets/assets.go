package assets

import (
	"bytes"
	"embed"
	"image"
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

//go:embed sprites/lpc-thatched-roof-cottage
var cottageFS embed.FS

//go:embed sprites/lpc-windows-doors-v2
var windowsDoorsFS embed.FS

//go:embed sprites/container-v4_2
var containerFS embed.FS

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

	// ThatchedRoofSheet is the lpc-thatched-roof-cottage thatched-roof.png (512×512).
	ThatchedRoofSheet *ebiten.Image
	// CottageSheet is the lpc-thatched-roof-cottage cottage.png (512×512).
	CottageSheet *ebiten.Image
	// WindowsDoorsSheet is the lpc-windows-doors-v2 windows-doors.png (1024×1024).
	WindowsDoorsSheet *ebiten.Image
	// ContainerSheet is the container-v4_2 container.png (512×2048).
	ContainerSheet *ebiten.Image
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
	Dirt = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/dirt.png")
	House = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/house.png")
	Barrel = loadFromFS(tilesFS, "sprites/lpc_base_assets/tiles/barrel.png")

	Player = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier.png")
	Villager = loadFromFS(peopleFS, "sprites/lpc_base_assets/sprites/people/soldier_altcolor.png")
	TreesGreen = loadFromFS(lpcTreesFS, "sprites/lpc-trees/trees-green.png")
	TerrainSheet = loadFromFS(lpcTerrainsFS, "sprites/lpc-terrains/terrain-v7.png")

	// Grassland tile: 32×32 sprite at (224, 384) in the terrain spritesheet.
	GrassTile = ebiten.NewImageFromImage(TerrainSheet.SubImage(image.Rect(224, 384, 224+32, 384+32)))

	ThatchedRoofSheet = loadFromFS(cottageFS, "sprites/lpc-thatched-roof-cottage/thatched-roof.png")
	CottageSheet = loadFromFS(cottageFS, "sprites/lpc-thatched-roof-cottage/cottage.png")
	WindowsDoorsSheet = loadFromFS(windowsDoorsFS, "sprites/lpc-windows-doors-v2/windows-doors.png")
	ContainerSheet = loadFromFS(containerFS, "sprites/container-v4_2/container.png")

	img, _, err := image.Decode(bytes.NewReader(playerSheetData))
	if err != nil {
		panic("assets: cannot decode player-spritesheet.png: " + err.Error())
	}
	PlayerSheet = ebiten.NewImageFromImage(img)
}
