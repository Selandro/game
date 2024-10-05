package menu

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"main.go/levels/level1"
)

type Menu struct {
	game        level1.GameInterface // Интерфейс для переключения уровней
	Player      *level1.Player
	cursorIndex int  // Индекс текущего поля для ввода (0 - имя, 1 - скин)
	ready       bool // Флаг, показывающий, что ввод завершен
}

// New инициализация меню
func New(game level1.GameInterface) *Menu {
	return &Menu{
		game:   game,
		Player: &level1.Player{},
	}
}

func (m *Menu) Update() error {
	// Убедимся, что ввод завершен
	if !m.ready {
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			// Проверяем завершение ввода имени и скина
			if m.cursorIndex == 1 && len(m.Player.Skin) > 0 {
				m.ready = true
			} else if m.cursorIndex == 0 && len(m.Player.Name) > 0 {
				m.cursorIndex = 1 // Переход на ввод скина
			}
		}

		// Удаление символов (Backspace)
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
			if m.cursorIndex == 0 && len(m.Player.Name) > 0 {
				m.Player.Name = m.Player.Name[:len(m.Player.Name)-1]
			} else if m.cursorIndex == 1 && len(m.Player.Skin) > 0 {
				m.Player.Skin = m.Player.Skin[:len(m.Player.Skin)-1]
			}
		}

		// Ввод имени и скина
		for _, char := range ebiten.InputChars() {
			if char == ' ' || char == '\n' || char == '\t' {
				continue // Игнорируем пробелы и спец. символы
			}
			if m.cursorIndex == 0 && len(m.Player.Name) < 20 {
				m.Player.Name += string(char)
			} else if m.cursorIndex == 1 && len(m.Player.Skin) < 20 {
				m.Player.Skin += string(char)
			}
		}
	} else {
		// Если ввод завершён, передаем имя и скин игрока
		// Теперь используем метод SetPlayerInfo через интерфейс
		m.game.SetPlayerInfo(m.Player.Name, m.Player.Skin)
		m.game.SwitchLevel(1)
	}

	return nil
}

// Draw отвечает за отрисовку меню
func (m *Menu) Draw(screen *ebiten.Image) {

	// Отображение текста для имени и скина
	var nameText string
	if m.cursorIndex == 0 {
		nameText = fmt.Sprintf("Enter Name: %s|", m.Player.Name)
	} else {
		nameText = fmt.Sprintf("Name: %s", m.Player.Name)
	}

	var skinText string
	if m.cursorIndex == 1 {
		skinText = fmt.Sprintf("Enter Skin: %s|", m.Player.Skin)
	} else {
		skinText = fmt.Sprintf("Skin: %s", m.Player.Skin)
	}

	// Сообщение о готовности
	var readyText string
	if m.ready {
		readyText = "Ready! Press Enter to start..."
	}

	// Отрисовка текста
	ebitenutil.DebugPrint(screen, nameText+"\n"+skinText+"\n"+readyText)
}

// Layout стандартный метод для задания размера окна
func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
