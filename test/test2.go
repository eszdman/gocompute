package main

import (
	"embed"
	_ "embed"
	"fmt"
	"gocompute"
	"log"
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

// Examples and testing for package functions

func main() {
	compute, _ := gocompute.NewComputing()
	//Add include loader firstly for include and functions examples
	compute.SetIncludeLoader(func(includeName string) string {
		data, err := includes.ReadFile("resources/include/" + includeName + ".glsl")
		if err != nil {
			println("include:", includeName, "not found")
			return ""
		}
		return string(data)
	})

	fmt.Println(compute)
}
