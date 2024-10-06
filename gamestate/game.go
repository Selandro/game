package gamestate

import (
	"log"
	"math"
	"time"

	"main.go/levels/level1"
	"main.go/levels/level5"
	"main.go/levels/menu"
	sprites "main.go/resourses/img"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameState int

const (
	Playing GameState = iota
	Loading
)

type Game struct {
	currentLevel ebiten.Game
	nextLevel    int
	state        GameState
	scale        float64
	loadingImage *ebiten.Image // Поле для хранения изображения загрузочного экрана
	playerName   string        // Поле для имени игрока
	playerSkin   string        // Поле для скина игрока
}

func NewGame() *Game {
	// Загрузка изображения для экрана загрузки
	loadingImage, _, err := ebitenutil.NewImageFromFile("gamestate/loadscreen.png")
	if err != nil {
		panic(err) // Обработка ошибки загрузки изображения
	}

	return &Game{
		loadingImage: loadingImage, // Инициализация изображения загрузочного экрана
	}
}
func (g *Game) SetPlayerInfo(name, skin string) {
	g.playerName = name
	g.playerSkin = skin
}
func (g *Game) Update() error {
	switch g.state {
	case Playing:
		if g.currentLevel != nil {
			return g.currentLevel.Update()
		}
	case Loading:
		g.loadNextLevel()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case Playing:
		if g.currentLevel != nil {
			g.currentLevel.Draw(screen)
		} else {
			ebitenutil.DebugPrint(screen, "No Level Loaded")
		}
	case Loading:
		if g.loadingImage != nil {
			screenWidth := screen.Bounds().Dx()
			screenHeight := screen.Bounds().Dy()
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(screenWidth)/float64(g.loadingImage.Bounds().Dx()), float64(screenHeight)/float64(g.loadingImage.Bounds().Dy()))
			screen.DrawImage(g.loadingImage, op)
			ebitenutil.DebugPrint(screen, "Loading...")
		} else {
			ebitenutil.DebugPrint(screen, "Loading...")
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	baseWidth, baseHeight := 1600, 900
	g.scale = math.Min(float64(outsideWidth)/float64(baseWidth), float64(outsideHeight)/float64(baseHeight))
	screenWidth = int(float64(baseWidth) * g.scale)
	screenHeight = int(float64(baseHeight) * g.scale)
	return screenWidth, screenHeight
}

func (g *Game) SwitchLevel(level int) {
	g.nextLevel = level
	g.state = Loading // Переход в состояние загрузки
}

func (g *Game) loadNextLevel() {
	time.Sleep(1 * time.Second) // Имитация загрузки

	switch g.nextLevel {
	case 1:

		g.currentLevel = level1.New(g, g.playerName, g.playerSkin)
	case 2:
		g.currentLevel = menu.New(g)
		if err := sprites.LoadSprites(); err != nil {
			log.Fatal("Ошибка загрузки спрайтов:", err)
		}
	case 5:
		g.currentLevel = level5.New(g)
	default:
		g.currentLevel = nil
	}
	g.state = Playing
}

func (g *Game) GetScale() float64 {
	return g.scale
}
