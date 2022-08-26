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

func (c *computing) NewBuffer() *gpuBuffer {
	return c.NewBufferV(gl.STATIC_DRAW)
}
func (c *computing) NewBufferV(usage uint32) *gpuBuffer {
	buffer := &gpuBuffer{}
	buffer.usage = usage
	gl.GenBuffers(1, &buffer.id)
	return buffer
}
func (b *gpuBuffer) Bind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, b.id)
}

func (b *gpuBuffer) Allocate(size int) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, size, C.malloc(C.size_t(size)), b.usage)
	b.UnBind()
}

func (b *gpuBuffer) LoadData(data []byte) {
	b.Bind()
	gl.BufferData(gl.SHADER_STORAGE_BUFFER, len(data), unsafe.Pointer(&data[0]), b.usage)
	b.UnBind()
}
func (b *gpuBuffer) BindBase(number int) {
	b.BindBaseV(number, gl.SHADER_STORAGE_BUFFER)
}
func (b *gpuBuffer) BindBaseV(number int, tType int) {
	gl.BindBufferBase(uint32(tType), uint32(number), b.id)
}
func (b *gpuBuffer) UnBind() {
	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
}

func toSlice(pointer unsafe.Pointer, size int) []byte {
	output := make([]byte, 0)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&output))
	sh.Data = uintptr(pointer)
	sh.Len = size
	sh.Cap = size
	return output
}
func (b *gpuBuffer) Read(size int) []byte {
	b.Bind()
	buffer := gl.MapBufferRange(gl.SHADER_STORAGE_BUFFER, 0, size, gl.MAP_READ_BIT)
	slice := toSlice(buffer, size)
	gl.UnmapBuffer(gl.SHADER_STORAGE_BUFFER)
	b.UnBind()
	return slice
}
