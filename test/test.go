package main

import (
	"embed"
	_ "embed"
	"gocompute"
	"log"
	"math"
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

//go:embed resources/include/*
var includes embed.FS

// Recommended to use 2 or 4 components for vectors
type pointsVecXY struct {
	vecX, vecY float32
}
type pointsVecXYZW struct {
	vecX, vecY float32
	vecZ, vecW float32
}

func logLoad(compute *gocompute.Computing, program string) int {
	programID, err := compute.LoadProgram(program)
	if err != nil {
		log.Println(err)
		return -1
	}
	return programID
}
func BufferExample(compute *gocompute.Computing, program int) {
	log.Println("BufferExample started")
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	//Allocate buffer data
	buffer2.AllocateFloat32(9)
	//Load data into buffer instead of allocation
	buffer.LoadFloat32([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9})
	//Change current program to selected
	compute.UseProgram(program)
	//Bind buffer to layout binding
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	//Run program with size
	compute.Realize(buffer2.Size, 1, 1)
	log.Println(buffer2.ReadFloat32(buffer2.Size))
	buffer.Close()
	buffer2.Close()
}
func BufferExample2(compute *gocompute.Computing, program int) {
	log.Println("BufferExample2 started")
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	gocompute.BufferAllocate[pointsVecXY](buffer2, 9)
	points := make([]pointsVecXY, 9)
	//Write into first point for example
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	gocompute.BufferLoad(buffer, points)
	compute.UseProgram(program)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println(gocompute.BufferRead[pointsVecXY](buffer, 9))
	log.Println(gocompute.BufferRead[pointsVecXY](buffer2, 9))
	buffer.Close()
	buffer2.Close()
}
func BufferExample3(compute *gocompute.Computing, program int) {
	log.Println("BufferExample3 started")
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	//Generic gocompute method
	//It's possible to pass structures into buffer memory
	gocompute.BufferAllocate[pointsVecXYZW](buffer2, 9)
	points := make([]pointsVecXYZW, 9)
	//Write into first point for example
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	points[0].vecZ = 0.75
	points[0].vecW = 1.0
	gocompute.BufferLoad(buffer, points)
	compute.UseProgram(program)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	//compute.SetFloat32("test", 0.5)
	compute.Realize(buffer2.Size, 1, 1)
	log.Println(gocompute.BufferRead[pointsVecXYZW](buffer, buffer.Size))
	log.Println(gocompute.BufferRead[pointsVecXYZW](buffer2, buffer2.Size))
	buffer.Close()
	buffer2.Close()
}
func TextureExample(compute *gocompute.Computing, program int) {
	log.Println("TextureExample started")
	texture := compute.NewTexture(gocompute.FLOAT32, 4)
	texture.Create1D(2)
	texture.Load1DFloat32([]float32{1, 1, 1, 1, 0, 0, 0, 0})
	compute.UseProgram(program)
	texture.SetBinding(0)
	//Add offset to gl_GlobalInvocationID
	compute.SetOffset(1, 0, 0)
	compute.Realize(2, 1, 1)
	log.Println(texture.ReadFloat32())
	//vec4 in texture(0,0) should be unchanged because of offset
	texture.Close()
}
func TextureExample2(compute *gocompute.Computing, program int) {
	log.Println("TextureExample2 started")
	texture := compute.NewTexture(gocompute.FLOAT32, 4)
	points := make([]pointsVecXYZW, 2)
	//Write into first point for example
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	texture.Create1D(2)
	//It's possible to pass structures into texture memory
	gocompute.TextureLoad1D(texture, points)
	compute.UseProgram(program)
	texture.SetBinding(0)
	compute.Realize(2, 1, 1)
	log.Println(texture.ReadFloat32())
	log.Println(gocompute.TextureRead[pointsVecXYZW](texture))
	texture.Close()
}

func FunctionsExample(compute *gocompute.Computing, program int) {
	log.Println("FunctionsExample started")
	buffer := compute.NewBuffer()
	//10 elements buffer zero initialized
	buffer.LoadFloat32(make([]float32, 16))
	compute.UseProgram(program)
	buffer.SetBinding(0)
	//Compute from 1 to 15
	compute.SetOffset(1, 0, 0)
	compute.Realize(14, 1, 1)
	log.Println(buffer.ReadFloat32(buffer.Size))
	//compare with cpu implementation
	test := make([]float32, 16)
	for i := 1; i < 15; i++ {
		//Convert integer range 0-16 to float 0-1
		idf := float64(i) / 15.0
		//Center range -0.5 - 0.5
		idf -= 0.5
		test[i] = 1.0 / float32(math.Exp(idf*idf))
	}
	//Comparing GPU simple fast approximation of gaussian function (1.0/x*x) with CPU result (1.0/exp(x*x))
	log.Println(test)
	buffer.Close()
}

// Examples and testing for package functions
func main() {
	compute, _ := gocompute.NewComputing(true)
	//Add include loader firstly for include and functions examples
	compute.SetIncludeLoader(func(includeName string) string {
		data, err := includes.ReadFile("resources/include/" + includeName + ".glsl")
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
	//Buffer usage examples
	BufferExample(compute, bufferProgram)
	BufferExample2(compute, bufferProgram2)
	BufferExample3(compute, bufferProgram3)
	//Texture usage examples
	TextureExample(compute, textureProgram)
	TextureExample2(compute, textureProgram2)
	//Include and functions examples
	FunctionsExample(compute, functionsProgram)

	//Debugger examples
	//debugger1 := gocompute.CreateDebugger()
	//debugger1.StartWindow()
	//for {
	//}
}
