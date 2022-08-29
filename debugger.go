package gocompute

import (
	"errors"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"math"
)

type debugger struct {
	opened   bool
	dataSize int
}

var vertexShader = "#version 430\n" +
	"precision mediump float;\n" +
	"in vec4 vPosition;\n" +
	"void main() {\n" +
	"gl_Position = vPosition;\n" +
	"}\n"

func CreateDebugger() *debugger {
	output := &debugger{}
	output.dataSize = 100 * 800
	return output
}

func (d *debugger) StartWindow() {
	d.opened = true
	d.debugWindow()
}

func (d *debugger) CloseWindow() {
	d.opened = false
}

func (d *debugger) debugWindow() {
	go func() {
		for {
			println("opening")
			err := glfw.Init()
			if err != nil {
				println(errors.New("failed to initialize glfw: " + err.Error()))
			}
			glfw.WindowHint(glfw.ContextVersionMajor, 4)
			glfw.WindowHint(glfw.ContextVersionMinor, 3)
			glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
			glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
			glfw.WindowHint(glfw.Visible, glfw.True)
			size := int(math.Sqrt(float64(d.dataSize)))

			window, err := glfw.CreateWindow(size, size, "Debug", nil, nil)
			if err != nil {
				glfw.Terminate()
				println(errors.New("failed to create window: " + err.Error()))
			}

			window.MakeContextCurrent()
			_ = gl.Init()

			cnt := 0
			for d.opened {
				col := float32(cnt%255) / 255.0
				gl.ClearColor(col, 0.5, 0.5, 0.5)
				gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

				window.SwapBuffers()
				size = int(math.Sqrt(float64(d.dataSize)))
				window.SetSize(size, size)
				gl.Viewport(0, 0, int32(size), int32(size))
				d.dataSize += 40
				cnt++
				//opened = <-d.changeDebug
			}

		}
	}()

}
