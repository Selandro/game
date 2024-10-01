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
)

func LoadSprites() error {
	var err error

	// Загружаем спрайты для игрока
	PlayerSprite, err = loadAnimatedSprite([]string{
		"resourses/img/sprites/character_run_0.png",
		"resourses/img/sprites/character_run_1.png",
		"resourses/img/sprites/character_run_2.png",
		"resourses/img/sprites/character_run_3.png",
	})
	if err != nil {
		return err
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

// Метод для отрисовки анимационного спрайта с учетом отражения
func (s *AnimatedSprite) Draw(screen *ebiten.Image, x, y float64, op *ebiten.DrawImageOptions) {
	s.Timer++
	if s.Timer >= s.Interval {
		s.Timer = 0
		s.Current = (s.Current + 1) % len(s.Frames) // Переход к следующему кадру
	}

	// Получаем текущие размеры кадра
	frameWidth, frameHeight := float64(s.Frames[s.Current].Bounds().Max.X), float64(s.Frames[s.Current].Bounds().Max.Y)

	// Корректируем смещение по X, если изображение отражено
	if op.GeoM.Element(0, 0) < 0 { // Проверяем отражение по оси X
		op.GeoM.Translate(frameWidth, 0) // Сдвигаем изображение вправо на его ширину
	}

	// Смещаем на центр спрайта
	op.GeoM.Translate(x-frameWidth/2, y-frameHeight/2-60)

	// Рисуем текущий кадр
	screen.DrawImage(s.Frames[s.Current], op)
}

// Метод для отрисовки обычного спрайта
func (s *Sprite) Draw(screen *ebiten.Image, x, y float64) {
	op := &ebiten.DrawImageOptions{}
	// Смещаем на центр спрайта
	spriteWidth, spriteHeight := float64(s.Image.Bounds().Max.X), float64(s.Image.Bounds().Max.Y)
	op.GeoM.Translate(x-spriteWidth/2, y-spriteHeight/2) // Центрируем по координатам
	screen.DrawImage(s.Image, op)
}
