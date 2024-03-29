package main

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

var img []*ebiten.Image

//go:embed assets/*
var assets embed.FS

const (
	FieldSize     = 576
	ScreenWidth   = FieldSize
	ScreenHeight  = FieldSize + 64
	SampleRate    = 48000
	SymbolSize    = 64
	GridSize      = 192
	SymbolsInGrid = 3
)

type GameData struct {
	PlayerTurn   int
	PlayArea     [][]int  // 9x9 entries
	BlockedGrids [][]bool // 3x3 entries
	ClaimedGrids [][]int  // 3x3 entries
}

var touchCooldown = 0

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
	Player
	Winner
	Recycle
)

const (
	C int = iota
	CMajor
)

func init() {
	imageFilenames := []string{"assets/empty.png", "assets/circle.png", "assets/cross.png", "assets/grid.png", "assets/blocked.png", "assets/player.png", "assets/winner.png", "assets/recycle.png"}

	for _, filename := range imageFilenames {
		fileData, err := assets.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		imgReader := bytes.NewReader(fileData)

		loadedImage, _, err := image.Decode(imgReader)
		if err != nil {
			log.Fatal(err)
		}
		ebitenImage := ebiten.NewImageFromImage(loadedImage)

		img = append(img, ebitenImage)
	}
}

type Game struct {
	Players []*audio.Player
}

func handleMousePressed(g *Game) {
	areaLocationX := mouse.X / SymbolSize
	areaLocationY := mouse.Y / SymbolSize
	bigLocationX := areaLocationX / 3
	bigLocationY := areaLocationY / 3
	inGridX := areaLocationX % 3
	inGridY := areaLocationY % 3

	// Was mouse clicked inside the recycle button?
	if mouse.Y >= FieldSize && mouse.X >= 512 {
		initGameData()
	}

	// Was mouse clicked inside the play grid?
	if mouse.X >= 0 && mouse.X < FieldSize && mouse.Y >= 0 && mouse.Y < FieldSize && !gameData.BlockedGrids[bigLocationX][bigLocationY] && gameData.PlayArea[mouse.X/SymbolSize][mouse.Y/SymbolSize] == 0 {

		// Put a mark in an area and play a sound
		gameData.PlayArea[areaLocationX][areaLocationY] = gameData.PlayerTurn + 1
		gameData.PlayerTurn = (gameData.PlayerTurn + 1) % 2
		if len(g.Players) > 0 && !g.Players[C].IsPlaying() {
			g.Players[C].Rewind()
			g.Players[C].Play()
		}

		// Check for "small" winner
		for y := 0; y < 3; y++ {
			for x := 0; x < 3; x++ {
				status := checkWinner(extract3x3(gameData.PlayArea, x, y))
				if status != gameData.ClaimedGrids[x][y] {
					if len(g.Players) > 0 && !g.Players[CMajor].IsPlaying() {
						g.Players[CMajor].Rewind()
						g.Players[CMajor].Play()
					}
				}
				gameData.ClaimedGrids[x][y] = status
			}
		}

		// Block off grids
		if checkWinner(gameData.ClaimedGrids) > 0 {
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					gameData.BlockedGrids[x][y] = true
				}
			}
		} else if gameData.ClaimedGrids[inGridX][inGridY] == 0 {
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					gameData.BlockedGrids[x][y] = true
				}
			}
			gameData.BlockedGrids[inGridX][inGridY] = false
		} else {
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					gameData.BlockedGrids[x][y] = gameData.ClaimedGrids[x][y] > 0
				}
			}
		}

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
					gameData.BlockedGrids[x][y] = false
				}
			}
		}

	} else {
		// If the player clicks somewhere "wrong"
	}
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
			handleMousePressed(g)
			mouse.Depressed = false
		}
	}

	// Initialize an empty slice to store touch IDs
	var touchIDs []ebiten.TouchID

	// Append active touch IDs to the slice
	touchIDs = ebiten.AppendTouchIDs(touchIDs[:0])

	// Iterate through each active touch ID
	foundTouch := 0
	for _, id := range touchIDs {
		foundTouch++
		// Get the X and Y coordinates of the touch event
		mouse.X, mouse.Y = ebiten.TouchPosition(id)

	}
	if foundTouch > 0 && touchCooldown > 30 {
		handleMousePressed(g)
		touchCooldown = 0
	}
	touchCooldown++
	return nil
}

func createOptionsSymbol(x int, y int, scale int) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x*SymbolSize), float64(y*SymbolSize))
	op.GeoM.Scale(float64(scale), float64(scale))
	op.ColorScale.ScaleAlpha(1)
	return op
}

func createOptionsAbsolute(x int, y int, scale int) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.GeoM.Scale(float64(scale), float64(scale))
	op.ColorScale.ScaleAlpha(1)
	return op
}

func (g *Game) Draw(screen *ebiten.Image) {

	// Draw big grid
	op := createOptionsSymbol(0, 0, 3)
	screen.DrawImage(img[Grid], op)

	// Draw all circles and crosses
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			op := createOptionsSymbol(x, y, 1)
			screen.DrawImage(img[gameData.PlayArea[x][y]], op)
		}
	}

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			op := createOptionsSymbol(x*3, y*3, 1)

			// Draw small grid
			screen.DrawImage(img[Grid], op)

			// Draw greyed out block
			if gameData.BlockedGrids[x][y] {
				screen.DrawImage(img[Blocked], op)
			}
		}
	}

	// Draw claimed areas
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			op = createOptionsSymbol(x, y, 3)
			screen.DrawImage(img[gameData.ClaimedGrids[x][y]], op)
		}
	}

	if checkWinner(gameData.ClaimedGrids) > 0 {
		// Draw "winner:" text
		op = createOptionsAbsolute(8, 148, 4)
		screen.DrawImage(img[Winner], op)

		// Draw winning player icon
		op = createOptionsSymbol(4, 9, 1)
		screen.DrawImage(img[checkWinner(gameData.ClaimedGrids)], op)

	} else {
		// Draw "player:" text
		op = createOptionsAbsolute(8, 148, 4)
		screen.DrawImage(img[Player], op)

		// Draw current player icon
		op = createOptionsSymbol(4, 9, 1)
		screen.DrawImage(img[gameData.PlayerTurn+1], op)
	}
	op = createOptionsSymbol(8, 9, 1)
	screen.DrawImage(img[Recycle], op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func loadAudioFiles(filenames []string) ([]*audio.Player, error) {
	audioContext := audio.NewContext(44100)
	var players []*audio.Player

	for _, filename := range filenames {
		data, err := assets.ReadFile(filename) // Read from the embedded file system
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

func checkWinner(T [][]int) int {
	// Check rows
	for i := 0; i < 3; i++ {
		if T[i][0] == T[i][1] && T[i][1] == T[i][2] && T[i][0] != 0 {
			return T[i][0]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if T[0][i] == T[1][i] && T[1][i] == T[2][i] && T[0][i] != 0 {
			return T[0][i]
		}
	}

	// Check diagonals
	if T[0][0] == T[1][1] && T[1][1] == T[2][2] && T[0][0] != 0 {
		return T[0][0]
	}
	if T[0][2] == T[1][1] && T[1][1] == T[2][0] && T[0][2] != 0 {
		return T[0][2]
	}

	return 0 // No winner
}

// Function to extract a 3x3 slice from a 9x9 2D slice
func extract3x3(slice [][]int, xGridOffset int, yGridOffset int) [][]int {
	xOffset := xGridOffset * 3
	yOffset := yGridOffset * 3
	var newSlice [][]int
	for i := 0; i < 3; i++ {
		row := slice[i+xOffset][yOffset : 3+yOffset]
		newSlice = append(newSlice, row)
	}
	return newSlice
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
	gameData.BlockedGrids = [][]bool{{false, false, false}, {false, false, false}, {false, false, false}}
	gameData.ClaimedGrids = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
}

func main() {
	initGameData()
	audioFiles := []string{"assets/c.wav", "assets/c-major.wav"}
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
