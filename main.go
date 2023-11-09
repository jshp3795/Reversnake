package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Position struct {
	posX int
	posY int
}

const (
	screenWidth  = 1080
	screenHeight = 720

	blockWidth    = 40
	blockHeight   = 40
	maxFrameCount = 4

	initialStarveDelay   = 5000
	starveDelayIncrement = 1000

	DIRECTION_NIL   = -1
	DIRECTION_LEFT  = 0
	DIRECTION_RIGHT = 1
	DIRECTION_UP    = 2
	DIRECTION_DOWN  = 3
)

var (
	textFont font.Face = nil
)

type Game struct {
	title        string
	timestamp    int64
	endTimestamp int64
	started      bool
	rand         *rand.Rand
	snake        Snake
	food         Food
	goldenFood   GoldenFood
	item         Item
}

type Snake struct {
	frameCount      int
	positions       []Position
	direction       int
	lastDirection   int
	hitWall         bool
	starveTimestamp int64
	starveDelay     int64
}

type Food struct {
	frameCount  int
	position    Position
	direction   int
	immuneUntil int64
	isGolden    bool
}

type GoldenFood struct {
	position Position
	visible  bool
	duration int64
}

type Item struct {
	used     bool
	duration int64
}

func (g *Game) Start() error {
	g.title = "SPACE를 눌러 게임을 시작하세요"
	g.timestamp = time.Now().UnixMilli()
	g.endTimestamp = g.timestamp
	g.started = true
	g.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	g.snake = Snake{
		frameCount:      0,
		positions:       []Position{{posX: 11, posY: 2}, {posX: 10, posY: 2}, {posX: 9, posY: 2}, {posX: 8, posY: 2}, {posX: 7, posY: 2}, {posX: 6, posY: 2}, {posX: 5, posY: 2}, {posX: 4, posY: 2}, {posX: 3, posY: 2}, {posX: 2, posY: 2}},
		direction:       DIRECTION_RIGHT,
		lastDirection:   DIRECTION_RIGHT,
		hitWall:         false,
		starveTimestamp: g.timestamp + initialStarveDelay,
		starveDelay:     initialStarveDelay}
	g.food = Food{
		frameCount:  maxFrameCount,
		position:    Position{posX: 24, posY: 13},
		direction:   DIRECTION_NIL,
		immuneUntil: 0,
		isGolden:    false}
	g.goldenFood = GoldenFood{
		position: Position{posX: g.rand.Intn(20) + 3, posY: g.rand.Intn(9) + 3},
		visible:  true,
		duration: 5000}
	g.item = Item{
		used:     false,
		duration: 3000}
	return nil
}

func (g *Game) Finish(title string) {
	g.title = title
	g.started = false
	g.endTimestamp = time.Now().UnixMilli()
}

func (g *Game) Update() error {
	if g.started {
		if inpututil.IsKeyJustPressed(ebiten.KeyA) && (g.snake.lastDirection != DIRECTION_RIGHT || g.snake.hitWall) {
			g.snake.direction = DIRECTION_LEFT
		} else if inpututil.IsKeyJustPressed(ebiten.KeyD) && (g.snake.lastDirection != DIRECTION_LEFT || g.snake.hitWall) {
			g.snake.direction = DIRECTION_RIGHT
		} else if inpututil.IsKeyJustPressed(ebiten.KeyW) && (g.snake.lastDirection != DIRECTION_DOWN || g.snake.hitWall) {
			g.snake.direction = DIRECTION_UP
		} else if inpututil.IsKeyJustPressed(ebiten.KeyS) && (g.snake.lastDirection != DIRECTION_UP || g.snake.hitWall) {
			g.snake.direction = DIRECTION_DOWN
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.food.direction = DIRECTION_LEFT
		} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.food.direction = DIRECTION_RIGHT
		} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.food.direction = DIRECTION_UP
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.food.direction = DIRECTION_DOWN
		}

		if g.food.direction == DIRECTION_LEFT && inpututil.IsKeyJustReleased(ebiten.KeyLeft) ||
			g.food.direction == DIRECTION_RIGHT && inpututil.IsKeyJustReleased(ebiten.KeyRight) ||
			g.food.direction == DIRECTION_UP && inpututil.IsKeyJustReleased(ebiten.KeyUp) ||
			g.food.direction == DIRECTION_DOWN && inpututil.IsKeyJustReleased(ebiten.KeyDown) {
			g.food.direction = DIRECTION_NIL

			pressedKeys := inpututil.AppendPressedKeys(nil)
			for _, key := range pressedKeys {
				if key == ebiten.KeyArrowLeft {
					g.food.direction = DIRECTION_LEFT
				} else if key == ebiten.KeyArrowRight {
					g.food.direction = DIRECTION_RIGHT
				} else if key == ebiten.KeyArrowUp {
					g.food.direction = DIRECTION_UP
				} else if key == ebiten.KeyArrowDown {
					g.food.direction = DIRECTION_DOWN
				}
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySlash) && g.item.used == false {
			g.food.immuneUntil = time.Now().UnixMilli() + g.item.duration
			g.item.used = true
		}
	} else {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.Start()
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.started {
		g.snake.frameCount++
		if g.snake.frameCount == maxFrameCount {
			g.snake.frameCount = 0
			moveSnake(g)
			if g.food.immuneUntil < time.Now().UnixMilli() {
				checkCollision(g)
			}
			if g.food.isGolden {
				if g.food.immuneUntil < time.Now().UnixMilli() {
					g.food.isGolden = false
				} else {
					checkGoldenCollision(g)
				}
			}
		}

		g.food.frameCount++
		if g.food.frameCount >= maxFrameCount {
			if g.food.direction != DIRECTION_NIL {
				g.food.frameCount = 0
				moveFood(g)
				if g.food.immuneUntil < time.Now().UnixMilli() {
					checkCollision(g)
				}
				if g.food.isGolden {
					if g.food.immuneUntil < time.Now().UnixMilli() {
						g.food.isGolden = false
					} else {
						checkGoldenCollision(g)
					}
				}
			}
		}

		timeString := fmt.Sprintf("%.1f 초", float64(time.Now().UnixMilli()-g.timestamp)/1000)

		bounds, _ := font.BoundString(textFont, timeString)

		text.Draw(screen, timeString, textFont, (screenWidth-bounds.Max.X.Floor())/2, screenHeight-50, color.White)
		text.Draw(screen, fmt.Sprintf("%d / %d초간 굶음", (g.snake.starveDelay-(g.snake.starveTimestamp-time.Now().UnixMilli()))/1000, g.snake.starveDelay/1000), textFont, 40, screenHeight-50, color.White)
	} else {
		timeString := fmt.Sprintf("%.1f 초", float64(g.endTimestamp-g.timestamp)/1000)

		bounds, _ := font.BoundString(textFont, timeString)

		if len(g.title) == 0 {
			text.Draw(screen, "SPACE를 눌러 게임을 시작하세요", textFont, 40, screenHeight-50, color.White)
		} else {
			text.Draw(screen, g.title, textFont, 40, screenHeight-50, color.White)
			text.Draw(screen, timeString, textFont, (screenWidth-bounds.Max.X.Floor())/2, screenHeight-50, color.White)
		}
	}

	for x := 1; x < screenWidth/blockWidth-1; x++ {
		for y := 1; y < screenHeight/blockHeight-3; y++ {
			if g.food.position.posX == x && g.food.position.posY == y {
				c := color.RGBA{
					R: 0x20,
					G: 0xd0,
					B: 0x20,
					A: 0xf0,
				}

				if g.food.isGolden {
					c = color.RGBA{
						R: 0xe0,
						G: 0xe0,
						B: 0x20,
						A: 0xff,
					}
				} else if g.food.immuneUntil >= time.Now().UnixMilli() {
					c = color.RGBA{
						R: 0x20,
						G: 0x80,
						B: 0x20,
						A: 0xff,
					}
				}

				vector.DrawFilledRect(
					screen,
					float32(x*blockWidth), float32(y*blockHeight),
					blockWidth, blockHeight,
					c, true)
			}
			if g.goldenFood.visible && g.goldenFood.position.posX == x && g.goldenFood.position.posY == y {
				c := color.RGBA{
					R: 0xe0,
					G: 0xe0,
					B: 0x20,
					A: 0xff,
				}

				vector.DrawFilledRect(
					screen,
					float32(x*blockWidth), float32(y*blockHeight),
					blockWidth, blockHeight,
					c, true)
			}
			if slices.Contains(g.snake.positions, Position{posX: x, posY: y}) {
				snakeIndex := slices.Index(g.snake.positions, Position{posX: x, posY: y})
				c := color.RGBA{
					R: 0xf0 - uint8(0x0f*snakeIndex),
					G: 0xf0 - uint8(0x0f*snakeIndex),
					B: 0xf0 - uint8(0x0f*snakeIndex),
					A: 0x20,
				}

				vector.DrawFilledRect(
					screen,
					float32(x*blockWidth), float32(y*blockHeight),
					blockWidth, blockHeight,
					c, true)
			} else {
				vector.StrokeRect(
					screen,
					float32(x*blockWidth), float32(y*blockHeight),
					blockWidth, blockHeight,
					1, color.White, true)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func moveSnake(g *Game) {
	if len(g.snake.positions) == 0 {
		g.Finish("먹이에게 잡아먹힘")
		return
	}

	lastPosition := g.snake.positions[0]
	newPosition := Position{posX: lastPosition.posX, posY: lastPosition.posY}

	g.snake.lastDirection = g.snake.direction

	g.snake.hitWall = false
	if g.snake.direction == DIRECTION_LEFT {
		if newPosition.posX > 1 {
			newPosition.posX--
		} else {
			g.snake.hitWall = true
		}
	} else if g.snake.direction == DIRECTION_RIGHT {
		if newPosition.posX < screenWidth/blockWidth-2 {
			newPosition.posX++
		} else {
			g.snake.hitWall = true
		}
	} else if g.snake.direction == DIRECTION_UP {
		if newPosition.posY > 1 {
			newPosition.posY--
		} else {
			g.snake.hitWall = true
		}
	} else if g.snake.direction == DIRECTION_DOWN {
		if newPosition.posY < screenHeight/blockHeight-4 {
			newPosition.posY++
		} else {
			g.snake.hitWall = true
		}
	}

	if g.snake.starveTimestamp < time.Now().UnixMilli() {
		if len(g.snake.positions) == 1 {
			g.Finish("먹이를 먹지 못해 아사함")
			return
		}
		g.snake.starveDelay += starveDelayIncrement
		g.snake.starveTimestamp = time.Now().UnixMilli() + g.snake.starveDelay
		g.snake.positions = append([]Position{newPosition}, g.snake.positions[:len(g.snake.positions)-2]...)
	} else {
		g.snake.positions = append([]Position{newPosition}, g.snake.positions[:len(g.snake.positions)-1]...)
	}
}

func moveFood(g *Game) {
	if g.food.direction == DIRECTION_LEFT && g.food.position.posX > 1 {
		g.food.position.posX--
	} else if g.food.direction == DIRECTION_RIGHT && g.food.position.posX < screenWidth/blockWidth-2 {
		g.food.position.posX++
	} else if g.food.direction == DIRECTION_UP && g.food.position.posY > 1 {
		g.food.position.posY--
	} else if g.food.direction == DIRECTION_DOWN && g.food.position.posY < screenHeight/blockHeight-4 {
		g.food.position.posY++
	}

	if g.goldenFood.visible && g.food.position.posX == g.goldenFood.position.posX && g.food.position.posY == g.goldenFood.position.posY {
		g.goldenFood.visible = false
		g.food.immuneUntil = time.Now().UnixMilli() + g.goldenFood.duration
		g.food.isGolden = true
	}
}

func checkCollision(g *Game) {
	for _, p := range g.snake.positions {
		if p.posX == g.food.position.posX &&
			p.posY == g.food.position.posY {
			g.Finish("뱀에게 잡아먹힘")
			return
		}
	}
}

func checkGoldenCollision(g *Game) {
	for i, p := range g.snake.positions {
		if p.posX == g.food.position.posX && p.posY == g.food.position.posY {
			if len(g.snake.positions) == 0 {
				g.Finish("먹이에게 잡아먹힘")
				return
			} else if i <= len(g.snake.positions)/2 {
				g.snake.positions = g.snake.positions[:i]
			} else {
				g.snake.positions = g.snake.positions[i:]
				slices.Reverse(g.snake.positions)
			}
		}
	}
}

func main() {
	fontPath := "AppleSDGothicNeoH.ttf"

	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		log.Fatalf("Failed to read font file: %v", err)
	}

	font, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatalf("Failed to parse font: %v", err)
	}

	textFont, err = opentype.NewFace(font, &opentype.FaceOptions{
		Size: 40,
		DPI:  72,
	})

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Reversnake (Snake Reversed)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
