package compute

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type OpenGLBackend struct {
	Program       uint32
	RenderProgram uint32
	SSBOIn        uint32
	SSBOOut       uint32
	VAO           uint32
	NumParticles  int32
	Initialized   bool
}

func NewOpenGLBackend(numParticles int) *OpenGLBackend {
	return &OpenGLBackend{NumParticles: int32(numParticles)}
}

func (c *OpenGLBackend) InitRender(vertPath, fragPath string) error {
	program, err := createRenderProgram(vertPath, fragPath)
	if err != nil {
		return err
	}
	c.RenderProgram = program
	return nil
}

func (c *OpenGLBackend) Init(shaderPath string, initialData []float32) error {
	if err := gl.Init(); err != nil {
		return fmt.Errorf("failed to init opengl: %v", err)
	}

	program, err := createComputeProgram(shaderPath)
	if err != nil {
		return err
	}
	c.Program = program

	size := int(c.NumParticles) * 8 * 4

	gl.GenBuffers(1, &c.SSBOIn)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, c.SSBOIn)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size, gl.Ptr(initialData), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, c.SSBOIn)

	gl.GenBuffers(1, &c.SSBOOut)
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, c.SSBOOut)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size, nil, gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, c.SSBOOut)

	gl.GenVertexArrays(1, &c.VAO)
	c.Initialized = true

	var maxWorkGroupCount [3]int32
	var maxWorkGroupSize [3]int32
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 0, &maxWorkGroupCount[0])
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_SIZE, 0, &maxWorkGroupSize[0])
	fmt.Printf("OpenGL Compute Initialized. Max WorkGroups: %v, Max WorkGroup Size: %v\n", maxWorkGroupCount, maxWorkGroupSize)

	return nil
}

func (c *OpenGLBackend) Step(dt, softening, mouseX, mouseY, mouseStr float32) {
	if !c.Initialized {
		return
	}

	gl.UseProgram(c.Program)

	locDt := gl.GetUniformLocation(c.Program, gl.Str("dt\x00"))
	gl.Uniform1f(locDt, dt)

	locN := gl.GetUniformLocation(c.Program, gl.Str("numParticles\x00"))
	gl.Uniform1i(locN, c.NumParticles)

	locSoft := gl.GetUniformLocation(c.Program, gl.Str("softening\x00"))
	gl.Uniform1f(locSoft, softening)

	locMouse := gl.GetUniformLocation(c.Program, gl.Str("mouse\x00"))
	active := float32(0.0)
	if mouseStr != 0 {
		active = 1.0
	}
	gl.Uniform4f(locMouse, mouseX, mouseY, mouseStr, active)

	numGroups := (c.NumParticles + 255) / 256
	gl.DispatchCompute(uint32(numGroups), 1, 1)
	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)

	c.SSBOIn, c.SSBOOut = c.SSBOOut, c.SSBOIn
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, c.SSBOIn)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1, c.SSBOOut)
}

func createComputeProgram(path string) (uint32, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	content := string(source) + "\x00"

	shader := gl.CreateShader(gl.COMPUTE_SHADER)
	csources, free := gl.Strs(content)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile compute shader: %v", log)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, shader)
	gl.LinkProgram(program)

	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		return 0, fmt.Errorf("failed to link program")
	}

	gl.DeleteShader(shader)
	return program, nil
}

func createRenderProgram(vertPath, fragPath string) (uint32, error) {
	vSource, err := ioutil.ReadFile(vertPath)
	if err != nil {
		return 0, err
	}
	fSource, err := ioutil.ReadFile(fragPath)
	if err != nil {
		return 0, err
	}

	vContent := string(vSource) + "\x00"
	fContent := string(fSource) + "\x00"

	vShader := gl.CreateShader(gl.VERTEX_SHADER)
	vStrs, vFree := gl.Strs(vContent)
	gl.ShaderSource(vShader, 1, vStrs, nil)
	vFree()
	gl.CompileShader(vShader)

	fShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fStrs, fFree := gl.Strs(fContent)
	gl.ShaderSource(fShader, 1, fStrs, nil)
	fFree()
	gl.CompileShader(fShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vShader)
	gl.AttachShader(program, fShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		return 0, fmt.Errorf("failed to link render program")
	}

	gl.DeleteShader(vShader)
	gl.DeleteShader(fShader)
	return program, nil
}
