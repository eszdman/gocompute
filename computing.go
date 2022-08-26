package gocompute

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"strconv"
)

type computeGroup struct {
	X, Y, Z int
}

type computing struct {
	version        string
	programCounter int
	programs       map[int]uint32
	computeGroups  []computeGroup
	defines        []string
	defineNames    []string
}

func checkErr(operation string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		msg := operation + ": glError: " + strconv.Itoa(int(err))
		log.Println(msg)
	}
}
func NewComputing(createContext bool) (*computing, error) {
	compute := &computing{}
	compute.version = "#version 430"
	compute.programs = make(map[int]uint32)
	compute.defines = make([]string, 0)
	compute.defineNames = make([]string, 0)
	if createContext {
		err := glfw.Init()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize glfw: %v", err)
		}
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
		glfw.WindowHint(glfw.Maximized, glfw.True)
		glfw.WindowHint(glfw.Visible, glfw.False)
		window, err := glfw.CreateWindow(1, 1, "computing", nil, nil)
		if err != nil {
			glfw.Terminate()
			return nil, errors.New("failed to create window: " + err.Error())
		}
		window.MakeContextCurrent()
		err = gl.Init()
		if err != nil {
			return nil, err
		}
	}
	return compute, nil
}

func compileShader(shaderType int, shaderProgram string) (uint32, error) {
	shaderHandle := gl.CreateShader(uint32(shaderType))
	if shaderHandle != 0 {
		point, free := gl.Strs(shaderProgram + "\x00")
		defer free()
		gl.ShaderSource(shaderHandle, 1, point, nil)
		gl.CompileShader(shaderHandle)
		compileStatus := int32(0)
		gl.GetShaderiv(shaderHandle, gl.COMPILE_STATUS, &compileStatus)
		if compileStatus == gl.FALSE {
			outLen := int32(0)
			var infoLog [1024]byte
			gl.GetShaderInfoLog(shaderHandle, 1024, &outLen, &(infoLog[0]))
			errStr := "Error compiling shader: " + string(infoLog[:])
			gl.DeleteShader(shaderHandle)
			return 0, errors.New(errStr)
		}
	} else {
		return 0, errors.New("error creating shader")
	}
	return shaderHandle, nil
}
func (c *computing) LoadProgram(programText string) (int, error) {
	c.programCounter++
	count := c.programCounter - 1
	shaderHandle, err := compileShader(gl.COMPUTE_SHADER, programText)
	if err != nil {
		return 0, err
	}
	program := gl.CreateProgram()
	gl.AttachShader(program, shaderHandle)
	gl.LinkProgram(program)
	c.programs[count] = program
	return count, nil
}
func (c *computing) Define(Name string, value string) {
	c.defines = append(c.defines, value)
	c.defineNames = append(c.defineNames, Name)
}

func (c *computing) DefineInt(Name string, value int) {
	c.Define(Name, strconv.Itoa(value))
}

func (c *computing) DefineFloat(Name string, value float64) {
	c.Define(Name, strconv.FormatFloat(value, 'e', 8, 32))
}

func (c *computing) DefineDouble(Name string, value float64) {
	c.Define(Name, strconv.FormatFloat(value, 'e', 8, 64))
}

func (c *computing) UseProgram(programNumber int) {
	if len(c.defines) > 0 {
		log.Println("Warning: using defines with preloaded program")
	}
	gl.UseProgram(c.programs[programNumber])
}
func (c *computing) Realize(x, y, z int) {
	gl.DispatchCompute(uint32(x), uint32(y), uint32(z))
}
func (c *computing) UseLoadProgram(programText string) {
	c.defines = make([]string, 0)
	c.defineNames = make([]string, 0)
	shaderHandle, _ := compileShader(gl.COMPUTE_SHADER, programText)
	program := gl.CreateProgram()
	gl.AttachShader(program, shaderHandle)
	gl.LinkProgram(program)
}
func (c *computing) Close() {

}
