package menu

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"main.go/levels/level1"
	sprites "main.go/resourses/img"
)

type Menu struct {
	game              level1.GameInterface // Интерфейс для переключения уровней
	Player            *level1.Player
	cursorIndex       int       // Индекс текущего поля для ввода (0 - имя, 1 - скин)
	ready             bool      // Флаг, показывающий, что ввод завершен
	skinOptions       []string  // Список доступных скинов
	selectedSkinIndex int       // Индекс выбранного скина
	enterPressedTime  time.Time // Время последнего нажатия Enter для предотвращения мгновенного переключения
	arrowPressedTime  time.Time // Время последнего нажатия стрелки для предотвращения многократного переключения
}

// New инициализация меню
func New(game level1.GameInterface) *Menu {
	return &Menu{
		game:              game,
		Player:            &level1.Player{},
		skinOptions:       []string{"01Knight", "02Knight", "03Knight", "04Knight", "05Knight", "06Knight", "07Knight", "08Knight", "09Knight", "10Knight"},
		selectedSkinIndex: 0,          // По умолчанию выбран первый скин
		enterPressedTime:  time.Now(), // Инициализация времени последнего нажатия Enter
		arrowPressedTime:  time.Now(), // Инициализация времени последнего нажатия стрелок
	}
}

func (m *Menu) Update() error {
	// Убедимся, что ввод завершен
	if !m.ready {
		// Проверяем завершение ввода имени и скина, задержка для Enter
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && time.Since(m.enterPressedTime) > 300*time.Millisecond {
			m.enterPressedTime = time.Now() // Обновляем время нажатия Enter

			if m.cursorIndex == 1 && m.selectedSkinIndex >= 0 {
				// Завершаем выбор скина и переходим к игре
				m.Player.Skin = m.skinOptions[m.selectedSkinIndex]
				m.ready = true
			} else if m.cursorIndex == 0 && len(m.Player.Name) > 0 {
				// Переход к выбору скина после ввода имени
				m.cursorIndex = 1
			}
		}

		// Удаление символов (Backspace)
		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && m.cursorIndex == 0 && len(m.Player.Name) > 0 {
			m.Player.Name = m.Player.Name[:len(m.Player.Name)-1]
		}

		// Ввод имени
		if m.cursorIndex == 0 {
			for _, char := range ebiten.InputChars() {
				if char == ' ' || char == '\n' || char == '\t' {
					continue // Игнорируем пробелы и спец. символы
				}
				if len(m.Player.Name) < 20 {
					m.Player.Name += string(char)
				}
			}
		}

		// Выбор скина с помощью стрелок, с задержкой на повторное нажатие
		if m.cursorIndex == 1 {
			if time.Since(m.arrowPressedTime) > 300*time.Millisecond { // Задержка 300 мс между переключениями
				if ebiten.IsKeyPressed(ebiten.KeyArrowUp) && m.selectedSkinIndex > 0 {
					m.selectedSkinIndex--
					m.arrowPressedTime = time.Now() // Обновляем время последнего нажатия стрелки
				} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && m.selectedSkinIndex < len(m.skinOptions)-1 {
					m.selectedSkinIndex++
					m.arrowPressedTime = time.Now() // Обновляем время последнего нажатия стрелки
				}
			}
		}
	} else {
		// Если ввод завершён, передаем имя и скин игрока и переключаем на игру
		m.game.SetPlayerInfo(m.Player.Name, m.Player.Skin)
		m.game.SwitchLevel(1)
	}

	return nil
}

// Draw отвечает за отрисовку меню
func (m *Menu) Draw(screen *ebiten.Image) {

	// Отображение текста для имени
	var nameText string
	if m.cursorIndex == 0 {
		nameText = fmt.Sprintf("Enter Name: %s|", m.Player.Name)
	} else {
		nameText = fmt.Sprintf("Name: %s", m.Player.Name)
	}

	// Отображение текста для выбора скина
	var skinText string
	if m.cursorIndex == 1 {
		skinText = fmt.Sprintf("Select Skin: %s (use Up/Down to switch)", m.skinOptions[m.selectedSkinIndex])
	} else {
		skinText = fmt.Sprintf("Skin: %s", m.skinOptions[m.selectedSkinIndex])
	}

	// Сообщение о готовности
	var readyText string
	if m.ready {
		readyText = "Ready! Press Enter to start..."
	}

	// Отрисовка текста
	ebitenutil.DebugPrint(screen, nameText+"\n"+skinText+"\n"+readyText)

	// Отрисовка выбранного скина
	if sprite, ok := sprites.Sprites[m.skinOptions[m.selectedSkinIndex]]; ok {
		op := &ebiten.DrawImageOptions{}
		sprite.Draw(screen, 400, 300, 2.0, false, op) // Координаты и масштаб можно настроить
	}
}

// Layout стандартный метод для задания размера окна
func (m *Menu) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
