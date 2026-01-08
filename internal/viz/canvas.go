package viz

import (
	"strings"
)

// Braille Patterns: 2x4 dots
// 1 4
// 2 5
// 3 6
// 7 8
//
// Unicode offset 0x2800
var pixelMap = [4][2]int{
	{0x1, 0x8},
	{0x2, 0x10},
	{0x4, 0x20},
	{0x40, 0x80},
}

type Canvas struct {
	Width, Height int
	Grid          [][]rune
}

func NewCanvas(w, h int) *Canvas {
	c := &Canvas{
		Width:  w,
		Height: h,
		Grid:   make([][]rune, h),
	}
	for i := range c.Grid {
		c.Grid[i] = make([]rune, w)
		for j := range c.Grid[i] {
			c.Grid[i][j] = 0x2800 // Empty braille char
		}
	}
	return c
}

// SetPixel sets a pixel at (x, y) where x,y are in "sub-pixel" coordinates.
// The canvas size in sub-pixels is (Width*2) x (Height*4).
func (c *Canvas) Set(x, y int) {
	// Early bounds check for negative coordinates
	if x < 0 || y < 0 {
		return
	}

	col := x / 2
	row := y / 4
	if col >= c.Width || row >= c.Height {
		return
	}

	subX := x % 2
	subY := y % 4

	c.Grid[row][col] |= rune(pixelMap[subY][subX])
}

// Unset clears a pixel
func (c *Canvas) Unset(x, y int) {
	if x < 0 || y < 0 {
		return
	}

	col := x / 2
	row := y / 4
	if col >= c.Width || row >= c.Height {
		return
	}

	subX := x % 2
	subY := y % 4

	mask := ^rune(pixelMap[subY][subX])
	c.Grid[row][col] &= mask
	if c.Grid[row][col] < 0x2800 {
		c.Grid[row][col] = 0x2800
	}
}

// Clear resets the canvas
func (c *Canvas) Clear() {
	for i := range c.Grid {
		for j := range c.Grid[i] {
			c.Grid[i][j] = 0x2800
		}
	}
}

// DrawLine draws a line using Bresenham's algorithm
func (c *Canvas) DrawLine(x0, y0, x1, y1 int) {
	dx := absInt(x1 - x0)
	dy := absInt(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		c.Set(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (c *Canvas) String() string {
	var b strings.Builder
	for _, row := range c.Grid {
		b.WriteString(string(row) + "\n")
	}
	return b.String()
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
