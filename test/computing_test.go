package test

import (
	"embed"
	_ "embed"
	gc "github.com/eszdman/gocompute"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"math"
	"strings"
	"testing"
	"time"
)

//go:embed resources/bufferTest.glsl
var bufferTest string

//go:embed resources/bufferTest2.glsl
var bufferTest2 string

//go:embed resources/bufferTest3.glsl
var bufferTest3 string

//go:embed resources/textureTest.glsl
var textureTest string

//go:embed resources/textureTest2.glsl
var textureTest2 string

//go:embed resources/functionsTest.glsl
var functionsTest string

//go:embed resources/speedTest2.glsl
var speedTest2 string

//go:embed resources/include/*
var includes embed.FS

func logLoad(compute *gc.Computing, program string) int {
	programID, err := compute.LoadProgram(program)
	if err != nil {
		log.Println("E", err)
		return -1
	}
	return programID
}
func BufferExample(compute *gc.Computing, program int) {
	log.Println("D", "BufferExample started")
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	//Allocate buffer data
	buffer2.AllocateFloat32(9)
	//Load data into buffer instead of allocation
	target := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	buffer.LoadFloat32(target)
	//Change current program to selected
	compute.UseProgram(program)
	//Bind buffer to layout binding
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	//Run program with size
	compute.Realize(buffer2.Size, 1, 1)
	read := buffer2.ReadFloat32(buffer2.Size)
	log.Println("D", read)
	for i := 0; i < len(read); i++ {
		val := target[i] + float32(i)
		if val != read[i] {
			log.Println("E", "Wrong action", "ind:", i, "val input:", val, "val output", read[i])
		}
	}
	buffer.Close()
	buffer2.Close()
}
func BufferExample2(compute *gc.Computing, program int) {
	log.Println("D", "BufferExample2 started")
	//Possible to set buffer memory usage hint
	buffer := compute.NewBufferV(gc.STATIC_READ)
	buffer2 := compute.NewBufferV(gc.STATIC_WRITE)

	//Possible to use any structure
	gc.BufferAllocate[gc.Vec2](buffer2, 9)
	points := make([]gc.Vec2, 9)
	//Write into first point for example
	points[0].X = 0.25
	points[0].Y = 0.5
	gc.BufferLoad(buffer, points)

	compute.UseProgram(program)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println("D", gc.BufferRead[gc.Vec2](buffer, 9))
	log.Println("D", gc.BufferRead[gc.Vec2](buffer2, 9))
	buffer.Close()
	buffer2.Close()
}
func BufferExample3(compute *gc.Computing, program int) {
	log.Println("D", "BufferExample3 started")
	buffer := compute.NewBufferV(gc.STATIC_READ)
	buffer2 := compute.NewBufferV(gc.STATIC_WRITE)
	//Generic gc method

	//It's possible to pass any element structures into buffer memory
	gc.BufferAllocate[gc.Vec4](buffer2, 9)
	points := make([]gc.Vec4, 9)
	//Write into first point for example
	points[0].X = 0.25
	points[0].Y = 0.5
	points[0].Z = 0.75
	points[0].W = 1.0
	gc.BufferLoad(buffer, points)

	compute.UseProgram(program)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	//compute.SetFloat32("test", 0.5)
	compute.Realize(buffer2.Size, 1, 1)
	log.Println("D", gc.BufferRead[gc.Vec4](buffer, buffer.Size))
	log.Println("D", gc.BufferRead[gc.Vec4](buffer2, buffer2.Size))
	buffer.Close()
	buffer2.Close()
}
func TextureExample(compute *gc.Computing, program int) {
	log.Println("D", "TextureExample started")
	texture := compute.NewTexture(gc.FLOAT32, 4)
	texture.Create1D(2)
	input := []float32{1, 1, 1, 1, 0, 0, 0, 0}
	gc.InternalBuffer[float32](texture)
	texture.Load1DFloat32(input)
	compute.UseProgram(program)
	texture.SetBinding(0)
	//Add offset to gl_GlobalInvocationID
	compute.SetOffset(1, 0, 0)
	compute.Realize(2, 1, 1)
	read := texture.ReadFloat32()
	log.Println("D", texture.ReadFloat32())
	for i := 0; i < len(read); i++ {
		idx := 1.0 - float32(i/4)/10.0
		val := input[i] + idx
		if val != read[i] {
			log.Println("E", "Wrong action", "ind:", i, "val input:", val, "val output", read[i])
		}
	}
	//vec4 in texture(0,0) should be unchanged because of offset
	texture.Close()
}
func TextureExample2(compute *gc.Computing, program int) {
	log.Println("D", "TextureExample2 started")
	texture := compute.NewTexture(gc.FLOAT32, 4)
	points := make([]gc.Vec4, 2)
	//Write into first point for example
	points[0].X = 0.25
	points[0].Y = 0.5
	texture.Create1D(2)
	gc.InternalBuffer[float32](texture)
	//It's possible to pass structures into texture memory
	gc.TextureLoad1D(texture, points)
	compute.UseProgram(program)
	texture.SetBinding(0)
	compute.Realize(2, 1, 1)
	log.Println("D", texture.ReadFloat32())
	log.Println("D", gc.TextureRead[float32](texture))
	texture.Close()
}

func FunctionsExample(compute *gc.Computing, program int) {
	log.Println("D", "FunctionsExample started")
	buffer := compute.NewBuffer()
	//10 elements buffer zero initialized
	buffer.LoadFloat32(make([]float32, 16))
	compute.UseProgram(program)
	compute.SetInt("N0", 1)
	buffer.SetBinding(0)
	//Compute from 1 to 15
	compute.SetOffset(1, 0, 0)
	compute.Realize(14, 1, 1)
	log.Println("D", buffer.ReadFloat32(buffer.Size))
	//compare with cpu implementation
	test := make([]float32, 16)
	for i := 1; i < 15; i++ {
		//Convert integer range 0-16 to float 0-1
		idf := float64(i) / 15.0
		//Center range -0.5 - 0.5
		idf -= 0.5
		test[i] = 1.0 / float32(math.Exp(idf*idf))
	}
	//Comparing GPU simple fast approximation of gaussian function (1.0/(x*x+1)) with CPU result (1.0/exp(x*x))
	log.Println("D", test)
	buffer.Close()
}
func SpeedTest(compute *gc.Computing, program int) {
	log.Println("D", "SpeedTest started")
	buffer := compute.NewBufferV(gc.STATIC_READ)
	buffer2 := compute.NewBufferV(gc.STATIC_WRITE)
	//Allocate buffer data, maximal buffer 65535
	elementsCount := 65535
	buffer2.AllocateFloat32(elementsCount)
	//Load data into buffer instead of allocation
	b := make([]float32, elementsCount)
	for i := 0; i < elementsCount; i++ {
		b[i] = float32(i)
	}
	buffer.LoadFloat32(b)
	//Change current program to selected
	compute.UseProgram(program)
	//Bind buffer to layout binding
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	//Run program with size
	msStart := time.Now().UnixNano() / int64(time.Nanosecond)
	compute.Realize(buffer2.Size, 1, 1)

	msEnd := time.Now().UnixNano() / int64(time.Nanosecond)
	ns := msEnd - msStart
	log.Println(gc.BufferRead[float32](buffer2, buffer2.Size)[elementsCount-1])
	log.Println("D", "GPU Speed test")
	log.Println("D", "Time elapsed:", ns, "ns")
	count := float32(elementsCount) / 1000000
	log.Println("D", "Operations count:", count, "M")
	seconds := float64(ns) / float64(time.Second)
	log.Println("D", "Sum per second:", uint64(math.Round(float64(count)/seconds)), "M/s")

	in1 := b
	in2 := make([]float32, elementsCount)
	//Compare with CPU
	msStart = time.Now().UnixNano() / int64(time.Nanosecond)
	for ind := 0; ind < elementsCount; ind++ {
		in2[ind] = float32(ind) + in1[ind]
	}

	msEnd = time.Now().UnixNano() / int64(time.Nanosecond)

	log.Println("D", "CPU Speed test")
	ns = msEnd - msStart
	log.Println("D", "Time elapsed:", ns, "ns")
	log.Println("D", "Time elapsed:", ns/1000, "ns")
	count = float32(elementsCount) / 1000000
	seconds = float64(ns) / float64(time.Second)
	log.Println("D", "Operations count:", count, "M")
	log.Println("D", "Sum per second:", uint64(math.Round(float64(count)/seconds)), "M/s")
	buffer.Close()
	buffer2.Close()
}

func SpeedTest2(compute *gc.Computing, program int) {
	log.Println("D", "SpeedTest2 started")
	elementsCount := 8000
	elementsCount2 := 8000
	var b = make([]float32, elementsCount*elementsCount2)
	var out = make([]float32, elementsCount*elementsCount2)
	buffer := compute.NewTexture(gc.FLOAT32, 1)
	buffer2 := compute.NewTexture(gc.FLOAT32, 1)
	buffer2.Create2D(elementsCount, elementsCount2)
	buffer.Create2D(elementsCount, elementsCount2)
	// Init input-output slices before the loop to avoid memory leak
	gc.InternalBuffer[float32](buffer) //Internal buffer is required for reading texture
	//Buffer increase stability against garbage collector but memory consumption also
	gc.InternalBuffer[float32](buffer2)

	for j := 1; j < 1000; j++ {
		//Load data into buffer instead of allocation
		for i := 0; i < elementsCount*elementsCount2; i++ {
			b[i] = float32(i)
		}

		buffer.Load2DFloat32(b)

		if gc.TextureRead[float32](buffer)[elementsCount*elementsCount2-1] != float32(elementsCount*elementsCount2-1) {
			log.Println("E", "Wrong buffer value:", gc.TextureRead[float32](buffer)[elementsCount*elementsCount2-1],
				"instead of:", float32(elementsCount*elementsCount2-1))
			break
		}

		buffer2.Load2DFloat32(out)
		//Change current program to selected
		compute.UseProgram(program)
		//Bind buffer to layout binding
		buffer.SetBinding(0)
		buffer2.SetBinding(1)
		//Run program with size
		compute.Realize(buffer2.SizeX, buffer2.SizeY, 1)
		log.Println(j)
	}
	buffer.Close()
	buffer2.Close()
}

// Examples and testing for package functions
func TestComputing(t *testing.T) {
	compute, _ := gc.NewComputing()

	// Create context
	err := glfw.Init()
	if err != nil {
		log.Println("E", "failed to initialize glfw:", err)
		return
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
	glfw.WindowHint(glfw.Visible, glfw.False)
	window, err := glfw.CreateWindow(1, 1, "Computing", nil, nil)
	if err != nil {
		glfw.Terminate()
		log.Println("E", "failed to create window:", err)
	}
	window.MakeContextCurrent()
	err = gl.Init()
	if err != nil {
		log.Println("E", err)
		return
	}

	//Add include loader firstly for include and functions examples
	compute.SetIncludeLoader(func(includeName string) string {
		includeName = strings.Replace(includeName, "\"", "", -1)
		data, err := includes.ReadFile("resources/include/" + includeName)
		if err != nil {
			println("include:", includeName, "not found")
			return ""
		}
		return string(data)
	})

	//Precompiled programs
	bufferProgram := logLoad(compute, bufferTest)
	bufferProgram2 := logLoad(compute, bufferTest2)
	bufferProgram3 := logLoad(compute, bufferTest3)
	textureProgram := logLoad(compute, textureTest)
	textureProgram2 := logLoad(compute, textureTest2)
	functionsProgram := logLoad(compute, functionsTest)
	//speedProgram2 := logLoad(compute, speedTest2)
	//buffer usage examples
	BufferExample(compute, bufferProgram)
	BufferExample2(compute, bufferProgram2)
	BufferExample3(compute, bufferProgram3)
	//Texture usage examples
	TextureExample(compute, textureProgram)
	TextureExample2(compute, textureProgram2)
	//Include and functions examples
	FunctionsExample(compute, functionsProgram)

	//SpeedTest(compute, bufferProgram)
	//SpeedTest2(compute, speedProgram2)
	//Debugger examples
	//debugger1 := gc.CreateDebugger()
	//debugger1.StartWindow()
	//for {
	//}
}
