package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 1000
	screenHeight = 800
	cellSize     = 10

	gridWidth  = screenWidth / cellSize
	gridHeight = screenHeight / cellSize
)

type Rules struct {
	Birth   map[int]bool
	Survive map[int]bool
}

func ParseRules(rule string) Rules {

	r := Rules{
		Birth:   make(map[int]bool),
		Survive: make(map[int]bool),
	}

	parts := strings.Split(strings.ToUpper(rule), "/")

	for _, p := range parts {

		if len(p) < 2 {
			continue
		}

		mode := p[0]
		values := p[1:]

		for _, ch := range values {

			n, err := strconv.Atoi(string(ch))
			if err != nil {
				continue
			}

			if mode == 'B' {
				r.Birth[n] = true
			}

			if mode == 'S' {
				r.Survive[n] = true
			}
		}
	}

	return r
}

type Game struct {
	grid      [][]bool
	running   bool
	rules     Rules
	ruleStr   string
	inputMode bool
	inputText string
}

func NewGame(rule string) *Game {

	grid := make([][]bool, gridHeight)

	for y := range grid {
		grid[y] = make([]bool, gridWidth)
	}

	return &Game{
		grid:      grid,
		running:   false,
		rules:     ParseRules(rule),
		ruleStr:   rule,
		inputMode: false,
		inputText: "",
	}
}

func (g *Game) CountNeighbors(x, y int) int {

	count := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {

			if dx == 0 && dy == 0 {
				continue
			}

			nx := (x + dx + gridWidth) % gridWidth
			ny := (y + dy + gridHeight) % gridHeight

			if g.grid[ny][nx] {
				count++
			}
		}
	}

	return count
}

func (g *Game) Step() {

	next := make([][]bool, gridHeight)

	for y := range next {
		next[y] = make([]bool, gridWidth)
	}

	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {

			neighbors := g.CountNeighbors(x, y)

			if g.grid[y][x] {
				next[y][x] = g.rules.Survive[neighbors]
			} else {
				next[y][x] = g.rules.Birth[neighbors]
			}
		}
	}

	g.grid = next
}

func (g *Game) Update() error {

	if g.inputMode {

		for _, r := range ebiten.InputChars() {

			if unicode.IsLetter(r) ||
				unicode.IsDigit(r) ||
				r == '/' {

				g.inputText += string(r)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {

			if len(g.inputText) > 0 {
				g.inputText = g.inputText[:len(g.inputText)-1]
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {

			g.ruleStr = strings.ToUpper(g.inputText)
			g.rules = ParseRules(g.ruleStr)

			g.inputMode = false
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.inputMode = false
		}

		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.running = !g.running
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.Step()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) {

		for y := range g.grid {
			for x := range g.grid[y] {
				g.grid[y][x] = false
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {

		for y := range g.grid {
			for x := range g.grid[y] {
				g.grid[y][x] = rand.Intn(2) == 1
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {

		g.inputMode = true
		g.inputText = ""
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

		mx, my := ebiten.CursorPosition()

		x := mx / cellSize
		y := my / cellSize

		if x >= 0 && x < gridWidth &&
			y >= 0 && y < gridHeight {

			g.grid[y][x] = true
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {

		mx, my := ebiten.CursorPosition()

		x := mx / cellSize
		y := my / cellSize

		if x >= 0 && x < gridWidth &&
			y >= 0 && y < gridHeight {

			g.grid[y][x] = false
		}
	}

	if g.running {

		g.Step()
		time.Sleep(80 * time.Millisecond)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.Black)

	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {

			if g.grid[y][x] {

				ebitenutil.DrawRect(
					screen,
					float64(x*cellSize),
					float64(y*cellSize),
					cellSize,
					cellSize,
					color.White,
				)
			}
		}
	}

	gridColor := color.RGBA{50, 50, 50, 255}

	for x := 0; x < screenWidth; x += cellSize {

		ebitenutil.DrawLine(
			screen,
			float64(x),
			0,
			float64(x),
			screenHeight,
			gridColor,
		)
	}

	for y := 0; y < screenHeight; y += cellSize {

		ebitenutil.DrawLine(
			screen,
			0,
			float64(y),
			screenWidth,
			float64(y),
			gridColor,
		)
	}

	info := fmt.Sprintf(
		"SPACE-Start | N-Step | C-Clear | R-Random | TAB-Change Rules | Rules: %s",
		g.ruleStr,
	)

	text.Draw(
		screen,
		info,
		basicfont.Face7x13,
		10,
		20,
		color.RGBA{0, 255, 0, 255},
	)

	if g.inputMode {

		text.Draw(
			screen,
			"Enter rule: "+g.inputText,
			basicfont.Face7x13,
			10,
			50,
			color.RGBA{255, 255, 0, 255},
		)

		text.Draw(
			screen,
			"Example: B3/S23 or B36/S23",
			basicfont.Face7x13,
			10,
			70,
			color.RGBA{255, 255, 0, 255},
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	rand.Seed(time.Now().UnixNano())

	game := NewGame("B3/S23")

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Game of Life")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
