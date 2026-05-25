package main

import (
	"math/rand"
	"syscall/js"
	"time"
)

// Board dimensions
const (
	boardWidth  = 10
	boardHeight = 20
	blockSize   = 30
	canvasWidth  = boardWidth * blockSize  // 300
	canvasHeight = boardHeight * blockSize // 600
	sidePanelWidth = 120
	totalWidth    = canvasWidth + sidePanelWidth
)

// Tetromino definitions (7 standard pieces)
var tetrominoes = [][][]int{
	// I
	{{1, 1, 1, 1}},
	// O
	{{1, 1},
		{1, 1}},
	// T
	{{0, 1, 0},
		{1, 1, 1}},
	// S
	{{0, 1, 1},
		{1, 1, 0}},
	// Z
	{{1, 1, 0},
		{0, 1, 1}},
	// J
	{{1, 0, 0},
		{1, 1, 1}},
	// L
	{{0, 0, 1},
		{1, 1, 1}},
}

var pieceColors = []string{
	"#00f0f0", // I - cyan
	"#f0f000", // O - yellow
	"#a000f0", // T - purple
	"#00f000", // S - green
	"#f00000", // Z - red
	"#0000f0", // J - blue
	"#f0a000", // L - orange
}

type Piece struct {
	shape   [][]int
	color   string
	x, y    int
	typeIdx int
}

type Game struct {
	board      [boardHeight][boardWidth]int
	colors     [boardHeight][boardWidth]string
	current    *Piece
	next       *Piece
	score      int
	level      int
	lines      int
	gameOver   bool
	paused     bool
	dropTimer  int
	dropInterval int
	canvas     js.Value
	ctx        js.Value
	lastTime   int64
	touchStartX int
	touchStartY int
	touchStartTime int64
}

func NewGame(canvas js.Value) *Game {
	ctx := canvas.Call("getContext", "2d")
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		canvas:        canvas,
		ctx:           ctx,
		dropInterval:  500,
		lastTime:      time.Now().UnixMilli(),
		touchStartX:   -1,
		touchStartY:   -1,
	}
	g.next = g.randomPiece()
	g.spawnPiece()
	return g
}

func (g *Game) randomPiece() *Piece {
	idx := rand.Intn(len(tetrominoes))
	shape := tetrominoes[idx]
	color := pieceColors[idx]
	piece := &Piece{
		shape:   shape,
		color:   color,
		x:       boardWidth/2 - len(shape[0])/2,
		y:       0,
		typeIdx: idx,
	}
	return piece
}

func (g *Game) spawnPiece() {
	g.current = g.next
	g.next = g.randomPiece()
	g.current.x = boardWidth/2 - len(g.current.shape[0])/2
	g.current.y = 0

	if g.collides(g.current.shape, g.current.x, g.current.y) {
		g.gameOver = true
	}
}

func (g *Game) collides(shape [][]int, x, y int) bool {
	for row := 0; row < len(shape); row++ {
		for col := 0; col < len(shape[row]); col++ {
			if shape[row][col] != 0 {
				boardX := x + col
				boardY := y + row
				if boardX < 0 || boardX >= boardWidth || boardY >= boardHeight {
					return true
				}
				if boardY >= 0 && g.board[boardY][boardX] != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (g *Game) lockPiece() {
	for row := 0; row < len(g.current.shape); row++ {
		for col := 0; col < len(g.current.shape[row]); col++ {
			if g.current.shape[row][col] != 0 {
				boardY := g.current.y + row
				boardX := g.current.x + col
				if boardY >= 0 && boardY < boardHeight && boardX >= 0 && boardX < boardWidth {
					g.board[boardY][boardX] = 1
					g.colors[boardY][boardX] = g.current.color
				}
			}
		}
	}
	g.clearLines()
	g.spawnPiece()
}

func (g *Game) clearLines() {
	cleared := 0
	for row := boardHeight - 1; row >= 0; row-- {
		full := true
		for col := 0; col < boardWidth; col++ {
			if g.board[row][col] == 0 {
				full = false
				break
			}
		}
		if full {
			// Shift everything down
			for r := row; r > 0; r-- {
				g.board[r] = g.board[r-1]
				g.colors[r] = g.colors[r-1]
			}
			g.board[0] = [boardWidth]int{}
			g.colors[0] = [boardWidth]string{}
			cleared++
			row++ // Re-check this row
		}
	}

	if cleared > 0 {
		points := []int{0, 100, 300, 500, 800}
		if cleared <= 4 {
			g.score += points[cleared]
		} else {
			g.score += points[4]
		}
		g.lines += cleared
		g.level = g.lines / 10
		newInterval := 500 - g.level*50
		if newInterval < 100 {
			newInterval = 100
		}
		g.dropInterval = newInterval
	}
}

func (g *Game) moveLeft() {
	if g.current != nil && !g.gameOver {
		if !g.collides(g.current.shape, g.current.x-1, g.current.y) {
			g.current.x--
		}
	}
}

func (g *Game) moveRight() {
	if g.current != nil && !g.gameOver {
		if !g.collides(g.current.shape, g.current.x+1, g.current.y) {
			g.current.x++
		}
	}
}

func (g *Game) rotate() {
	if g.current == nil || g.gameOver {
		return
	}
	shape := g.current.shape
	rows := len(shape)
	cols := len(shape[0])
	rotated := make([][]int, cols)
	for i := range rotated {
		rotated[i] = make([]int, rows)
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			rotated[c][rows-1-r] = shape[r][c]
		}
	}
	if !g.collides(rotated, g.current.x, g.current.y) {
		g.current.shape = rotated
	}
}

func (g *Game) drop() {
	if g.current == nil || g.gameOver {
		return
	}
	for !g.collides(g.current.shape, g.current.x, g.current.y+1) {
		g.current.y++
	}
	g.lockPiece()
}

func (g *Game) softDrop() {
	if g.current != nil && !g.gameOver {
		if !g.collides(g.current.shape, g.current.x, g.current.y+1) {
			g.current.y++
			g.score++
		}
	}
}

func (g *Game) update() {
	if g.gameOver || g.paused || g.current == nil {
		return
	}
	now := time.Now().UnixMilli()
	if now-g.lastTime >= int64(g.dropInterval) {
		if !g.collides(g.current.shape, g.current.x, g.current.y+1) {
			g.current.y++
		} else {
			g.lockPiece()
		}
		g.lastTime = now
	}
}

func (g *Game) draw() {
	g.ctx.Call("clearRect", 0, 0, totalWidth, canvasHeight)

	// Draw background
	g.ctx.Set("fillStyle", "#1a1a2e")
	g.ctx.Call("fillRect", 0, 0, totalWidth, canvasHeight)

	// Draw grid
	g.ctx.Set("strokeStyle", "#2a2a3e")
	g.ctx.Set("lineWidth", 0.5)
	for y := 0; y <= boardHeight; y++ {
		g.ctx.Call("beginPath")
		g.ctx.Call("moveTo", 0, y*blockSize)
		g.ctx.Call("lineTo", canvasWidth, y*blockSize)
		g.ctx.Call("stroke")
	}
	for x := 0; x <= boardWidth; x++ {
		g.ctx.Call("beginPath")
		g.ctx.Call("moveTo", x*blockSize, 0)
		g.ctx.Call("lineTo", x*blockSize, canvasHeight)
		g.ctx.Call("stroke")
	}

	// Draw locked pieces
	for y := 0; y < boardHeight; y++ {
		for x := 0; x < boardWidth; x++ {
			if g.board[y][x] != 0 {
				g.drawBlock(x, y, g.colors[y][x])
			}
		}
	}

	// Draw ghost piece
	if g.current != nil {
		ghostY := g.current.y
		for !g.collides(g.current.shape, g.current.x, ghostY+1) {
			ghostY++
		}
		for row := 0; row < len(g.current.shape); row++ {
			for col := 0; col < len(g.current.shape[row]); col++ {
				if g.current.shape[row][col] != 0 {
					g.drawGhostBlock(g.current.x+col, ghostY+row, g.current.color)
				}
			}
		}
	}

	// Draw current piece
	if g.current != nil {
		for row := 0; row < len(g.current.shape); row++ {
			for col := 0; col < len(g.current.shape[row]); col++ {
				if g.current.shape[row][col] != 0 {
					g.drawBlock(g.current.x+col, g.current.y+row, g.current.color)
				}
			}
		}
	}

	// Draw side panel
	g.drawSidePanel()

	// Game over overlay
	if g.gameOver {
		g.ctx.Set("fillStyle", "rgba(0,0,0,0.7)")
		g.ctx.Call("fillRect", 0, 0, totalWidth, canvasHeight)
		g.ctx.Set("fillStyle", "#ffffff")
		g.ctx.Set("font", "bold 36px Arial")
		g.ctx.Set("textAlign", "center")
		g.ctx.Call("fillText", "GAME OVER", totalWidth/2, canvasHeight/2-20)
		g.ctx.Set("font", "18px Arial")
		g.ctx.Call("fillText", "Tap to restart", totalWidth/2, canvasHeight/2+20)
	}
}

func (g *Game) drawBlock(x, y int, color string) {
	px := x * blockSize
	py := y * blockSize

	g.ctx.Set("fillStyle", color)
	g.ctx.Call("fillRect", px+1, py+1, blockSize-2, blockSize-2)

	// Highlight
	g.ctx.Set("fillStyle", "rgba(255,255,255,0.2)")
	g.ctx.Call("fillRect", px+1, py+1, blockSize-2, 4)
	g.ctx.Call("fillRect", px+1, py+1, 4, blockSize-2)

	// Shadow
	g.ctx.Set("fillStyle", "rgba(0,0,0,0.2)")
	g.ctx.Call("fillRect", px+blockSize-5, py+1, 4, blockSize-2)
	g.ctx.Call("fillRect", px+1, py+blockSize-5, blockSize-2, 4)
}

func (g *Game) drawGhostBlock(x, y int, color string) {
	px := x * blockSize
	py := y * blockSize

	g.ctx.Set("strokeStyle", color)
	g.ctx.Set("lineWidth", 2)
	g.ctx.Set("globalAlpha", 0.3)
	g.ctx.Call("strokeRect", px+2, py+2, blockSize-4, blockSize-4)
	g.ctx.Set("globalAlpha", 1.0)
}

func (g *Game) drawSidePanel() {
	panelX := canvasWidth + 10

	g.ctx.Set("fillStyle", "#ffffff")
	g.ctx.Set("font", "bold 16px Arial")
	g.ctx.Set("textAlign", "left")
	g.ctx.Call("fillText", "NEXT", panelX, 30)

	// Draw next piece preview
	if g.next != nil {
		previewBlockSize := 20
		offsetX := panelX + 10
		offsetY := 50
		for row := 0; row < len(g.next.shape); row++ {
			for col := 0; col < len(g.next.shape[row]); col++ {
				if g.next.shape[row][col] != 0 {
					g.ctx.Set("fillStyle", g.next.color)
					g.ctx.Call("fillRect", offsetX+col*previewBlockSize+1, offsetY+row*previewBlockSize+1, previewBlockSize-2, previewBlockSize-2)
				}
			}
		}
	}

	g.ctx.Set("fillStyle", "#ffffff")
	g.ctx.Set("font", "bold 16px Arial")
	g.ctx.Call("fillText", "SCORE", panelX, 140)
	g.ctx.Set("font", "24px Arial")
	g.ctx.Call("fillText", formatScore(g.score), panelX, 170)

	g.ctx.Set("font", "bold 16px Arial")
	g.ctx.Call("fillText", "LEVEL", panelX, 210)
	g.ctx.Set("font", "24px Arial")
	g.ctx.Call("fillText", formatScore(g.level), panelX, 240)

	g.ctx.Set("font", "bold 16px Arial")
	g.ctx.Call("fillText", "LINES", panelX, 280)
	g.ctx.Set("font", "24px Arial")
	g.ctx.Call("fillText", formatScore(g.lines), panelX, 310)

	// Controls hint
	g.ctx.Set("fillStyle", "#888888")
	g.ctx.Set("font", "11px Arial")
	g.ctx.Call("fillText", "← → Move", panelX, 370)
	g.ctx.Call("fillText", "↑ Rotate", panelX, 390)
	g.ctx.Call("fillText", "↓ Soft Drop", panelX, 410)
	g.ctx.Call("fillText", "Space Drop", panelX, 430)
}

func formatScore(s int) string {
	return js.Global().Get("String").New(s).String()
}

func (g *Game) handleKey(key string) {
	if g.gameOver {
		if key == " " || key == "Enter" || key == "r" {
			g.reset()
		}
		return
	}

	switch key {
	case "ArrowLeft":
		g.moveLeft()
	case "ArrowRight":
		g.moveRight()
	case "ArrowUp":
		g.rotate()
	case "ArrowDown":
		g.softDrop()
	case " ":
		g.drop()
	case "p":
		g.paused = !g.paused
	}
}

func (g *Game) reset() {
	g.board = [boardHeight][boardWidth]int{}
	g.colors = [boardHeight][boardWidth]string{}
	g.score = 0
	g.level = 0
	g.lines = 0
	g.gameOver = false
	g.paused = false
	g.dropInterval = 500
	g.next = g.randomPiece()
	g.spawnPiece()
	g.lastTime = time.Now().UnixMilli()
}

func main() {
	canvas := js.Global().Get("document").Call("getElementById", "gameCanvas")
	if canvas.IsUndefined() || canvas.IsNull() {
		// Try to create canvas if not found
		body := js.Global().Get("document").Get("body")
		canvas = js.Global().Get("document").Call("createElement", "canvas")
		canvas.Set("id", "gameCanvas")
		canvas.Set("width", totalWidth)
		canvas.Set("height", canvasHeight)
		body.Call("appendChild", canvas)
	}
	canvas.Set("width", totalWidth)
	canvas.Set("height", canvasHeight)

	game := NewGame(canvas)

	// Keyboard handler
	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("key").String()
		game.handleKey(key)
		event.Call("preventDefault")
		return nil
	}))

	// Touch handler variables
	var touchStartX, touchStartY float64
	touchActive := false

	// Touch start - records initial touch position
	canvas.Call("addEventListener", "touchstart", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		touch := event.Get("touches").Index(0)
		touchStartX = touch.Get("clientX").Float()
		touchStartY = touch.Get("clientY").Float()
		touchActive = true

		if game.gameOver {
			game.reset()
			return nil
		}

		return nil
	}))

	// Touch move - swipe left/right to move, swipe down for soft drop
	canvas.Call("addEventListener", "touchmove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		if !touchActive || game.gameOver {
			return nil
		}
		touch := event.Get("touches").Index(0)
		dx := touch.Get("clientX").Float() - touchStartX
		dy := touch.Get("clientY").Float() - touchStartY

		if dx > 20 {
			game.moveRight()
			touchStartX = touch.Get("clientX").Float()
		} else if dx < -20 {
			game.moveLeft()
			touchStartX = touch.Get("clientX").Float()
		}

		if dy > 40 {
			game.softDrop()
			touchStartY = touch.Get("clientY").Float()
		}
		return nil
	}))

	// Touch end - unified handler: swipe up=rotate, tap=rotate, double tap=hard drop
	var tapCount int
	var lastTapTime int64

	canvas.Call("addEventListener", "touchend", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]

		if game.gameOver {
			return nil
		}

		if !touchActive {
			return nil
		}
		touchActive = false

		// Get the final touch position
		touch := event.Get("changedTouches").Index(0)
		endX := touch.Get("clientX").Float()
		endY := touch.Get("clientY").Float()

		dx := endX - touchStartX
		dy := endY - touchStartY

		if dy < -30 {
			// Swipe up = rotate
			game.rotate()
		} else if dx*dx+dy*dy < 400 {
			// Tap or small movement - double tap for hard drop, single tap for rotate
			now := time.Now().UnixMilli()
			if now-lastTapTime < 300 {
				tapCount++
			} else {
				tapCount = 1
			}
			lastTapTime = now
			if tapCount >= 2 {
				game.drop()
				tapCount = 0
			} else {
				game.rotate()
			}
		}
		return nil
	}))

	// Click fallback (for desktop + game over restart)
	canvas.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if game.gameOver {
			game.reset()
			return nil
		}
		return nil
	}))

	// Game loop
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		game.update()
		game.draw()
		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", renderFrame)

	// Keep program running
	select {}
}
