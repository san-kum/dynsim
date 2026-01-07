package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/san-kum/dynsim/internal/sim"
)

const (
	width       = 70
	height      = 20
	clearScreen = "\033[2J\033[H"
	hideCursor  = "\033[?25l"
	showCursor  = "\033[?25h"
)

type LiveRenderer struct {
	model     string
	frameRate int
	lastFrame time.Time
	canvas    [][]rune
	trail     []struct{ x, y int }
}

func NewLiveRenderer(model string, frameRate int) *LiveRenderer {
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
	}
	return &LiveRenderer{
		model:     model,
		frameRate: frameRate,
		canvas:    canvas,
		trail:     make([]struct{ x, y int }, 0, 50),
	}
}

func (r *LiveRenderer) OnStep(x sim.State, u sim.Control, t float64) {
	elapsed := time.Since(r.lastFrame)
	if elapsed < time.Second/time.Duration(r.frameRate) {
		return
	}
	r.lastFrame = time.Now()

	r.clear()

	switch r.model {
	case "pendulum":
		r.drawPendulum(x)
	case "double_pendulum":
		r.drawDoublePendulum(x)
	case "cartpole":
		r.drawCartpole(x)
	case "spring_mass", "spring_chain":
		r.drawSpring(x)
	case "drone":
		r.drawDrone(x)
	default:
		r.drawGeneric(x)
	}

	r.render(x, t)
}

func (r *LiveRenderer) clear() {
	for y := range r.canvas {
		for x := range r.canvas[y] {
			r.canvas[y][x] = ' '
		}
	}
}

func (r *LiveRenderer) set(x, y int, c rune) {
	if x >= 0 && x < width && y >= 0 && y < height {
		r.canvas[y][x] = c
	}
}

func (r *LiveRenderer) line(x1, y1, x2, y2 int, c rune) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		r.set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (r *LiveRenderer) drawPendulum(x sim.State) {
	if len(x) < 2 {
		return
	}
	theta := x[0]
	px, py := width/2, 3
	length := 10.0
	bx := px + int(length*math.Sin(theta))
	by := py + int(length*math.Cos(theta))

	r.trail = append(r.trail, struct{ x, y int }{bx, by})
	if len(r.trail) > 40 {
		r.trail = r.trail[1:]
	}

	for i, pt := range r.trail {
		if i < len(r.trail)/2 {
			r.set(pt.x, pt.y, '.')
		} else {
			r.set(pt.x, pt.y, 'o')
		}
	}

	r.set(px, py, '+')
	r.line(px, py, bx, by, '|')
	r.set(bx, by, 'O')
}

func (r *LiveRenderer) drawDoublePendulum(x sim.State) {
	if len(x) < 4 {
		return
	}
	t1, t2 := x[0], x[1]
	px, py := width/2, 2
	length := 6.0

	b1x := px + int(length*math.Sin(t1))
	b1y := py + int(length*math.Cos(t1))
	b2x := b1x + int(length*math.Sin(t2))
	b2y := b1y + int(length*math.Cos(t2))

	r.trail = append(r.trail, struct{ x, y int }{b2x, b2y})
	if len(r.trail) > 50 {
		r.trail = r.trail[1:]
	}

	for _, pt := range r.trail {
		r.set(pt.x, pt.y, '.')
	}

	r.set(px, py, '+')
	r.line(px, py, b1x, b1y, '|')
	r.set(b1x, b1y, 'o')
	r.line(b1x, b1y, b2x, b2y, '|')
	r.set(b2x, b2y, 'O')
}

func (r *LiveRenderer) drawCartpole(x sim.State) {
	if len(x) < 4 {
		return
	}
	pos, theta := x[0], x[2]
	gy := height - 4
	cx := width/2 + int(pos*8)

	for i := 5; i < width-5; i++ {
		r.set(i, gy+1, '=')
	}
	for dx := -3; dx <= 3; dx++ {
		r.set(cx+dx, gy, '#')
	}

	plen := 8.0
	px := cx + int(plen*math.Sin(theta))
	py := gy - int(plen*math.Cos(theta))
	r.line(cx, gy-1, px, py, '|')
	r.set(px, py, 'o')
}

func (r *LiveRenderer) drawSpring(x sim.State) {
	if len(x) < 2 {
		return
	}
	pos := x[0]
	cy := height / 2

	for y := cy - 2; y <= cy+2; y++ {
		r.set(5, y, '#')
	}

	mx := 20 + int(pos*8)
	for i := 6; i < mx-2; i += 2 {
		r.set(i, cy, '~')
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			r.set(mx+dx, cy+dy, '#')
		}
	}
}

func (r *LiveRenderer) drawDrone(x sim.State) {
	if len(x) < 6 {
		return
	}
	px, py, theta := x[0], x[1], x[2]

	for i := 3; i < width-3; i++ {
		r.set(i, height-2, '_')
	}

	dx := width/2 + int(px*3)
	dy := height/2 - int(py*1.2)

	r.trail = append(r.trail, struct{ x, y int }{dx, dy})
	if len(r.trail) > 30 {
		r.trail = r.trail[1:]
	}
	for _, pt := range r.trail {
		r.set(pt.x, pt.y, '.')
	}

	arm := 4.0
	lx := dx - int(arm*math.Cos(theta))
	ly := dy - int(arm*math.Sin(theta))
	rx := dx + int(arm*math.Cos(theta))
	ry := dy + int(arm*math.Sin(theta))

	r.line(lx, ly, rx, ry, '-')
	r.set(dx, dy, 'X')
	r.set(lx, ly, 'o')
	r.set(rx, ry, 'o')
}

func (r *LiveRenderer) drawGeneric(x sim.State) {
	cy := height / 2
	for i := 5; i < width-5; i++ {
		r.set(i, cy, '-')
	}

	if len(x) == 0 {
		return
	}

	bw := (width - 15) / len(x)
	if bw < 3 {
		bw = 3
	}

	maxVal := 1.0
	for _, v := range x {
		if math.Abs(v) > maxVal {
			maxVal = math.Abs(v)
		}
	}

	for i, v := range x {
		bx := 8 + i*bw
		bh := int((v / maxVal) * float64(height/3))
		if bh > 0 {
			for y := cy - 1; y >= cy-bh && y >= 1; y-- {
				r.set(bx, y, '#')
			}
		} else {
			for y := cy + 1; y <= cy-bh && y < height-1; y++ {
				r.set(bx, y, '#')
			}
		}
	}
}

func (r *LiveRenderer) render(x sim.State, t float64) {
	var b strings.Builder
	b.WriteString(clearScreen)
	b.WriteString(fmt.Sprintf("  %s  t=%.2fs\n", r.model, t))
	b.WriteString("  " + strings.Repeat("-", width) + "\n")

	for _, row := range r.canvas {
		b.WriteString("  ")
		b.WriteString(string(row))
		b.WriteString("\n")
	}

	b.WriteString("  " + strings.Repeat("-", width) + "\n")

	stateStr := "  "
	for i, v := range x {
		if i >= 4 {
			break
		}
		stateStr += fmt.Sprintf("x%d=%.2f ", i, v)
	}
	b.WriteString(stateStr + "\n")

	fmt.Print(b.String())
}

func (r *LiveRenderer) Start() { fmt.Print(hideCursor) }
func (r *LiveRenderer) Stop()  { fmt.Print(showCursor) }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
