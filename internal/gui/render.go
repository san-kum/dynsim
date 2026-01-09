package gui

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-gl/gl/v4.3-core/gl"
)

func (a *App) RenderSPH() {
	rl.BeginBlendMode(rl.BlendAdditive)
	n := len(a.State) / 4
	for i := 0; i < n; i++ {
		x, y, vx, vy := float32(a.State[i*4]), float32(a.State[i*4+1]), a.State[i*4+2], a.State[i*4+3]
		pos := rl.NewVector3(x, y, 0)
		vel := math.Sqrt(vx*vx + vy*vy)
		var col rl.Color
		if vel < 5.0 {
			col = rl.ColorFromNormalized(rl.Vector4{X: 0.0, Y: 0.2 + float32(vel/10.0), Z: 0.8, W: 0.2})
		} else if vel < 15.0 {
			t := float32((vel - 5.0) / 10.0)
			col = rl.ColorFromNormalized(rl.Vector4{X: 0.0, Y: 0.7 + t*0.3, Z: 1.0, W: 0.3})
		} else {
			col = rl.ColorFromNormalized(rl.Vector4{X: 0.8, Y: 0.9, Z: 1.0, W: 0.5})
		}
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, 2.5, col)
	}
	rl.EndBlendMode()
	rl.DrawCubeWires(rl.NewVector3(30, 20, 0), 60, 40, 2, rl.ColorAlpha(rl.Gray, 0.5))
}

func (a *App) RenderNBody() {
	a.RenderStarfield()
	rl.BeginBlendMode(rl.BlendAdditive)
	n := len(a.State) / 4
	for i := 0; i < n; i++ {
		x, y, vx, vy := a.State[i*4], a.State[i*4+1], a.State[i*4+2], a.State[i*4+3]
		vel := math.Sqrt(vx*vx + vy*vy)
		var col rl.Color
		if vel < 2.0 {
			t := float32(vel / 2.0)
			col = rl.ColorFromNormalized(rl.Vector4{X: 1.0, Y: t * 0.5, Z: 0.1, W: 0.8})
		} else if vel < 6.0 {
			t := float32((vel - 2.0) / 4.0)
			col = rl.ColorFromNormalized(rl.Vector4{X: 1.0, Y: 0.5 + t*0.5, Z: 0.1 + t*0.9, W: 0.9})
		} else {
			t := float32((vel - 6.0) / 4.0)
			if t > 1.0 {
				t = 1.0
			}
			col = rl.ColorFromNormalized(rl.Vector4{X: 1.0 - t*0.2, Y: 1.0 - t*0.1, Z: 1.0, W: 1.0})
		}
		pos := rl.NewVector3(float32(x), float32(y), 0)
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, 0.4, col)
		haloCol := col
		haloCol.A = 10
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, 2.0, haloCol)
		if a.ShowVectors {
			end := rl.NewVector3(float32(x+vx*0.5), float32(y+vy*0.5), 0)
			rl.DrawLine3D(pos, end, rl.Fade(rl.White, 0.5))
		}
	}
	rl.EndBlendMode()
}

func (a *App) RenderStarfield() {
	n := len(a.Stars) / 3
	for i := 0; i < n; i++ {
		pos := rl.NewVector3(float32(a.Stars[i*3]), float32(a.Stars[i*3+1]), float32(a.Stars[i*3+2]))
		rl.DrawPixel(int32(0), int32(0), rl.White)
		rl.DrawLine3D(pos, rl.NewVector3(pos.X, pos.Y+0.1, pos.Z), rl.NewColor(150, 150, 150, 100))
	}
}

func (a *App) RenderHybrid() {
	a.RenderStarfield()
	rl.BeginBlendMode(rl.BlendAdditive)
	bass := float32(a.Audio.Bass)
	mid := float32(a.Audio.Mid)
	high := float32(a.Audio.High)
	nStars := 8192
	nGas := 4096
	if len(a.State) != (nStars+nGas)*4 {
		return
	}
	for i := 0; i < nStars; i++ {
		idx := i * 4
		x, y := a.State[idx], a.State[idx+1]
		vx, vy := a.State[idx+2], a.State[idx+3]
		vel := math.Sqrt(vx*vx + vy*vy)
		var col rl.Color
		if vel < 2.0 {
			t := float32(vel / 2.0)
			col = rl.ColorFromNormalized(rl.Vector4{X: 1.0, Y: t * 0.5, Z: 0.1, W: 0.8})
		} else {
			t := float32((vel - 2.0) / 4.0)
			col = rl.ColorFromNormalized(rl.Vector4{X: 1.0, Y: 0.5 + t*0.5 + high*0.5, Z: 0.1 + t*0.9 + high, W: 0.9})
		}
		pos := rl.NewVector3(float32(x), float32(y), 0)
		size := float32(0.4) + mid*0.4
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, size, col)
	}
	baseIdx := nStars * 4
	for i := 0; i < nGas; i++ {
		idx := baseIdx + i*4
		x, y := a.State[idx], a.State[idx+1]
		r := uint8(200 - bass*100)
		g := uint8(50 + high*200)
		b := uint8(255)
		alpha := uint8(30 + bass*50)
		col := rl.NewColor(r, g, b, alpha)
		pos := rl.NewVector3(float32(x), float32(y), 0)
		size := float32(2.0) + bass*2.0
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, size, col)
		rl.DrawBillboard(a.Camera, a.ParticleTex, pos, size*2.0, rl.NewColor(100, 0, 200, 10))
	}
	rl.EndBlendMode()
}

func (a *App) RenderAttractor() {
	a.RenderTrails(3)
	if len(a.State) < 2 {
		return
	}
	x, y, z := float32(a.State[0]), float32(a.State[1]), float32(0)
	if len(a.State) >= 3 {
		z = float32(a.State[2])
	}
	rl.BeginBlendMode(rl.BlendAdditive)
	rl.DrawSphere(rl.NewVector3(x, y, z), 0.5, rl.White)
	rl.EndBlendMode()
}

func (a *App) RenderTrails(dim int) {
	if len(a.History) < 2 {
		return
	}
	rl.BeginBlendMode(rl.BlendAdditive)
	n := len(a.History[0]) / dim
	for i := 0; i < n; i++ {
		for h := 0; h < len(a.History)-1; h++ {
			s1 := a.History[h]
			s2 := a.History[h+1]
			alpha := uint8(float32(h) / float32(a.MaxHistory) * 50)
			col := rl.NewColor(255, 255, 255, alpha)
			var p1, p2 rl.Vector3
			if dim >= 3 {
				p1 = rl.NewVector3(float32(s1[i*dim]), float32(s1[i*dim+1]), float32(s1[i*dim+2]))
				p2 = rl.NewVector3(float32(s2[i*dim]), float32(s2[i*dim+1]), float32(s2[i*dim+2]))
			} else {
				p1 = rl.NewVector3(float32(s1[i*dim]), float32(s1[i*dim+1]), 0)
				p2 = rl.NewVector3(float32(s2[i*dim]), float32(s2[i*dim+1]), 0)
			}
			rl.DrawLine3D(p1, p2, col)
		}
	}
	rl.EndBlendMode()
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
	a.RenderTrails(4)
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

func (a *App) RenderComputeNBody() {
	if a.GLBackend == nil || !a.GLBackend.Initialized || a.GLBackend.RenderProgram == 0 {
		return
	}
	gl.UseProgram(a.GLBackend.RenderProgram)
	matView := rl.GetCameraMatrix(a.Camera)
	matProj := rl.GetMatrixProjection()
	matMVP := rl.MatrixMultiply(matView, matProj)
	locMVP := gl.GetUniformLocation(a.GLBackend.RenderProgram, gl.Str("mvp\x00"))
	gl.UniformMatrix4fv(locMVP, 1, false, &matMVP.M0)
	gl.BindVertexArray(a.GLBackend.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.GLBackend.SSBOIn)
	gl.Enable(gl.PROGRAM_POINT_SIZE)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 32, nil)
	rl.BeginBlendMode(rl.BlendAdditive)
	gl.DrawArrays(gl.POINTS, 0, a.GLBackend.NumParticles)
	rl.EndBlendMode()
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.Disable(gl.PROGRAM_POINT_SIZE)
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.UseProgram(0)
}
