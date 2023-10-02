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

var playArea [9][9]int

type Mouse struct {
	X         int
	Y         int
	Depressed bool
}

var mouse = Mouse{0, 0, false}
var playerTurn int = 0

const (
	Empty int = iota
	Circle
	Cross
	Grid
)

func init() {
	image_filename := []string{"resources/empty.png", "resources/circle.png", "resources/cross.png", "resources/grid.png"}
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

func loadAudio(filename string, audioContext *audio.Context) (*audio.Player, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	byteReader := bytes.NewReader(data)
	decoded, err := wav.DecodeWithSampleRate(44100, byteReader)
	if err != nil {
		return nil, err
	}

	return audioContext.NewPlayer(decoded)
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
			if playArea[mouse.X/SymbolSize][mouse.Y/SymbolSize] == 0 {
				playArea[mouse.X/SymbolSize][mouse.Y/SymbolSize] = playerTurn + 1
				playerTurn = (playerTurn + 1) % 2
				if len(g.Players) > 0 && !g.Players[1].IsPlaying() {
					g.Players[1].Rewind()
					g.Players[1].Play()
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

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			op := createOptions(x*3, y*3, 1)
			screen.DrawImage(img[Grid], op)
		}
	}

	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			op := createOptions(x, y, 1)
			screen.DrawImage(img[playArea[x][y]], op)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {

	audioContext := audio.NewContext(44100)

	game := &Game{
		Players: []*audio.Player{}, // Initialize as empty slice
	}

	// Load multiple audio files
	player1, err := loadAudio("resources/wet.wav", audioContext)
	if err != nil {
		log.Fatal(err)
	}
	game.Players = append(game.Players, player1)

	player2, err := loadAudio("resources/stomp.wav", audioContext)
	if err != nil {
		log.Fatal(err)
	}
	game.Players = append(game.Players, player2)

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Super Tic-Tac-Toe")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
