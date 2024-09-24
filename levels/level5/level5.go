package level5

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameInterface interface {
	SwitchLevel(level int)
	GetScale() float64 // Метод для получения масштаба
}

type Level5 struct {
	game GameInterface
}

func New(game GameInterface) *Level5 {
	return &Level5{game: game}
}

func (l *Level5) Update() error {
	// Пример: переход на уровень 1 при нажатии на клавишу '1'
	if ebiten.IsKeyPressed(ebiten.Key1) {
		l.game.SwitchLevel(1)
	}
	return nil
}

func (l *Level5) Draw(screen *ebiten.Image) {
	scale := l.game.GetScale()
	screenWidth, screenHeight := ebiten.WindowSize()

	// Размер и позиция кнопки с учетом масштабирования
	buttonWidth := int(200 * scale)
	buttonHeight := int(50 * scale)
	buttonX := (screenWidth - buttonWidth) / 2
	buttonY := (screenHeight - buttonHeight) / 2

	// Отрисовка текста уровня
	ebitenutil.DebugPrint(screen, "Level 5")

	// Отрисовка кнопки (простого прямоугольника) с учетом масштабирования
	ebitenutil.DrawRect(screen, float64(buttonX), float64(buttonY), float64(buttonWidth), float64(buttonHeight), color.RGBA{0, 0, 255, 255})

	// Отрисовка текста на кнопке
	ebitenutil.DebugPrintAt(screen, "Go to Level 1", buttonX+10, buttonY+buttonHeight/2-10)
}

func (l *Level5) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
