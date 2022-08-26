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
	buffer2.Allocate(4)
	buffer.LoadData([]byte{1, 2, 3, 4})
	buffer.BindBase(1)
	buffer2.BindBase(2)
	compute.Realize(1, 1, 1)
	log.Println(buffer.Read(4))
	log.Println(buffer2.Read(4))
}
