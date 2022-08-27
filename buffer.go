package gocompute

import "C"
import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"golang.org/x/exp/constraints"
	"reflect"
	"unsafe"
)

type gpuBuffer struct {
	id    uint32
	usage uint32
}

func (c *computing) NewBuffer() gpuBuffer {
	return c.NewBufferV(gl.STATIC_DRAW)
}
func (c *computing) NewBufferV(usage uint32) gpuBuffer {
	buffer := gpuBuffer{}
	buffer.usage = usage
	gl.GenBuffers(1, &buffer.id)
	return buffer
}
func (b gpuBuffer) Bind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, b.id)
}

func (b gpuBuffer) Allocate(size int) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size, C.malloc(C.size_t(size)), b.usage)
	b.UnBind()
}
func (b gpuBuffer) AllocateInt32(size int) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size*4, C.malloc(C.size_t(size)*4), b.usage)
	b.UnBind()
}

func (b gpuBuffer) LoadData(data []byte) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data), unsafe.Pointer(&data[0]), b.usage)
	b.UnBind()
}
func (b gpuBuffer) LoadDataInt32(data []int32) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), b.usage)
	b.UnBind()
}
func (b gpuBuffer) BindBase(number int) {
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
func read[V constraints.Float | constraints.Integer](b gpuBuffer, size int) []V {
	b.Bind()
	var outType V
	buffer := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, 0, size*int(unsafe.Sizeof(outType)), gl.MAP_READ_BIT)
	slice := toSlice[V](buffer, size)
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	b.UnBind()
	return slice
}
func (b gpuBuffer) Read(size int) []byte {
	return read[byte](b, size)
}
func (b gpuBuffer) ReadInt32(size int) []int32 {
	return read[int32](b, size)
}
