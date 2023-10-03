package main

import (
	"bytes"
	_ "image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var img []*ebiten.Image

const (
	ScreenWidth   = 576
	ScreenHeight  = 576
	SampleRate    = 48000
	SymbolSize    = 64
	GridSize      = 192
	SymbolsInGrid = 3
)

type GameData struct {
	PlayerTurn   int
	PlayArea     [][]int // 9x9 entries
	BlockedGrids [][]int // 3x3 entries
	ClaimedGrids [][]int // 3x3 entries
}

var gameData GameData

type Mouse struct {
	X         int
	Y         int
	Depressed bool
}

var mouse = Mouse{0, 0, false}

const (
	Empty int = iota
	Circle
	Cross
	Grid
	Blocked
)

func init() {
	image_filename := []string{"resources/empty.png", "resources/circle.png", "resources/cross.png", "resources/grid.png", "resources/blocked.png"}
	for _, filename := range image_filename {
		loadedImage, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		img = append(img, loadedImage)
	}
}

type Game struct {
	Players []*audio.Player
}

func (g *Game) Update() error {

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if !mouse.Depressed {
			mouse.X = mx
			mouse.Y = my
			mouse.Depressed = true
		}
	} else {
		if mouse.Depressed {
			areaLocationX := mouse.X / SymbolSize
			areaLocationY := mouse.Y / SymbolSize
			bigLocationX := areaLocationX / 3
			bigLocationY := areaLocationY / 3
			inGridX := areaLocationX % 3
			inGridY := areaLocationY % 3

			if gameData.BlockedGrids[bigLocationX][bigLocationY] == 0 && gameData.PlayArea[mouse.X/SymbolSize][mouse.Y/SymbolSize] == 0 {

				// Put a mark in an area and play a sound
				gameData.PlayArea[areaLocationX][areaLocationY] = gameData.PlayerTurn + 1
				gameData.PlayerTurn = (gameData.PlayerTurn + 1) % 2
				if len(g.Players) > 0 && !g.Players[1].IsPlaying() {
					g.Players[1].Rewind()
					g.Players[1].Play()
				}

				// Block off grids
				for y := 0; y < 3; y++ {
					for x := 0; x < 3; x++ {
						gameData.BlockedGrids[x][y] = 1
					}
				}
				gameData.BlockedGrids[inGridX][inGridY] = 0

				// Check if available slot is full - if so, unlock all slots
				emptySlots := 0
				for y := 0; y < 3; y++ {
					for x := 0; x < 3; x++ {
						if gameData.PlayArea[x+inGridX*3][y+inGridY*3] == 0 {
							emptySlots++
						}
					}
				}
				if emptySlots == 0 {
					for y := 0; y < 3; y++ {
						for x := 0; x < 3; x++ {
							gameData.BlockedGrids[x][y] = 0
						}
					}
				}

			} else {
				if len(g.Players) > 0 && !g.Players[0].IsPlaying() {
					g.Players[0].Rewind()
					g.Players[0].Play()
				}
			}
			mouse.Depressed = false
		}
	}
	return nil
}

func createOptions(x int, y int, scale int) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x*SymbolSize), float64(y*SymbolSize))
	op.GeoM.Scale(float64(scale), float64(scale))
	op.ColorScale.ScaleAlpha(1)
	return op
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := createOptions(0, 0, 3)
	screen.DrawImage(img[Grid], op)

	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			op := createOptions(x, y, 1)
			screen.DrawImage(img[gameData.PlayArea[x][y]], op)
		}
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			op := createOptions(x*3, y*3, 1)
			screen.DrawImage(img[Grid], op)

			if gameData.BlockedGrids[x][y] > 0 {
				screen.DrawImage(img[Blocked], op)
			}
		}
	}

	// Draw claimed areas
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			op = createOptions(x, y, 3)
			screen.DrawImage(img[gameData.ClaimedGrids[x][y]], op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func loadAudioFiles(filenames []string) ([]*audio.Player, error) {
	audioContext := audio.NewContext(44100)
	var players []*audio.Player
	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		byteReader := bytes.NewReader(data)
		decoded, err := wav.DecodeWithSampleRate(44100, byteReader)
		if err != nil {
			return nil, err
		}

		player, err := audioContext.NewPlayer(decoded)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, nil
}

func initGameData() {
	gameData.PlayArea = [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	gameData.BlockedGrids = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	gameData.ClaimedGrids = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
}

func main() {
	initGameData()
	gameData.ClaimedGrids[0][0] = 1
	audioFiles := []string{"resources/wet.wav", "resources/stomp.wav"}
	players, err := loadAudioFiles(audioFiles)
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		Players: []*audio.Player{},
	}

	game.Players = players

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Super Tic-Tac-Toe")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
