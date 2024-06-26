package gocompute

import (
	"github.com/go-gl/gl/all-core/gl"
	"log"
	"runtime"
	"runtime/debug"
	"unsafe"
)

type GpuTexture struct {
	id       uint32
	channels int
	texType  TextureType
	typeSize int
	sampler  uint32
	levels   int32
	level    int32
	SizeX    int
	SizeY    int
	SizeZ    int
	buffer   interface{}
}
type TextureType int

const (
	NONE TextureType = iota
	SIGNED8
	UNSIGNED8
	SIMPLE8
	SIGNED16
	UNSIGNED16
	SIMPLE16
	FLOAT16
	SIGNED32
	UNSIGNED32
	FLOAT32
)

func (c *Computing) NewTexture(texType TextureType, channels int) *GpuTexture {
	t := GpuTexture{}
	t.channels = channels
	t.texType = texType
	t.levels = 1
	t.level = 0
	t.typeSize = 1
	t.SizeX = 1
	t.SizeY = 1
	t.SizeZ = 1
	switch {
	case texType >= SIGNED32:
		t.typeSize = 4
	case texType >= SIGNED16:
		t.typeSize = 2
	}
	gl.GenTextures(1, &t.id)
	return &t
}

func (t *GpuTexture) GetID() uint32 {
	return t.id
}

func (t *GpuTexture) Bind() {
	gl.BindTexture(t.sampler, t.id)
}
func (t *GpuTexture) UnBind() {
	gl.BindTexture(t.sampler, 0)
}

func (t *GpuTexture) SetLevels(levels int) {
	t.levels = int32(levels)
}

func (t *GpuTexture) SetLevel(level int) {
	if int32(level) >= t.levels {
		t.level = int32(level)
	} else {
		println("Ignored: Level greater than texture level count!")
	}
}

func InternalBuffer[V any](t *GpuTexture) {
	sys := debug.SetGCPercent(-1)
	size := tSize[V]()
	t.buffer = make([]V, t.SizeX*t.SizeY*t.SizeZ*t.channels*t.typeSize/size)
	debug.SetGCPercent(sys)
}

func (t *GpuTexture) Create1D(X int) {
	t.sampler = gl.TEXTURE_1D
	t.Bind()
	gl.TexStorage1D(t.sampler, t.levels, t.InternalFormat(), int32(X))
	t.SizeX = X
	t.SizeY = 1
	t.SizeZ = 1
}
func (t *GpuTexture) Create2D(X, Y int) {
	t.sampler = gl.TEXTURE_2D
	t.Bind()
	gl.TexStorage2D(t.sampler, t.levels, t.InternalFormat(), int32(X), int32(Y))
	CheckErr("TexStorage2D")
	t.SizeX = X
	t.SizeY = Y
	t.SizeZ = 1
}
func (t *GpuTexture) Create3D(X, Y, Z int) {
	t.sampler = gl.TEXTURE_3D
	t.Bind()
	gl.TexStorage3D(t.sampler, t.levels, t.InternalFormat(), int32(X), int32(Y), int32(Z))
	t.SizeX = X
	t.SizeY = Y
	t.SizeZ = Z
}

func TextureLoad1DRange[V any](t *GpuTexture, data []V, offset int) {
	if t.check() {
		return
	}
	t.Bind()
	gl.TexSubImage1D(t.sampler, 0, int32(offset), int32(t.SizeX), t.Format(), t.XType(), unsafe.Pointer(&data[0]))
	CheckErr("TextureSubImage1D")
	t.UnBind()
}

func TextureLoad2DRange[V any](t *GpuTexture, data []V, offsetX, offsetY int) {
	if t.check() {
		return
	}
	sys := debug.SetGCPercent(-1)
	t.Bind()
	if t.buffer != nil {
		copy(t.buffer.([]V), data)
		gl.TexSubImage2D(t.sampler, 0, int32(offsetX), int32(offsetY), int32(t.SizeX), int32(t.SizeY), t.Format(), t.XType(), unsafe.Pointer(&t.buffer.([]V)[0]))
	} else {
		gl.TexSubImage2D(t.sampler, 0, int32(offsetX), int32(offsetY), int32(t.SizeX), int32(t.SizeY), t.Format(), t.XType(), unsafe.Pointer(&data[0]))
	}
	CheckErr("TexSubImage2D")
	debug.SetGCPercent(sys)
	//t.UnBind()
}

func TextureLoad3DRange[V any](t *GpuTexture, data []V, offsetX, offsetY, offsetZ int) {
	if t.check() {
		return
	}
	t.Bind()
	t.SizeX = len(data)
	gl.TexSubImage3D(t.sampler, 0, int32(offsetX), int32(offsetY), int32(offsetZ), int32(t.SizeX), int32(t.SizeY), int32(t.SizeZ), t.Format(), t.XType(), unsafe.Pointer(&data[0]))
	//t.UnBind()
}

func TextureLoad3D[V any](t *GpuTexture, data []V) {
	TextureLoad3DRange(t, data, 0, 0, 0)
}

func TextureLoad2D[V any](t *GpuTexture, data []V) {
	TextureLoad2DRange(t, data, 0, 0)
}

func TextureLoad1D[V any](t *GpuTexture, data []V) {
	TextureLoad1DRange(t, data, 0)
}

func TextureRead[V any](t *GpuTexture) []V {
	t.Bind()
	if t.buffer != nil {
		gl.GetTextureSubImage(t.id, t.level, 0, 0, 0, int32(t.SizeX), int32(t.SizeY), int32(t.SizeZ),
			t.Format(), t.XType(), int32(t.SizeX*t.SizeY*t.SizeZ*t.channels*t.typeSize), unsafe.Pointer(&t.buffer.([]V)[0]))
	} else {
		log.Println("E", "Texture internal buffer is nil! Internal buffer is required for reading texture")
	}
	CheckErr("GetTextureSubImage")
	return t.buffer.([]V)
}

func (t *GpuTexture) SetBinding(number int) {
	if t.check() {
		return
	}
	gl.BindImageTexture(uint32(number), t.id, t.level, false, 0, gl.READ_WRITE, t.InternalFormat())
	CheckErr("BindImageTexture")
}

func (t *GpuTexture) Read() []byte {
	return TextureRead[byte](t)
}
func (t *GpuTexture) ReadInt32() []int32 {
	return TextureRead[int32](t)
}
func (t *GpuTexture) ReadFloat32() []float32 {
	return TextureRead[float32](t)
}

func (t *GpuTexture) Type() TextureType {
	return t.texType
}
func (t *GpuTexture) Channels() int {
	return t.channels
}
func (t *GpuTexture) TypeSize() int {
	return t.typeSize
}
func (t *GpuTexture) InternalFormat() uint32 {
	switch t.texType {
	case SIGNED8:
		{
			switch t.channels {
			case 1:
				return gl.R8I
			case 2:
				return gl.RG8I
			case 3:
				return gl.RGB8I
			case 4:
				return gl.RGBA8I
			}
		}
	case UNSIGNED8:
		{
			switch t.channels {
			case 1:
				return gl.R8UI
			case 2:
				return gl.RG8UI
			case 3:
				return gl.RGB8UI
			case 4:
				return gl.RGBA8UI
			}
		}
	case SIMPLE8:
		{
			switch t.channels {
			case 1:
				return gl.R8
			case 2:
				return gl.RG8
			case 3:
				return gl.RGB8
			case 4:
				return gl.RGBA8
			}
		}
	case SIGNED16:
		{
			switch t.channels {
			case 1:
				return gl.R16I
			case 2:
				return gl.RG16I
			case 3:
				return gl.RGB16I
			case 4:
				return gl.RGBA16I
			}
		}
	case UNSIGNED16:
		{
			switch t.channels {
			case 1:
				return gl.R16UI
			case 2:
				return gl.RG16UI
			case 3:
				return gl.RGB16UI
			case 4:
				return gl.RGBA16UI
			}
		}
	case SIMPLE16:
		{
			switch t.channels {
			case 1:
				return gl.R16
			case 2:
				return gl.RG16
			case 3:
				return gl.RGB16
			case 4:
				return gl.RGBA16
			}
		}
	case FLOAT16:
		{
			switch t.channels {
			case 1:
				return gl.R16F
			case 2:
				return gl.RG16F
			case 3:
				return gl.RGB16F
			case 4:
				return gl.RGBA16F
			}
		}
	case SIGNED32:
		{
			switch t.channels {
			case 1:
				return gl.R32I
			case 2:
				return gl.RG32I
			case 3:
				return gl.RGB32I
			case 4:
				return gl.RGBA32I
			}
		}
	case UNSIGNED32:
		{
			switch t.channels {
			case 1:
				return gl.R32UI
			case 2:
				return gl.RG32UI
			case 3:
				return gl.RGB32UI
			case 4:
				return gl.RGBA32UI
			}
		}
	case FLOAT32:
		{
			switch t.channels {
			case 1:
				return gl.R32F
			case 2:
				return gl.RG32F
			case 3:
				return gl.RGB32F
			case 4:
				return gl.RGBA32F
			}
		}
	}
	return gl.RGBA8
}
func (t *GpuTexture) Format() uint32 {
	switch t.texType {
	case SIMPLE8, SIMPLE16, SIGNED8, UNSIGNED8, FLOAT16, FLOAT32:
		{
			switch t.channels {
			case 1:
				return gl.RED
			case 2:
				return gl.RG
			case 3:
				return gl.RGB
			case 4:
				return gl.RGBA
			}
		}
	case SIGNED16, UNSIGNED16, SIGNED32, UNSIGNED32:
		{
			switch t.channels {
			case 1:
				return gl.RED_INTEGER
			case 2:
				return gl.RG_INTEGER
			case 3:
				return gl.RGB_INTEGER
			case 4:
				return gl.RGBA_INTEGER
			}
		}
	}
	return gl.RGBA
}

func (t *GpuTexture) XType() uint32 {
	switch t.texType {
	case FLOAT16, FLOAT32:
		return gl.FLOAT
	case SIMPLE8, UNSIGNED8:
		return gl.UNSIGNED_BYTE
	case SIGNED8:
		return gl.BYTE
	case SIGNED16:
		return gl.UNSIGNED_SHORT
	case UNSIGNED16:
		return gl.SHORT
	case SIGNED32:
		return gl.INT
	case UNSIGNED32:
		return gl.UNSIGNED_INT
	case SIMPLE16:
		return gl.UNSIGNED_SHORT
	}
	return gl.FLOAT
}
func (t *GpuTexture) check() bool {
	if t.id == 0xFFFFFFFF {
		println("Texture object with ID:", t.id, "already closed!")
		return t.id == 0xFFFFFFFF
	}
	return false
}

func (t *GpuTexture) Load1D(data []byte) {
	TextureLoad1D(t, data)
}
func (t *GpuTexture) Load2D(data []byte) {
	TextureLoad2D(t, data)
}
func (t *GpuTexture) Load3D(data []byte) {
	TextureLoad3D(t, data)
}

func (t *GpuTexture) Load1DFloat32(data []float32) {
	TextureLoad1D(t, data)
}
func (t *GpuTexture) Load2DFloat32(data []float32) {
	TextureLoad2D(t, data)
}
func (t *GpuTexture) Load3DFloat32(data []float32) {
	TextureLoad3D(t, data)
}

func (t *GpuTexture) Close() {
	if t.check() {
		return
	}
	runtime.KeepAlive(t.buffer)
	gl.DeleteTextures(1, &t.id)
	t.id = 0xFFFFFFFF
}
