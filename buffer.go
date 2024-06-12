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
	BStaticWrite BufferUsage = gl.STATIC_DRAW
	BStaticRead              = gl.STATIC_READ
	BStaticCopy              = gl.STATIC_COPY

	BDynamicWrite = gl.DYNAMIC_DRAW
	BDynamicRead  = gl.DYNAMIC_READ
	BDynamicCopy  = gl.DYNAMIC_COPY

	BStreamDraw = gl.STREAM_DRAW
	BStreamRead = gl.STREAM_READ
	BStreamCopy = gl.STREAM_COPY
)

type BufferType uint32

const (
	BStorage BufferType = gl.SHADER_STORAGE_BUFFER
	BUniform            = gl.UNIFORM_BUFFER
)

type GpuBuffer struct {
	id    uint32
	usage BufferUsage
	bType uint32
	Size  int
}

func (c *Computing) NewBuffer() *GpuBuffer {
	return c.NewBufferV(BStaticWrite, BStorage)
}
func (c *Computing) NewBufferV(usage BufferUsage, bType BufferType) *GpuBuffer {
	buffer := &GpuBuffer{}
	buffer.usage = usage
	buffer.bType = uint32(bType)
	gl.GenBuffers(1, &buffer.id)
	return buffer
}
func (b *GpuBuffer) Bind() {
	gl.BindBuffer(b.bType, b.id)
}
func (b *GpuBuffer) check() bool {
	if b.id == 0xFFFFFFFF {
		println("buffer object with ID:", b.id, "already closed!")
		return b.id == 0xFFFFFFFF
	}
	return false
}

// BufferAllocate Allocating memory for buffer with (element x size) bytes count
// Warning: for high memory usage per element , use BufferAllocateBytes instead
func BufferAllocate[V any](b *GpuBuffer, size int) int {
	typeSize := tSize[V]()
	return BufferAllocateBytes(b, size, typeSize)
}

func BufferAllocateBytes(b *GpuBuffer, size, typeSize int) int {
	if b.check() {
		return 0
	}
	b.Bind()
	b.Size = size / typeSize
	//gl.BufferStorage()
	gl.BufferData(b.bType, size*typeSize, nil, uint32(b.usage))
	b.UnBind()
	return size * typeSize
}

func BufferLoad[V any](b *GpuBuffer, data []V) int {
	if b.check() {
		return 0
	}
	typeSize := tSizeInst[V](data)
	b.Bind()
	b.Size = len(data)
	gl.BufferData(b.bType, len(data)*typeSize, unsafe.Pointer(&data[0]), uint32(b.usage))
	b.UnBind()
	return len(data) * typeSize
}
func BufferPartialLoad[V any](b *GpuBuffer, data []V, offsetBytes int) int {
	if b.check() {
		return 0
	}
	typeSize := tSizeInst[V](data)
	b.Bind()
	b.Size = len(data)
	gl.BufferSubData(b.bType, offsetBytes, len(data)*typeSize, unsafe.Pointer(&data[0]))
	b.UnBind()
	return len(data) * typeSize
}

func (b *GpuBuffer) SetBinding(number int) {
	b.BindBaseV(number, b.bType)
}
func (b *GpuBuffer) BindBaseV(number int, tType uint32) {
	gl.BindBufferBase(tType, uint32(number), b.id)
}
func (b *GpuBuffer) UnBind() {
	gl.BindBuffer(b.bType, 0)
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
	buffer := gl.MapBufferRange(b.bType, 0, size*typeSize, gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, size)
	gl.UnmapBuffer(b.bType)
	CheckErr("BufferRead")
	//b.UnBind()
	return slice
}
func BufferReadRange[V any](b *GpuBuffer, min, extent int) []V {
	if b.check() {
		return nil
	}
	typeSize := tSize[V]()
	b.Bind()
	buffer := gl.MapBufferRange(b.bType, min, extent*typeSize, gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, extent)
	gl.UnmapBuffer(b.bType)
	CheckErr("BufferReadRange")
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

func (b *GpuBuffer) Load(data []byte) int {
	return BufferLoad(b, data)
}
func (b *GpuBuffer) LoadInt32(data []int32) int {
	return BufferLoad(b, data)
}
func (b *GpuBuffer) LoadFloat32(data []float32) int {
	return BufferLoad(b, data)
}
func (b *GpuBuffer) LoadFloat64(data []float64) int {
	return BufferLoad(b, data)
}

func (b *GpuBuffer) Close() {
	gl.DeleteBuffers(1, &b.id)
	b.id = 0xFFFFFFFF
}
