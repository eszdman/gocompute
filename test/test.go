package main

import (
	_ "embed"
	"gocompute"
	"log"
)

//go:embed resources/test.glsl
var testProg string

func main() {
	compute, _ := gocompute.NewComputing(true)
	program, err := compute.LoadProgram(testProg)
	if err != nil {
		log.Println(err)
		return
	}
	compute.UseProgram(program)
	buffer := compute.NewBuffer()
	buffer2 := compute.NewBuffer()
	buffer2.AllocateInt32(9)
	buffer.LoadDataInt32([]int32{1, 2, 3, 4, 5, 6, 7, 8, 9})
	buffer.BindBase(1)
	buffer2.BindBase(2)
	compute.Realize(9, 1, 1)
	log.Println(buffer.ReadInt32(9))
	log.Println(buffer2.ReadInt32(9))

}
