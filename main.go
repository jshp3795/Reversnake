package main

import (
	"image/color"
	_ "image/png"
	"log"

	"slices"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
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
	maxFrameCount = 6

	maxLength = 10

	DIRECTION_NIL   = -1
	DIRECTION_LEFT  = 0
	DIRECTION_RIGHT = 1
	DIRECTION_UP    = 2
	DIRECTION_DOWN  = 3
)

var (
	gameStarted = true

	snakePos           = []Position{{posX: 10, posY: 1}, {posX: 9, posY: 1}, {posX: 8, posY: 1}, {posX: 7, posY: 1}, {posX: 6, posY: 1}, {posX: 5, posY: 1}, {posX: 4, posY: 1}, {posX: 3, posY: 1}, {posX: 2, posY: 1}, {posX: 1, posY: 1}}
	snakeDirection     = DIRECTION_RIGHT
	lastSnakeDirection = DIRECTION_RIGHT
	snakeHitWall       = false

	foodPos       = Position{posX: 16, posY: 16}
	foodDirection = DIRECTION_NIL

	frameCount     = 0
	foodFrameCount = maxFrameCount

	textFont font.Face = nil
)

type Game struct {
	count int
}

func (g *Game) Update() error {
	g.count++

	if inpututil.IsKeyJustPressed(ebiten.KeyA) && (lastSnakeDirection != DIRECTION_RIGHT || snakeHitWall) {
		snakeDirection = DIRECTION_LEFT
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) && (lastSnakeDirection != DIRECTION_LEFT || snakeHitWall) {
		snakeDirection = DIRECTION_RIGHT
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) && (lastSnakeDirection != DIRECTION_DOWN || snakeHitWall) {
		snakeDirection = DIRECTION_UP
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) && (lastSnakeDirection != DIRECTION_UP || snakeHitWall) {
		snakeDirection = DIRECTION_DOWN
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		foodDirection = DIRECTION_LEFT
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		foodDirection = DIRECTION_RIGHT
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		foodDirection = DIRECTION_UP
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		foodDirection = DIRECTION_DOWN
	}

	if foodDirection == DIRECTION_LEFT && inpututil.IsKeyJustReleased(ebiten.KeyLeft) ||
		foodDirection == DIRECTION_RIGHT && inpututil.IsKeyJustReleased(ebiten.KeyRight) ||
		foodDirection == DIRECTION_UP && inpututil.IsKeyJustReleased(ebiten.KeyUp) ||
		foodDirection == DIRECTION_DOWN && inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		foodDirection = DIRECTION_NIL

		pressedKeys := inpututil.AppendPressedKeys(nil)
		for _, key := range pressedKeys {
			if key == ebiten.KeyArrowLeft {
				foodDirection = DIRECTION_LEFT
			} else if key == ebiten.KeyArrowRight {
				foodDirection = DIRECTION_RIGHT
			} else if key == ebiten.KeyArrowUp {
				foodDirection = DIRECTION_UP
			} else if key == ebiten.KeyArrowDown {
				foodDirection = DIRECTION_DOWN
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if gameStarted {
		frameCount++
		if frameCount == maxFrameCount {
			frameCount = 0
			moveSnake()
		}

		foodFrameCount++
		if foodFrameCount >= maxFrameCount {
			if foodDirection != DIRECTION_NIL {
				foodFrameCount = 0
				moveFood()
			}
		}
	} else {
		text.Draw(screen, "GAME ENDED", textFont, 200, 40, color.White)
	}

	for x := 1; x < screenWidth/blockWidth-1; x++ {
		for y := 1; y < screenHeight/blockHeight-1; y++ {
			if slices.Contains(snakePos, Position{posX: x, posY: y}) {
				snakeIndex := slices.Index(snakePos, Position{posX: x, posY: y})
				c := color.RGBA{
					R: 0xff - uint8(0x0f*snakeIndex),
					G: 0xff - uint8(0x0f*snakeIndex),
					B: 0xff - uint8(0x0f*snakeIndex),
					A: 0xff,
				}

				vector.DrawFilledRect(
					screen,
					float32(x*blockWidth), float32(y*blockHeight),
					blockWidth, blockHeight,
					c, true)
			} else if foodPos.posX == x && foodPos.posY == y {
				c := color.RGBA{
					R: 0x20,
					G: 0xc0,
					B: 0x20,
					A: 0xff,
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
	//vector.StrokeRect(screen, 0, 0, blockWidth, blockHeight, 1, c, true)
	//screen.DrawImage(runnerImage.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image), op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func moveSnake() {
	lastPosition := snakePos[0]
	newPosition := Position{posX: lastPosition.posX, posY: lastPosition.posY}

	lastSnakeDirection = snakeDirection

	snakeHitWall = false
	if snakeDirection == DIRECTION_LEFT {
		if newPosition.posX <= 1 {
			snakeHitWall = true
		} else {
			newPosition.posX--
		}
	} else if snakeDirection == DIRECTION_RIGHT {
		if newPosition.posX >= screenWidth/blockWidth-2 {
			snakeHitWall = true
		} else {
			newPosition.posX++
		}
	} else if snakeDirection == DIRECTION_UP {
		if newPosition.posY <= 1 {
			snakeHitWall = true
		} else {
			newPosition.posY--
		}
	} else if snakeDirection == DIRECTION_DOWN {
		if newPosition.posY >= screenHeight/blockHeight-2 {
			snakeHitWall = true
		} else {
			newPosition.posY++
		}
	}

	snakePos = append([]Position{newPosition}, snakePos[:len(snakePos)-1]...)
}

func moveFood() {
	newPosition := Position{posX: foodPos.posX, posY: foodPos.posY}

	if foodDirection == DIRECTION_LEFT {
		newPosition.posX--
		if newPosition.posX < 0 {
			newPosition.posX = screenWidth/blockWidth - 1
		}
	} else if foodDirection == DIRECTION_RIGHT {
		newPosition.posX++
		if newPosition.posX >= screenWidth/blockWidth {
			newPosition.posX = 0
		}
	} else if foodDirection == DIRECTION_UP {
		newPosition.posY--
		if newPosition.posY < 0 {
			newPosition.posY = screenHeight/blockHeight - 1
		}
	} else if foodDirection == DIRECTION_DOWN {
		newPosition.posY++
		if newPosition.posY >= screenHeight/blockHeight {
			newPosition.posY = 0
		}
	}

	foodPos = newPosition
}

func main() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	textFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Reversnake (Snake Reversed)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
