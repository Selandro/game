package level1

import (
	"encoding/json"
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
	ID             int       `json:"id"`
	X              float64   `json:"x"`
	Y              float64   `json:"y"`
	PrevX          float64   // предыдущая X позиция для интерполяции
	PrevY          float64   // предыдущая Y позиция для интерполяции
	LastUpdateTime time.Time // время последнего обновления с сервера
}

type CapturePoint struct {
	X               float64   `json:"x"`
	Y               float64   `json:"y"`
	Radius          float64   `json:"radius"`
	IsCaptured      bool      `json:"isCaptured"`
	CapturingPlayer int       `json:"capturingPlayer"`
	CaptureStart    time.Time `json:"captureStart"`
	EnterTime       time.Time `json:"enterTime"`
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
}

type Level1 struct {
	game          GameInterface
	playerID      int
	playerX       float64
	playerY       float64
	capturePoints []CapturePoint
	players       []Player
	points1       int
	points2       int
	conn          *net.UDPConn
	done          chan struct{}
	lastUpdate    time.Time
	serverAddr    *net.UDPAddr
}

// New инициализирует уровень и подключается к серверу через UDP
func New(game GameInterface) *Level1 {
	// Настраиваем UDP соединение
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal("Ошибка при резолве адреса UDP:", err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", "localhost:8082") // Уникальный порт для первого клиента
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
			l.players[i].PrevX = l.players[i].X
			l.players[i].PrevY = l.players[i].Y
			l.players[i].LastUpdateTime = time.Now()

			// Обновляем позицию игрока
			l.playerX = player.X
			l.playerY = player.Y
		} else {
			// Для других игроков, просто обновляем их предыдущие координаты
			l.players[i].PrevX = l.players[i].X
			l.players[i].PrevY = l.players[i].Y
			l.players[i].LastUpdateTime = time.Now()
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
			"id": l.playerID,
			"x":  l.playerX,
			"y":  l.playerY,
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
	// Определяем флаг для отражения спрайта игрока
	var flipX bool

	// Определяем направление движения игрока (налево)
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		flipX = true
	}

	// Подготавливаем параметры для отрисовки спрайта игрока
	playerOp := &ebiten.DrawImageOptions{}
	if flipX {
		playerOp.GeoM.Scale(-1, 1) // Отражаем по оси X
	}

	// Отрисовываем спрайт игрока с правильной позицией
	sprites.PlayerSprite.Draw(screen, l.playerX, l.playerY, playerOp)

	// Отрисовка врагов
	for _, p := range l.players {
		if p.ID == l.playerID {
			continue
		}
		t := time.Since(p.LastUpdateTime).Seconds() / 0.2
		if t > 1 {
			t = 1
		}

		x := lerp(p.PrevX, p.X, t)
		y := lerp(p.PrevY, p.Y, t)
		// Подготавливаем параметры для отрисовки спрайта врага
		enemyOp := &ebiten.DrawImageOptions{}

		// Устанавливаем позицию врага
		sprites.EnemySprite.Draw(screen, x, y, enemyOp) // Отрисовка врага
	}

	// Отрисовка точек захвата
	for _, cp := range l.capturePoints {
		ebitenutil.DebugPrintAt(screen, "CP: X="+strconv.FormatFloat(cp.X, 'f', 1, 64)+" Y="+strconv.FormatFloat(cp.Y, 'f', 1, 64), int(cp.X), int(cp.Y)-20)

		if cp.Radius < 10 {
			cp.Radius = 10
		}

		drawCircle(screen, cp.X, cp.Y, cp.Radius, color.RGBA{255, 0, 0, 100})

		if cp.IsCaptured {
			if cp.CapturingPlayer == 1 {
				drawCircle(screen, cp.X, cp.Y, cp.Radius, color.RGBA{0, 255, 0, 100})
			} else if cp.CapturingPlayer == 2 {
				drawCircle(screen, cp.X, cp.Y, cp.Radius, color.RGBA{0, 0, 255, 100})
			}
		}

		if !cp.IsCaptured && !cp.EnterTime.IsZero() {
			progress := int(time.Since(cp.EnterTime).Seconds())
			ebitenutil.DebugPrintAt(screen, "Progress: "+strconv.Itoa(progress)+"s", int(cp.X), int(cp.Y)-40)
		}
	}

	ebitenutil.DebugPrint(screen, "Player 1 Points: "+strconv.Itoa(l.points1)+"\nPlayer 2 Points: "+strconv.Itoa(l.points2))
}

func drawCircle(screen *ebiten.Image, x, y, radius float64, clr color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x-radius, y-radius)
	img := ebiten.NewImage(int(2*radius), int(2*radius))
	img.Fill(color.Transparent)
	vector.DrawFilledCircle(img, float32(radius), float32(radius), float32(radius), clr, true)
	screen.DrawImage(img, op)
}

func (l *Level1) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
