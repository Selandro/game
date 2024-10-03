package menu

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Menu struct {
	game        GameInterface // Интерфейс для переключения уровней
	name        string        // Имя игрока
	skin        string        // Название скина
	cursorIndex int           // Индекс текущего поля для ввода (0 - имя, 1 - скин)
	ready       bool          // Флаг, показывающий, что ввод завершен
}

type GameInterface interface {
	SwitchLevel(level int, name string, skin string) // Метод переключения уровня
	GetScale() float64
}

// New инициализация меню
func New(game GameInterface) *Menu {
	return &Menu{
		game: game,
	}
}

// Update отвечает за обновление состояния меню
func (m *Menu) Update() error {
	if !m.ready {
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			// Если нажат Enter, проверить, завершен ли ввод
			if m.cursorIndex == 1 && len(m.skin) > 0 {
				m.ready = true
			} else if m.cursorIndex == 0 && len(m.name) > 0 {
				m.cursorIndex = 1 // Переходим к вводу скина
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
			if m.cursorIndex == 0 && len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			} else if m.cursorIndex == 1 && len(m.skin) > 0 {
				m.skin = m.skin[:len(m.skin)-1]
			}
		}

		// Считываем ввод имени и скина
		for _, char := range ebiten.InputChars() {
			if char == ' ' || char == '\n' || char == '\t' {
				continue // Игнорируем пробелы и спец. символы
			}
			if m.cursorIndex == 0 && len(m.name) < 20 {
				m.name += string(char)
			} else if m.cursorIndex == 1 && len(m.skin) < 20 {
				m.skin += string(char)
			}
		}
	} else {
		// После завершения ввода можно переключить уровень
		m.game.SwitchLevel(1, m.name, m.skin)
	}

	return nil
}

// Draw отрисовывает меню
func (m *Menu) Draw(screen *ebiten.Image) {

	// Текст меню
	var nameText string
	if m.cursorIndex == 0 {
		nameText = fmt.Sprintf("Enter Name: %s|", m.name)
	} else {
		nameText = fmt.Sprintf("Name: %s", m.name)
	}

	var skinText string
	if m.cursorIndex == 1 {
		skinText = fmt.Sprintf("Enter Skin: %s|", m.skin)
	} else {
		skinText = fmt.Sprintf("Skin: %s", m.skin)
	}

	// Сообщение об успешном вводе
	var readyText string
	if m.ready {
		readyText = "Ready! Press Enter to start..."
	}

	// Печать текста на экране
	ebitenutil.DebugPrint(screen, nameText+"\n"+skinText+"\n"+readyText)
}

// Layout — стандартный метод для размеров экрана
func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
