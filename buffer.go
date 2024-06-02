package gocompute

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/go-gl/gl/all-core/gl"
	"reflect"
	"unsafe"
)

type BufferUsage int

const (
	STATIC_WRITE BufferUsage = gl.STATIC_DRAW
	STATIC_READ              = gl.STATIC_READ
	STATIC_COPY              = gl.STATIC_COPY

	DYNAMIC_WRITE = gl.DYNAMIC_DRAW
	DYNAMIC_READ  = gl.DYNAMIC_READ
	DYNAMIC_COPY  = gl.DYNAMIC_COPY

	STREAM_WRITE = gl.STREAM_DRAW
	STREAM_READ  = gl.STREAM_READ
	STREAM_COPY  = gl.STREAM_COPY
)

type GpuBuffer struct {
	id    uint32
	usage BufferUsage
	Size  int
}

func (c *Computing) NewBuffer() *GpuBuffer {
	return c.NewBufferV(STATIC_WRITE)
}
func (c *Computing) NewBufferV(usage BufferUsage) *GpuBuffer {
	buffer := &GpuBuffer{}
	buffer.usage = usage
	gl.GenBuffers(1, &buffer.id)
	return buffer
}
func (b *GpuBuffer) Bind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, b.id)
}
func (b *GpuBuffer) check() bool {
	if b.id == 0xFFFFFFFF {
		println("buffer object with ID:", b.id, "already closed!")
		return b.id == 0xFFFFFFFF
	}
	return false
}

// BufferAllocate Allocating memory for buffer with (element x size) bytes count
func BufferAllocate[V any](b *GpuBuffer, size int) {
	if b.check() {
		return
	}
	typeSize := tSize[V]()
	b.Bind()
	b.Size = size
	//gl.BufferStorage()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size*typeSize, nil, uint32(b.usage))
	//b.UnBind()
}
func BufferLoad[V any](b *GpuBuffer, data []V) {
	if b.check() {
		return
	}
	typeSize := tSize[V]()
	b.Bind()
	b.Size = len(data)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data)*typeSize, unsafe.Pointer(&data[0]), uint32(b.usage))
	//b.UnBind()
}
func (b *GpuBuffer) SetBinding(number int) {
	b.BindBaseV(number, gl.SHADER_STORAGE_BUFFER)
}
func (b *GpuBuffer) BindBaseV(number int, tType int) {
	gl.BindBufferBase(uint32(tType), uint32(number), b.id)
}
func (b *GpuBuffer) UnBind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

func toSlice[V any](pointer unsafe.Pointer, size int) []V {
	output := make([]V, 0)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&output))
	sh.Data = uintptr(pointer)
	sh.Len = size
	sh.Cap = size
	return output
}
func BufferRead[V any](b *GpuBuffer, size int) []V {
	if b.check() {
		return nil
	}
	typeSize := tSize[V]()
	b.Bind()
	buffer := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, 0, size*typeSize, gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, size)
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	//b.UnBind()
	return slice
}
func BufferReadRange[V any](b *GpuBuffer, min, extent int) []V {
	if b.check() {
		return nil
	}
	typeSize := tSize[V]()
	b.Bind()
	buffer := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, min, extent*typeSize, gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, extent)
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	//b.UnBind()
	return slice
}

func (b *GpuBuffer) Allocate(size int) {
	BufferAllocate[byte](b, size)
}
func (b *GpuBuffer) AllocateInt32(size int) {
	BufferAllocate[int32](b, size)
}
func (b *GpuBuffer) AllocateFloat32(size int) {
	BufferAllocate[float32](b, size)
}
func (b *GpuBuffer) AllocateFloat64(size int) {
	BufferAllocate[float64](b, size)
}

func (b *GpuBuffer) Read(size int) []byte {
	return BufferRead[byte](b, size)
}
func (b *GpuBuffer) ReadInt32(size int) []int32 {
	return BufferRead[int32](b, size)
}
func (b *GpuBuffer) ReadFloat32(size int) []float32 {
	return BufferRead[float32](b, size)
}
func (b *GpuBuffer) ReadFloat64(size int) []float64 {
	return BufferRead[float64](b, size)
}

func (b *GpuBuffer) Load(data []byte) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadInt32(data []int32) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadFloat32(data []float32) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadFloat64(data []float64) {
	BufferLoad(b, data)
}

func (b *GpuBuffer) Close() {
	gl.DeleteBuffers(1, &b.id)
	b.id = 0xFFFFFFFF
}
