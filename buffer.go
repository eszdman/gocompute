package gocompute

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"reflect"
	"unsafe"
)

type GpuBuffer struct {
	id    uint32
	usage uint32
	Size  int
}

func (c *Computing) NewBuffer() *GpuBuffer {
	return c.NewBufferV(gl.STATIC_DRAW)
}
func (c *Computing) NewBufferV(usage uint32) *GpuBuffer {
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
		println("Buffer object with ID:", b.id, "already closed!")
		return b.id == 0xFFFFFFFF
	}
	return false
}

func BufferAllocate[V any](b *GpuBuffer, size int) {
	if b.check() {
		return
	}
	typeSize := tSize[V]()
	b.Bind()
	b.Size = size
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size*typeSize, nil, b.usage)
	//b.UnBind()
}
func BufferLoad[V any](b *GpuBuffer, data []V) {
	if b.check() {
		return
	}
	typeSize := tSize[V]()
	b.Bind()
	b.Size = len(data)
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data)*typeSize, unsafe.Pointer(&data[0]), b.usage)
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

func (b *GpuBuffer) LoadData(data []byte) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadDataInt32(data []int32) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadDataFloat32(data []float32) {
	BufferLoad(b, data)
}
func (b *GpuBuffer) LoadDataFloat64(data []float64) {
	BufferLoad(b, data)
}

func (b *GpuBuffer) Close() {
	gl.DeleteBuffers(1, &b.id)
	b.id = 0xFFFFFFFF
}
