package level1

import (
	"image/color"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
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
	targetX       float64 // Новые целевые координаты
	targetY       float64
	capturePoints []CapturePoint
	players       []Player
	points1       int
	points2       int
	conn          *websocket.Conn
	done          chan struct{}
	lastUpdate    time.Time
}

func New(game GameInterface, playerID int) *Level1 {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Ошибка подключения к WebSocket серверу:", err)
	}

	l := &Level1{
		game:     game,
		playerID: playerID,
		playerX:  400,
		playerY:  400,
		targetX:  400,
		targetY:  400,
		conn:     conn,
		capturePoints: []CapturePoint{
			{X: 300, Y: 200, Radius: 50},
			{X: 800, Y: 600, Radius: 50},
		},
		points1:    0,
		points2:    0,
		done:       make(chan struct{}),
		lastUpdate: time.Now(),
	}

	go l.listenForUpdates()

	return l
}

func (l *Level1) listenForUpdates() {
	for {
		select {
		case <-l.done:
			return
		default:
			var gameState GameState
			err := l.conn.ReadJSON(&gameState)
			if err != nil {
				log.Println("Ошибка получения состояния игры:", err)
				return
			}

			l.players = gameState.Players
			l.capturePoints = gameState.CapturePoints
			l.points1 = gameState.Points1
			l.points2 = gameState.Points2

		}
	}
}

func (l *Level1) Update() error {
	// Локальная обработка нажатий клавиш
	keys := map[string]bool{
		"w": ebiten.IsKeyPressed(ebiten.KeyW),
		"s": ebiten.IsKeyPressed(ebiten.KeyS),
		"a": ebiten.IsKeyPressed(ebiten.KeyA),
		"d": ebiten.IsKeyPressed(ebiten.KeyD),
	}

	speed := 64.0 * (1.0 / 60)

	// Обновляем целевые координаты при нажатии клавиш
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

	// Отправляем обновления только каждые 100 мс
	if time.Since(l.lastUpdate) > 16*time.Millisecond {
		// Отправляем только координаты игрока на сервер
		data := map[string]interface{}{
			"id": l.playerID,
			"x":  l.playerX,
			"y":  l.playerY,
		}
		err := l.conn.WriteJSON(data)
		if err != nil {
			log.Println("Ошибка отправки данных:", err)
			return err
		}
		l.lastUpdate = time.Now()
	}

	// Обрабатываем действия (например, "притяжение" и "отталкивание")
	if ebiten.IsKeyPressed(ebiten.KeyP) {
		l.sendAction("pull")
	}
	if ebiten.IsKeyPressed(ebiten.KeyO) {
		l.sendAction("push")
	}

	return nil
}

func (l *Level1) sendAction(action string) {
	data := map[string]interface{}{
		"id":     l.playerID,
		"action": action,
	}
	err := l.conn.WriteJSON(data)
	if err != nil {
		log.Println("Ошибка отправки действия:", err)
	}
}

func (l *Level1) Draw(screen *ebiten.Image) {
	scale := l.game.GetScale()

	// Отрисовка всех игроков
	for _, p := range l.players {
		playerColor := color.RGBA{0, 0, 255, 255} // Цвет по умолчанию для других игроков
		if p.ID == l.playerID {
			playerColor = color.RGBA{0, 255, 0, 255} // Цвет игрока, если это наш
		}
		drawRect(screen, p.X, p.Y, 20*scale, 20*scale, playerColor)
	}

	// Отрисовка захватных точек
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

func drawRect(screen *ebiten.Image, x, y, width, height float64, clr color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x-width/2, y-height/2)
	img := ebiten.NewImage(int(width), int(height))
	img.Fill(clr)
	screen.DrawImage(img, op)
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

func (l *Level1) Close() {
	close(l.done)
	l.conn.Close()
}
