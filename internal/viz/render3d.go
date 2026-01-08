package viz

import (
	"math"
	"sort"
)

type Vec3 struct {
	X, Y, Z float64
}

// Vec3 methods.
func (v Vec3) Add(o Vec3) Vec3      { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec3) Sub(o Vec3) Vec3      { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec3) Scale(s float64) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Length() float64      { return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }
func (v Vec3) Normalize() Vec3 {
	if l := v.Length(); l != 0 {
		return v.Scale(1 / l)
	}
	return Vec3{}
}
func (v Vec3) Dot(o Vec3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }
func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{v.Y*o.Z - v.Z*o.Y, v.Z*o.X - v.X*o.Z, v.X*o.Y - v.Y*o.X}
}

// Camera manages 3D projection to a 2D plane.
type Camera struct {
	Position, Target, Up Vec3
	FOV, Near, Far       float64
	RotX, RotY, RotZ     float64
	Zoom                 float64
}

func NewCamera() *Camera {
	return &Camera{Position: Vec3{0, 0, 50}, Up: Vec3{0, 1, 0}, FOV: math.Pi / 4, Near: 0.1, Far: 1000, Zoom: 1.0}
}

func (c *Camera) RotateX(a float64) { c.RotX += a }
func (c *Camera) RotateY(a float64) { c.RotY += a }
func (c *Camera) RotateZ(a float64) { c.RotZ += a }
func (c *Camera) ZoomIn()           { c.Zoom = math.Min(10, c.Zoom*1.2) }
func (c *Camera) ZoomOut()          { c.Zoom = math.Max(0.1, c.Zoom/1.2) }

// RotatePoint rotates a point around the camera's axes.
func (c *Camera) RotatePoint(p Vec3) Vec3 {
	cx, sx := math.Cos(c.RotX), math.Sin(c.RotX)
	p.Y, p.Z = p.Y*cx-p.Z*sx, p.Y*sx+p.Z*cx
	cy, sy := math.Cos(c.RotY), math.Sin(c.RotY)
	p.X, p.Z = p.X*cy+p.Z*sy, -p.X*sy+p.Z*cy
	cz, sz := math.Cos(c.RotZ), math.Sin(c.RotZ)
	p.X, p.Y = p.X*cz-p.Y*sz, p.X*sz+p.Y*cz
	return p
}

// Project converts 3D world coordinates to 2D screen coordinates.
// Returns x, y, depth, and visibility.
func (c *Camera) Project(p Vec3, sw, sh int) (int, int, float64, bool) {
	rot := c.RotatePoint(p).Scale(c.Zoom)
	dist := c.Position.Z
	if rot.Z >= dist-c.Near {
		return 0, 0, 0, false
	}
	scale := dist / (dist - rot.Z)
	minDim := float64(sh)
	if float64(sw) < minDim {
		minDim = float64(sw)
	}
	pScale := minDim / 3.0
	sx := int(rot.X*scale*pScale) + sw/2
	sy := int(-rot.Y*scale*pScale) + sh/2
	return sx, sy, rot.Z, sx >= 0 && sx < sw && sy >= 0 && sy < sh
}

type Edge struct {
	Start, End Vec3
	Color      rune
}

type Wireframe struct{ Edges []Edge }

func NewWireframe() *Wireframe                 { return &Wireframe{Edges: make([]Edge, 0)} }
func (w *Wireframe) AddEdge(s, e Vec3, c rune) { w.Edges = append(w.Edges, Edge{s, e, c}) }
func (w *Wireframe) AddPoint(p Vec3, c rune)   { w.Edges = append(w.Edges, Edge{p, p, c}) }
func (w *Wireframe) Clear()                    { w.Edges = w.Edges[:0] }

type ProjectedEdge struct {
	X1, Y1, X2, Y2 int
	Depth          float64
	Color          rune
	Visible        bool
}

// Render3D draws the wireframe to the canvas using a simple painter's algorithm.
func Render3D(c *Canvas, w *Wireframe, cam *Camera) {
	if c == nil || w == nil || cam == nil {
		return
	}
	cw, ch := c.Width, c.Height
	proj := make([]ProjectedEdge, 0, len(w.Edges))
	for _, e := range w.Edges {
		x1, y1, d1, v1 := cam.Project(e.Start, cw, ch)
		x2, y2, d2, v2 := cam.Project(e.End, cw, ch)
		if v1 || v2 {
			proj = append(proj, ProjectedEdge{x1, y1, x2, y2, (d1 + d2) / 2, e.Color, true})
		}
	}
	sort.Slice(proj, func(i, j int) bool { return proj[i].Depth < proj[j].Depth })
	for _, e := range proj {
		if e.X1 == e.X2 && e.Y1 == e.Y2 {
			c.Set(e.X1, e.Y1)
		} else {
			c.DrawLine(e.X1, e.Y1, e.X2, e.Y2)
		}
	}
}

func CreateCubeWireframe(size float64) *Wireframe {
	w, s := NewWireframe(), size/2
	v := []Vec3{{-s, -s, -s}, {s, -s, -s}, {s, s, -s}, {-s, s, -s}, {-s, -s, s}, {s, -s, s}, {s, s, s}, {-s, s, s}}
	ei := [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 7}, {7, 4}, {0, 4}, {1, 5}, {2, 6}, {3, 7}}
	for _, e := range ei {
		w.AddEdge(v[e[0]], v[e[1]], '█')
	}
	return w
}

func CreateAxesWireframe(l float64) *Wireframe {
	w, o := NewWireframe(), Vec3{0, 0, 0}
	w.AddEdge(o, Vec3{l, 0, 0}, 'X')
	w.AddEdge(o, Vec3{0, l, 0}, 'Y')
	w.AddEdge(o, Vec3{0, 0, l}, 'Z')
	return w
}
