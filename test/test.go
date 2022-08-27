package main

import (
	"embed"
	_ "embed"
	"gocompute"
	"log"
)

//go:embed resources/test.glsl
var testProg string

//go:embed resources/test2.glsl
var testProg2 string

//go:embed resources/test3.glsl
var testProg3 string

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
func Example1(compute *gocompute.Computing, program int) {
	log.Println("Example1 started")
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	buffer2.AllocateFloat32(9)
	buffer.LoadDataFloat32([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9})
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(buffer2.Size, 1, 1)
	log.Println(buffer.ReadFloat32(buffer.Size))
	log.Println(buffer2.ReadFloat32(buffer2.Size))
}
func Example2(compute *gocompute.Computing, program int) {
	log.Println("Example2 started")
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	gocompute.BufferAllocate[pointsVecXY](buffer2, 9)
	points := make([]pointsVecXY, 9)
	//Write into first point for example
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	gocompute.BufferLoad(buffer, points)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println(gocompute.BufferRead[pointsVecXY](buffer, 9))
	log.Println(gocompute.BufferRead[pointsVecXY](buffer2, 9))
}
func Example3(compute *gocompute.Computing, program int) {
	log.Println("Example3 started")
	compute.UseProgram(program)

	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	gocompute.BufferAllocate[pointsVecXYZW](buffer2, 9)
	points := make([]pointsVecXYZW, 9)
	//Write into first point for example
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	points[0].vecZ = 0.75
	points[0].vecW = 1.0
	gocompute.BufferLoad(buffer, points)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(buffer2.Size, 1, 1)
	log.Println(gocompute.BufferRead[pointsVecXYZW](buffer, buffer.Size))
	log.Println(gocompute.BufferRead[pointsVecXYZW](buffer2, buffer2.Size))
}
func main() {
	compute, _ := gocompute.NewComputing(true)
	compute.SetIncludeLoader(func(includeName string) string {
		data, err := includes.ReadFile(includeName + ".glsl")
		if err != nil {
			return ""
		}
		return string(data)
	})
	//Precompiled programs
	program0 := logLoad(compute, testProg)
	program1 := logLoad(compute, testProg2)
	program2 := logLoad(compute, testProg3)
	//Buffer usage examples
	Example1(compute, program0)
	Example2(compute, program1)
	Example3(compute, program2)

	//Texture usage examples

}
