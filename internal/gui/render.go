package gui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (a *App) RenderSPH() {
	n := len(a.State) / 4
	for i := 0; i < n; i++ {
		x, y := float32(a.State[i*4]), float32(a.State[i*4+1])
		pos := rl.NewVector3(x, y, 0)
		rl.DrawSphere(pos, 0.4, rl.NewColor(255, 255, 255, 200))
	}
	rl.DrawCubeWires(rl.NewVector3(30, 20, 0), 60, 40, 2, rl.ColorAlpha(rl.Gray, 0.5))
}

func (a *App) RenderNBody() {
	n := len(a.State) / 4
	for i := 0; i < n; i++ {
		x, y, vx, vy := a.State[i*4], a.State[i*4+1], a.State[i*4+2], a.State[i*4+3]
		vel := math.Sqrt(vx*vx + vy*vy)
		val := uint8(math.Min(100+vel*50, 255))

		pos := rl.NewVector3(float32(x), float32(y), 0)
		rl.DrawSphere(pos, 0.5, rl.NewColor(val, val, val, 255))
	}
}

func (a *App) RenderAttractor() {
	if len(a.State) < 2 {
		return
	}
	x, y, z := float32(a.State[0]), float32(a.State[1]), float32(0)
	if len(a.State) >= 3 {
		z = float32(a.State[2])
	}
	rl.DrawSphere(rl.NewVector3(x, y, z), 0.5, rl.White)
}

func (a *App) RenderDoubleWell() {
	x := float32(a.State[0])
	rl.DrawSphere(rl.NewVector3(x, 2, 0), 0.5, rl.White)

	for i := -20; i <= 20; i++ {
		xv := float32(i) / 10.0
		val := xv*xv*xv*xv - xv*xv
		xv2 := float32(i+1) / 10.0
		val2 := xv2*xv2*xv2*xv2 - xv2*xv2
		rl.DrawLine3D(rl.NewVector3(xv, val, 0), rl.NewVector3(xv2, val2, 0), rl.Gray)
	}
}

func (a *App) RenderPendulum() {
	theta, length := a.State[0], 5.0
	x, y := math.Sin(theta)*length, -math.Cos(theta)*length

	origin, bob := rl.NewVector3(0, 5, 0), rl.NewVector3(float32(x), float32(y)+5, 0)

	rl.DrawLine3D(origin, bob, rl.Gray)
	rl.DrawSphere(bob, 0.5, rl.White)
	rl.DrawSphere(origin, 0.2, rl.Gray)
}

func (a *App) RenderDoublePendulum() {
	th1, th2 := a.State[0], a.State[1]
	l1, l2 := 5.0, 5.0

	x1, y1 := math.Sin(th1)*l1, -math.Cos(th1)*l1
	x2, y2 := x1+math.Sin(th2)*l2, y1-math.Cos(th2)*l2

	origin := rl.NewVector3(0, 10, 0)
	p1 := rl.NewVector3(float32(x1), float32(y1)+10, 0)
	p2 := rl.NewVector3(float32(x2), float32(y2)+10, 0)

	rl.DrawLine3D(origin, p1, rl.Gray)
	rl.DrawLine3D(p1, p2, rl.Gray)
	rl.DrawSphere(p1, 0.5, rl.LightGray)
	rl.DrawSphere(p2, 0.5, rl.White)
}

func (a *App) RenderCartPole() {
	pos, theta := a.State[0], a.State[2]
	cartX, poleLen := float32(pos), float32(5.0)

	tipX := cartX + float32(math.Sin(theta))*poleLen
	tipY := float32(math.Cos(theta)) * poleLen

	cartPos, tipPos := rl.NewVector3(cartX, 0, 0), rl.NewVector3(tipX, tipY, 0)

	rl.DrawCube(cartPos, 2, 1, 1, rl.LightGray)
	rl.DrawLine3D(cartPos, tipPos, rl.Gray)
	rl.DrawSphere(tipPos, 0.3, rl.White)
	rl.DrawLine3D(rl.NewVector3(-10, -0.5, 0), rl.NewVector3(10, -0.5, 0), rl.DarkGray)
}

func (a *App) RenderSpringMass() {
	pos := a.State[0]
	anchor := rl.NewVector3(0, 10, 0)
	mass := rl.NewVector3(0, float32(pos)+5, 0)

	rl.DrawLine3D(anchor, mass, rl.White)
	rl.DrawSphere(mass, 0.5, rl.White)
	rl.DrawCube(anchor, 2, 0.5, 2, rl.Gray)
}

func (a *App) RenderDrone() {
	x, y, th := a.State[0], a.State[1], a.State[2]
	pos := rl.NewVector3(float32(x), float32(y), 0)
	rl.DrawCube(pos, 1.0, 0.2, 1.0, rl.White)

	armLen := float32(0.8)
	c, s := float32(math.Cos(th)), float32(math.Sin(th))

	p1 := rl.NewVector3(pos.X+c*armLen, pos.Y+s*armLen, 0)
	p2 := rl.NewVector3(pos.X-c*armLen, pos.Y-s*armLen, 0)

	rl.DrawSphere(p1, 0.2, rl.LightGray)
	rl.DrawSphere(p2, 0.2, rl.LightGray)
}

func (a *App) RenderThreeBody() {
	for i := 0; i < 3; i++ {
		x, y := a.State[i*4], a.State[i*4+1]
		col := rl.White
		if i == 1 {
			col = rl.LightGray
		}
		if i == 2 {
			col = rl.Gray
		}
		rl.DrawSphere(rl.NewVector3(float32(x), float32(y), 0), 0.4, col)
	}
}

func (a *App) RenderCoupled() {
	th1, th2 := a.State[0], a.State[2]
	y1, y2 := -math.Cos(th1)*5, -math.Cos(th2)*5
	x1, x2 := math.Sin(th1)*5-3, math.Sin(th2)*5+3

	p1 := rl.NewVector3(float32(x1), float32(y1)+5, 0)
	p2 := rl.NewVector3(float32(x2), float32(y2)+5, 0)

	rl.DrawLine3D(rl.NewVector3(-3, 5, 0), p1, rl.Gray)
	rl.DrawLine3D(rl.NewVector3(3, 5, 0), p2, rl.Gray)
	rl.DrawSphere(p1, 0.5, rl.White)
	rl.DrawSphere(p2, 0.5, rl.White)
	rl.DrawLine3D(p1, p2, rl.LightGray)
}

func (a *App) RenderMassChain() {
	n := len(a.State) / 2
	scale := 40.0 / float32(n)
	for i := 0; i < n; i++ {
		y := float32(a.State[i*2])
		x := float32(i) * scale
		pos := rl.NewVector3(x, y+10, 0)
		rl.DrawSphere(pos, 0.2, rl.White)
		if i > 0 {
			prevX := float32(i-1) * scale
			prevY := float32(a.State[(i-1)*2])
			rl.DrawLine3D(rl.NewVector3(prevX, prevY+10, 0), pos, rl.Gray)
		}
	}
}

func (a *App) RenderGyroscope() {
	phi, theta := a.State[0], a.State[1]
	len := float32(5.0)
	x := len * float32(math.Sin(theta)*math.Sin(phi))
	y := len * float32(math.Cos(theta))
	z := len * float32(math.Sin(theta)*math.Cos(phi))

	origin := rl.NewVector3(0, 0, 0)
	tip := rl.NewVector3(x, y, z)

	rl.DrawLine3D(origin, tip, rl.LightGray)
	rl.DrawSphere(tip, 0.5, rl.White)
	rl.DrawSphere(origin, 0.2, rl.Gray)
	rl.DrawCircle3D(rl.NewVector3(0, -2, 0), 2.0, rl.NewVector3(1, 0, 0), 90, rl.DarkGray)
}

func (a *App) RenderWave() {
	n := len(a.State) / 2
	scale := 60.0 / float32(n)
	for i := 0; i < n; i++ {
		u := float32(a.State[i*2])
		x := float32(i) * scale
		pos := rl.NewVector3(x, u+10, 0)
		rl.DrawSphere(pos, 0.2, rl.White)
	}
}

func (a *App) RenderDuffing() {
	x := float32(a.State[0])
	rl.DrawSphere(rl.NewVector3(x, 0, 0), 0.5, rl.White)
	rl.DrawLine3D(rl.NewVector3(-2, 0, 0), rl.NewVector3(2, 0, 0), rl.Gray)
}

func (a *App) RenderMagnetic() {
	x, y := float32(a.State[0]), float32(a.State[1])
	rl.DrawSphere(rl.NewVector3(x, y, 0), 0.5, rl.White)
	rl.DrawSphere(rl.NewVector3(1, 0, 0), 0.2, rl.Gray)
	rl.DrawSphere(rl.NewVector3(-0.5, 0.866, 0), 0.2, rl.Gray)
	rl.DrawSphere(rl.NewVector3(-0.5, -0.866, 0), 0.2, rl.Gray)
}

func (a *App) RenderGeneric() {
	for i, v := range a.State {
		if i >= 3 {
			break
		}
		rl.DrawSphere(rl.NewVector3(float32(i)*2, float32(v), 0), 0.5, rl.White)
	}
}
