package main

import (
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

// Recommended to use 2 or 4 components for vectors
type pointsXY struct {
	vecX, vecY float32
}
type pointsXYZW struct {
	vecX, vecY float32
	vecZ, vecW float32
}

func logLoad(compute *gocompute.Computing, program string) int {
	programID, err := compute.LoadProgram(program)
	if err != nil {
		log.Println(err)
		return 0
	}
	return programID
}
func Example1(compute *gocompute.Computing, program int) {
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()

	buffer2.AllocateFloat32(9)
	buffer.LoadDataFloat32([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9})
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println(buffer.ReadFloat32(9))
	log.Println(buffer2.ReadFloat32(9))
}
func Example2(compute *gocompute.Computing, program int) {
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	gocompute.BufferAllocate[pointsXY](buffer2, 9)
	points := make([]pointsXY, 9)
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	gocompute.BufferLoad(buffer, points)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println(gocompute.BufferRead[pointsXY](buffer, 9))
	log.Println(gocompute.BufferRead[pointsXY](buffer2, 9))
}
func Example3(compute *gocompute.Computing, program int) {
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	gocompute.BufferAllocate[pointsXYZW](buffer2, 9)
	points := make([]pointsXYZW, 9)
	points[0].vecX = 0.25
	points[0].vecY = 0.5
	points[0].vecZ = 0.75
	points[0].vecW = 1.0
	gocompute.BufferLoad(buffer, points)
	buffer.SetBinding(1)
	buffer2.SetBinding(2)
	compute.Realize(9, 1, 1)
	log.Println(gocompute.BufferRead[pointsXYZW](buffer, 9))
	log.Println(gocompute.BufferRead[pointsXYZW](buffer2, 9))
}
func main() {
	compute, _ := gocompute.NewComputing(true)
	program0 := logLoad(compute, testProg)
	program1 := logLoad(compute, testProg2)
	program2 := logLoad(compute, testProg3)

	Example1(compute, program0)
	Example2(compute, program1)
	Example3(compute, program2)

}
