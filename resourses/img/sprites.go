package sprites

import (
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	Image *ebiten.Image
}

type AnimatedSprite struct {
	Frames   []*ebiten.Image // Список кадров анимации
	Current  int             // Индекс текущего кадра
	Interval int             // Интервал смены кадров
	Timer    int             // Таймер для смены кадров
}

var (
	PlayerSprite *AnimatedSprite
	EnemySprite  *AnimatedSprite
	Sprites      map[string]*AnimatedSprite = make(map[string]*AnimatedSprite) // Инициализируем карту
)

func LoadSprites() error {
	var err error

	// Загружаем раскадровку спрайтов (6 строк и 8 столбцов)
	sheet, err := loadSprite("resourses/img/sprites/01Knight.png")
	if err != nil {
		return err
	}

	// Используем вторую строку для анимации бега (всего 8 кадров)
	frames, err := sliceSpriteSheet(sheet.Image, 6, 8, 2)
	if err != nil {
		return err
	}

	// Создаем анимированный спрайт для игрока
	Sprites["01Knight"] = &AnimatedSprite{
		Frames:   frames,
		Current:  0,
		Interval: 30,
		Timer:    0,
	}

	// Загружаем анимационный спрайт врага
	EnemySprite, err = loadAnimatedSprite([]string{
		"resourses/img/sprites/character_run_01.png",
		"resourses/img/sprites/character_run_12.png",
		"resourses/img/sprites/character_run_23.png",
		"resourses/img/sprites/character_run_34.png",
	})
	if err != nil {
		return err
	}

	return nil
}

// Загрузка анимационного спрайта
func loadAnimatedSprite(paths []string) (*AnimatedSprite, error) {
	frames := make([]*ebiten.Image, len(paths))
	for i, path := range paths {
		img, err := loadSprite(path)
		if err != nil {
			return nil, err
		}
		frames[i] = img.Image
	}
	return &AnimatedSprite{Frames: frames, Current: 0, Interval: 30, Timer: 0}, nil
}
func sliceSpriteSheet(sheet *ebiten.Image, rows, cols, targetRow int) ([]*ebiten.Image, error) {
	frames := []*ebiten.Image{}
	sheetWidth, sheetHeight := sheet.Size()

	frameWidth := sheetWidth / cols
	frameHeight := sheetHeight / rows

	// Обрезаем спрайты из целевой строки (targetRow), начиная с 0
	for col := 0; col < cols; col++ {
		x := col * frameWidth
		y := (targetRow - 1) * frameHeight // Индексация строк с 0, поэтому вычитаем 1

		// Создаем подизображение для каждого кадра
		frame := sheet.SubImage(image.Rect(x, y, x+frameWidth, y+frameHeight)).(*ebiten.Image)
		frames = append(frames, frame)
	}

	return frames, nil
}
func LoadSprites1() error {
	var err error

	// Загружаем раскадровку спрайтов (6 строк и 8 столбцов)
	sheet, err := loadSprite("resourses/img/sprites/01Knight.png")
	if err != nil {
		return err
	}

	// Используем вторую строку для анимации бега (всего 8 кадров)
	frames, err := sliceSpriteSheet(sheet.Image, 6, 8, 2)
	if err != nil {
		return err
	}

	// Создаем анимированный спрайт для игрока
	Sprites["01Knight"] = &AnimatedSprite{
		Frames:   frames,
		Current:  0,
		Interval: 30,
		Timer:    0,
	}

	return nil
}

func loadSprite(path string) (*Sprite, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	ebitenImg := ebiten.NewImageFromImage(img)
	return &Sprite{Image: ebitenImg}, nil
}

// Метод для отрисовки анимационного спрайта с учетом отражения и масштаба
func (s *AnimatedSprite) Draw(screen *ebiten.Image, x, y, scale float64, flipX bool, op *ebiten.DrawImageOptions) {
	s.Timer++
	if s.Timer >= s.Interval {
		s.Timer = 0
		s.Current = (s.Current + 1) % len(s.Frames) // Переход к следующему кадру
	}

	// Обнуляем матрицу перед каждым кадром, чтобы избежать накопления трансляций
	op.GeoM.Reset()

	// Масштабируем изображение по Y
	op.GeoM.Scale(scale, scale)

	// Проверяем направление для отражения по X
	if flipX {
		// Отражаем по X, то есть масштабируем по X в отрицательном направлении
		op.GeoM.Scale(-scale, scale)
	} else {
		// Масштабируем по X в положительном направлении (нормальное изображение)
		op.GeoM.Scale(scale, scale)
	}

	// Получаем текущий кадр
	frame := s.Frames[s.Current]

	// Получаем размеры текущего кадра
	frameWidth, frameHeight := frame.Size()

	// Рассчитываем смещение для центрирования
	scaledFrameWidth := float64(frameWidth) * scale
	scaledFrameHeight := float64(frameHeight) * scale

	// Если изображение отражено, корректируем смещение по X
	if flipX {
		// Если отражено, сдвигаем изображение на его ширину
		op.GeoM.Translate(scaledFrameWidth, 0)
	}

	// Центрируем спрайт относительно координат (x, y)
	op.GeoM.Translate(-scaledFrameWidth/2, -scaledFrameHeight/2)

	// Перемещаем спрайт к заданным координатам x, y
	op.GeoM.Translate(x, y)

	// Рисуем текущий кадр
	screen.DrawImage(frame, op)
}

// Метод для отрисовки обычного спрайта
func (s *Sprite) Draw(screen *ebiten.Image, x, y float64) {
	op := &ebiten.DrawImageOptions{}
	// Смещаем на центр спрайта
	spriteWidth, spriteHeight := float64(s.Image.Bounds().Max.X), float64(s.Image.Bounds().Max.Y)
	op.GeoM.Translate(x-spriteWidth/2, y-spriteHeight/2) // Центрируем по координатам
	screen.DrawImage(s.Image, op)
}
