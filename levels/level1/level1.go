package level1

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	sprites "main.go/resourses/img"
)

type Player struct {
	ID             int                     `json:"id"`
	X              float64                 `json:"x"`
	Y              float64                 `json:"y"`
	PrevX          float64                 // Предыдущая X позиция для интерполяции
	PrevY          float64                 // Предыдущая Y позиция для интерполяции
	LastUpdateTime time.Time               // Время последнего обновления с сервера
	Name           string                  `json:"name"` // Добавляем JSON-тег для имени
	Skin           string                  `json:"skin"` // Добавляем JSON-тег для скина
	FlipX          bool                    `json:"flipX"`
	Sprite         *sprites.AnimatedSprite // Спрайт игрока
}

type CapturePoint struct {
	X                      float64   `json:"x"`
	Y                      float64   `json:"y"`
	Radius                 float64   `json:"radius"`
	IsCaptured             bool      `json:"isCaptured"`
	CapturingPlayer        int       `json:"capturingPlayer"`
	CaptureStart           time.Time `json:"captureStart"`
	EnterTime              time.Time `json:"enterTime"`
	CurrentCapturingPlayer int       `json:"currentCapturingPlayer"`
}

type GameState struct {
	Players       []Player       `json:"players"`
	CapturePoints []CapturePoint `json:"capturePoints"`
	Points1       int            `json:"points1"`
	Points2       int            `json:"points2"`
}

type GameInterface interface {
	SwitchLevel(level int)
	GetScale() float64
	SetPlayerInfo(name, skin string)
}

type Level1 struct {
	game          GameInterface
	playerID      int
	playerX       float64
	playerY       float64
	capturePoints []CapturePoint
	players       []Player
	FlipX         bool
	points1       int
	points2       int
	playerName    string
	playerSkin    string
	conn          *net.UDPConn
	done          chan struct{}
	lastUpdate    time.Time
	serverAddr    *net.UDPAddr
}

// New инициализирует уровень и подключается к серверу через UDP
func New(game GameInterface, playerName, playerSkin string) *Level1 {
	// Настраиваем UDP соединение
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal("Ошибка при резолве адреса UDP:", err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", "localhost:8081") // Уникальный порт для первого клиента
	if err != nil {
		log.Fatal("Ошибка при резолве адреса UDP:", err)
	}
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		log.Fatal("Ошибка подключения к UDP серверу:", err)
	}

	level := &Level1{
		game:       game,
		conn:       conn,
		serverAddr: serverAddr,
		done:       make(chan struct{}),
		playerID:   0, // Пока ID неизвестен
		playerName: playerName,
		playerSkin: playerSkin,
	}

	// Получение playerID от сервера
	level.requestPlayerID()

	go level.listenForUpdates()

	return level
}

func (l *Level1) requestPlayerID() {
	// Отправляем запрос на получение playerID
	initialMsg := map[string]interface{}{
		"request": "get_player_id",
		"name":    l.playerName, // Передаем имя игрока
		"skin":    l.playerSkin, // Передаем скин игрока
	}
	data, _ := json.Marshal(initialMsg)
	l.conn.Write(data)

	// Ожидаем ответ от сервера с playerID
	buffer := make([]byte, 2048)
	n, _, err := l.conn.ReadFrom(buffer)
	if err != nil {
		log.Println("Ошибка получения данных от сервера:", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		log.Println("Ошибка разбора данных от сервера:", err)
		return
	}

	// Сохраняем playerID, полученный от сервера
	if id, ok := response["id"].(float64); ok {
		l.playerID = int(id)
		log.Printf("Получен playerID: %d", l.playerID)
	}
}

// listenForUpdates получает обновления от сервера
func (l *Level1) listenForUpdates() {
	buffer := make([]byte, 2048)
	for {
		n, _, err := l.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Ошибка при чтении данных от сервера:", err)
			return
		}

		var gameState GameState
		err = json.Unmarshal(buffer[:n], &gameState)
		if err != nil {
			log.Println("Ошибка при десериализации данных:", err)
			continue
		}

		// Обновляем состояние игры на основе полученных данных
		l.updateGameState(gameState)
	}
}

func (l *Level1) updateGameState(state GameState) {
	l.players = state.Players
	l.capturePoints = state.CapturePoints
	l.points1 = state.Points1
	l.points2 = state.Points2

	// Обновляем координаты только для своего игрока
	for i, player := range state.Players {
		if player.ID == l.playerID {
			// Сохраняем предыдущую позицию
			continue
		} else {
			// Для других игроков, просто обновляем их предыдущие координаты
			l.players[i].PrevX = l.players[i].X
			l.players[i].PrevY = l.players[i].Y
			l.players[i].LastUpdateTime = time.Now()
			l.players[i].Name = player.Name // Обновляем имя
			l.players[i].Skin = player.Skin // Обновляем скин
			l.players[i].FlipX = player.FlipX
			fmt.Println(l.players[i].PrevX, l.players[i].PrevY, l.players[i].Name, l.players[i].Skin, l.players[i].FlipX)

		}

	}
}

func (l *Level1) Update() error {

	keys := map[string]bool{
		"w": ebiten.IsKeyPressed(ebiten.KeyW),
		"s": ebiten.IsKeyPressed(ebiten.KeyS),
		"a": ebiten.IsKeyPressed(ebiten.KeyA),
		"d": ebiten.IsKeyPressed(ebiten.KeyD),
	}

	speed := 10.0
	originalX, originalY := l.playerX, l.playerY

	if keys["w"] {
		l.playerY -= speed
	}
	if keys["s"] {
		l.playerY += speed
	}
	if keys["a"] {
		l.playerX -= speed
		l.FlipX = true
	}
	if keys["d"] {
		l.playerX += speed
	}

	// Если позиция изменилась, отправляем данные на сервер
	if originalX != l.playerX || originalY != l.playerY {
		l.sendPositionUpdate()
	}

	if ebiten.IsKeyPressed(ebiten.KeyP) {
		l.sendAction("pull")
	}
	if ebiten.IsKeyPressed(ebiten.KeyO) {
		l.sendAction("push")
	}

	return nil
}

func (l *Level1) sendPositionUpdate() {
	if time.Since(l.lastUpdate) > 10*time.Millisecond {
		// Формируем данные для отправки
		data := map[string]interface{}{
			"id":    l.playerID,
			"x":     l.playerX,
			"y":     l.playerY,
			"flipX": l.FlipX, // Добавляем состояние FlipX
		}

		// Сериализуем данные в JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Println("Ошибка сериализации данных:", err)
			return
		}

		// Отправляем сериализованные данные через UDP
		_, err = l.conn.Write(jsonData)
		if err != nil {
			log.Println("Ошибка отправки данных через UDP:", err)
			return
		}

		// Обновляем время последней отправки
		l.lastUpdate = time.Now()
	}
}

func (l *Level1) sendAction(action string) {
	// Формируем данные для отправки
	data := map[string]interface{}{
		"id":     l.playerID,
		"action": action,
	}

	// Сериализуем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Ошибка сериализации данных:", err)
		return
	}

	// Отправляем данные через UDP
	_, err = l.conn.Write(jsonData)
	if err != nil {
		log.Println("Ошибка отправки данных через UDP:", err)
		return
	}
}
func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func lerp(start, end, t float64) float64 {
	t = easeInOut(t)
	return start + (end-start)*t
}

func (l *Level1) Draw(screen *ebiten.Image) {
	scale := l.game.GetScale() // Получаем масштаб

	// Определяем флаг для отражения спрайта игрока

	// Определяем направление движения игрока (налево)
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		l.FlipX = true
	} else {
		l.FlipX = false
	}

	// Подготавливаем параметры для отрисовки спрайта игрока
	playerOp := &ebiten.DrawImageOptions{}
	if l.FlipX {
		playerOp.GeoM.Scale(-1, 1) // Отражаем по оси X
	}

	// Масштабируем координаты игрока только для отрисовки
	scaledPlayerX := l.playerX * scale
	scaledPlayerY := l.playerY * scale

	// Отрисовываем спрайт игрока с правильной позицией
	sprites.Sprites[l.playerSkin].Draw(screen, scaledPlayerX, scaledPlayerY, scale, l.FlipX, playerOp)

	// Отрисовка врагов
	for _, p := range l.players {
		if p.ID == l.playerID {
			continue
		}

		t := time.Since(p.LastUpdateTime).Seconds() / 0.2
		if t > 1 {
			t = 1
		}

		// Масштабируем координаты только для отрисовки
		scaledPrevX := p.PrevX * scale
		scaledX := p.X * scale
		scaledPrevY := p.PrevY * scale
		scaledY := p.Y * scale

		x := lerp(scaledPrevX, scaledX, t)
		y := lerp(scaledPrevY, scaledY, t)

		// Подготавливаем параметры для отрисовки спрайта врага
		enemyOp := &ebiten.DrawImageOptions{}
		if p.FlipX {
			enemyOp.GeoM.Scale(-1, 1) // Отражаем по оси X
		}
		// Устанавливаем позицию врага
		sprites.Sprites[p.Skin].Draw(screen, x, y, scale, p.FlipX, enemyOp) // Отрисовка врага
	}

	for _, cp := range l.capturePoints {
		// Отображение информации о точке захвата
		cpX := cp.X * scale // Масштабируем координаты захватной точки
		cpY := cp.Y * scale

		ebitenutil.DebugPrintAt(screen, "CP: X="+strconv.FormatFloat(cp.X, 'f', 1, 64)+" Y="+strconv.FormatFloat(cp.Y, 'f', 1, 64), int(cpX), int(cpY)-int(20*scale))

		// Масштабируем радиус захватной точки
		if cp.Radius < 10 {
			cp.Radius = 10
		}
		radius := cp.Radius * scale

		// Основной круг точки захвата (красный, если не захвачена)
		if !cp.IsCaptured {
			drawCircleOutlineWithEffects(screen, cpX, cpY, radius, color.RGBA{255, 0, 0, 100})
		} else {
			// Отображаем цвет игрока, который владеет точкой
			playerColor := getPlayerColor(cp.CapturingPlayer)
			drawCircleOutlineWithEffects(screen, cpX, cpY, radius, playerColor)
		}

		// Проверяем, захватывается ли точка
		if cp.CurrentCapturingPlayer != 0 {
			if cp.CurrentCapturingPlayer == cp.CapturingPlayer {
				continue
			}

			// Прогресс захвата для текущего игрока
			progress := time.Since(cp.EnterTime).Seconds() / 5.0 // Захват занимает 5 секунд
			if progress > 1 {
				progress = 1
			}

			// Получаем цвет игрока, который сейчас захватывает
			capturingPlayerColor := getPlayerColor(cp.CurrentCapturingPlayer)

			// Рисуем растущий круг цвета игрока, который захватывает точку
			animatedRadius := radius * progress
			drawCircleOutlineWithEffects(screen, cpX, cpY, animatedRadius, capturingPlayerColor)

			// Информация о прогрессе
			progressText := fmt.Sprintf("Progress: %.0f%%", progress*100)
			ebitenutil.DebugPrintAt(screen, progressText, int(cpX), int(cpY)-int(40*scale))

			// Если захват завершён
			if progress == 1 {
				cp.IsCaptured = true
				cp.CapturingPlayer = cp.CurrentCapturingPlayer
				cp.EnterTime = time.Time{}
			}
		}

		// Если игрок пытается захватить точку
		if cp.CurrentCapturingPlayer == cp.CapturingPlayer {
			continue
		}
	}

	// Отображаем текст с учётом масштаба
	ebitenutil.DebugPrint(screen, "Player 1 Points: "+strconv.Itoa(l.points1)+"\nPlayer 2 Points: "+strconv.Itoa(l.points2))
}

// Функция для рисования контура круга с градиентом и свечением
func drawCircleOutlineWithEffects(screen *ebiten.Image, x, y, radius float64, clr color.Color) {
	if radius <= 0 {
		return // Не рисуем круг с некорректным радиусом
	}

	imgWidth := int(2 * radius)
	imgHeight := int(2 * radius)

	if imgWidth <= 0 || imgHeight <= 0 {
		return // Проверка на корректные размеры изображения
	}

	img := ebiten.NewImage(imgWidth, imgHeight)
	img.Fill(color.Transparent)

	// Добавляем эффект свечения через прозрачность и градиент
	for r := radius; r > radius-10; r-- { // Градиент по краю круга
		alpha := uint8(255 * (r / radius)) // Прозрачность градиента по радиусу
		gradientColor := color.RGBA{uint8(clr.(color.RGBA).R), uint8(clr.(color.RGBA).G), uint8(clr.(color.RGBA).B), alpha}
		vector.StrokeCircle(img, float32(radius), float32(radius), float32(r), 2, gradientColor, true) // Рисуем градиентный контур
	}

	// Рисуем основной контур круга
	vector.StrokeCircle(img, float32(radius), float32(radius), float32(radius), 1, clr, true) // 3 - толщина линии

	// Эффект свечения
	glowRadius := radius + 2 // Радиус свечения немного больше самого круга
	for r := radius; r < glowRadius; r++ {
		alpha := uint8(100 * ((glowRadius - r) / glowRadius)) // Прозрачность по краям свечения
		glowColor := color.RGBA{uint8(clr.(color.RGBA).R), uint8(clr.(color.RGBA).G), uint8(clr.(color.RGBA).B), alpha}
		vector.StrokeCircle(img, float32(radius), float32(radius), float32(r), 1, glowColor, true)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x-radius, y-radius)
	screen.DrawImage(img, op)
}

// Функция для получения уникального цвета игрока
func getPlayerColor(playerID int) color.Color {
	// Задаём фиксированный набор цветов
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Красный
		color.RGBA{0, 255, 0, 255},   // Зеленый
		color.RGBA{0, 0, 255, 255},   // Синий
		color.RGBA{255, 255, 0, 255}, // Желтый
		color.RGBA{255, 165, 0, 255}, // Оранжевый
		color.RGBA{128, 0, 128, 255}, // Фиолетовый
		// Добавьте больше цветов при необходимости
	}

	// Используем взятие остатка, чтобы ID игрока не превышал размер массива цветов
	return colors[playerID%len(colors)]
}

func (l *Level1) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
