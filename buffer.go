package gocompute

import "C"
import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"reflect"
	"unsafe"
)

type gpuBuffer struct {
	id    uint32
	usage uint32
}

func (c *Computing) NewBuffer() gpuBuffer {
	return c.NewBufferV(gl.STATIC_DRAW)
}
func (c *Computing) NewBufferV(usage uint32) gpuBuffer {
	buffer := gpuBuffer{}
	buffer.usage = usage
	gl.GenBuffers(1, &buffer.id)
	return buffer
}
func (b gpuBuffer) Bind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, b.id)
}
func tSize[V any]() int {
	var inType V
	return int(unsafe.Sizeof(inType))
}
func BufferAllocate[V any](b gpuBuffer, size int) {
	typeSize := tSize[V]()
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size*typeSize, C.malloc(C.size_t(size*typeSize)), b.usage)
	b.UnBind()
}
func BufferLoad[V any](b gpuBuffer, data []V) {
	typeSize := tSize[V]()
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data)*typeSize, unsafe.Pointer(&data[0]), b.usage)
	b.UnBind()
}
func (b gpuBuffer) SetBinding(number int) {
	b.BindBaseV(number, gl.SHADER_STORAGE_BUFFER)
}
func (b gpuBuffer) BindBaseV(number int, tType int) {
	gl.BindBufferBase(uint32(tType), uint32(number), b.id)
}
func (b gpuBuffer) UnBind() {
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
func BufferRead[V any](b gpuBuffer, size int) []V {
	typeSize := tSize[V]()
	b.Bind()
	buffer := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, 0, size*typeSize, gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, size)
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	b.UnBind()
	return slice
}

func (b gpuBuffer) Allocate(size int) {
	BufferAllocate[byte](b, size)
}
func (b gpuBuffer) AllocateInt32(size int) {
	BufferAllocate[int32](b, size)
}
func (b gpuBuffer) AllocateFloat32(size int) {
	BufferAllocate[float32](b, size)
}
func (b gpuBuffer) AllocateFloat64(size int) {
	BufferAllocate[float64](b, size)
}

func (b gpuBuffer) Read(size int) []byte {
	return BufferRead[byte](b, size)
}
func (b gpuBuffer) ReadInt32(size int) []int32 {
	return BufferRead[int32](b, size)
}
func (b gpuBuffer) ReadFloat32(size int) []float32 {
	return BufferRead[float32](b, size)
}
func (b gpuBuffer) ReadFloat64(size int) []float64 {
	return BufferRead[float64](b, size)
}

func (b gpuBuffer) LoadData(data []byte) {
	BufferLoad(b, data)
}
func (b gpuBuffer) LoadDataInt32(data []int32) {
	BufferLoad(b, data)
}
func (b gpuBuffer) LoadDataFloat32(data []float32) {
	BufferLoad(b, data)
}
func (b gpuBuffer) LoadDataFloat64(data []float64) {
	BufferLoad(b, data)
}
