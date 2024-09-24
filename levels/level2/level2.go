package level2

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameInterface interface {
	SwitchLevel(level int)
	GetScale() float64 // Метод для получения масштаба
}

type Level2 struct {
	game GameInterface
}

func New(game GameInterface) *Level2 {
	return &Level2{game: game}
}

func (l *Level2) Update() error {
	// Пример: переход на уровень 5 при нажатии на Enter
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		l.game.SwitchLevel(5)
	}
	return nil
}

func (l *Level2) Draw(screen *ebiten.Image) {
	scale := l.game.GetScale()
	screenWidth, screenHeight := ebiten.WindowSize()

	// Размер и позиция кнопки с учетом масштабирования
	buttonWidth := int(200 * scale)
	buttonHeight := int(50 * scale)
	buttonX := (screenWidth - buttonWidth) / 2
	buttonY := (screenHeight - buttonHeight) / 2

	// Отрисовка текста уровня
	ebitenutil.DebugPrint(screen, "Level 2")

	// Отрисовка кнопки (простого прямоугольника) с учетом масштабирования
	ebitenutil.DrawRect(screen, float64(buttonX), float64(buttonY), float64(buttonWidth), float64(buttonHeight), color.RGBA{0, 255, 0, 255})

	// Отрисовка текста на кнопке
	ebitenutil.DebugPrintAt(screen, "Go to Level 5", buttonX+10, buttonY+buttonHeight/2-10)
}

func (l *Level2) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
