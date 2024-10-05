package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"main.go/gamestate"
)

func main() {

	game := gamestate.NewGame()
	game.SwitchLevel(2) // Начальный уровень

	// Установка оконного режима
	ebiten.SetWindowSize(1600, 900)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) // Разрешаем изменение размера окна
	ebiten.SetWindowTitle("Level Switcher with Loading Screen")

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
