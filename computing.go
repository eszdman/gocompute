package gocompute

import (
	"bufio"
	"errors"
	"github.com/go-gl/gl/all-core/gl"
	"log"
	"strconv"
	"strings"
	"unsafe"
)

type computeGroup struct {
	X, Y, Z int
}

type Computing struct {
	currentProgram  int
	includeLoader   func(name string) string
	version         string
	programCounter  int
	programs        map[int]uint32
	maxComputeGroup computeGroup
	computeGroups   map[int]*computeGroup
	defineMap       map[string]string
}

func CheckErr(operation string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		msg := operation + ": glError: " + strconv.Itoa(int(err))
		log.Println(msg)
	}
}

func NewComputing() (*Computing, error) {
	compute := &Computing{}
	//Disable include loader by default
	compute.includeLoader = func(dummy string) string {
		return ""
	}
	//Minimal opengl compute version
	compute.version = "#version 430"
	compute.programs = make(map[int]uint32)
	compute.defineMap = make(map[string]string)
	compute.computeGroups = make(map[int]*computeGroup)
	return compute, nil
}

func (c *Computing) GetCurrentProgramID() uint32 {
	return c.programs[c.currentProgram]
}

func (c *Computing) GetUniformLocation(name string) int32 {
	return gl.GetUniformLocation(c.programs[c.currentProgram], gl.Str(name+"\x00"))
}

func (c *Computing) SetInt(name string, input ...int) {
	address := c.GetUniformLocation(name)
	if address == -1 {
		println("SetInt uniform:", name, "not found")
	}
	switch len(input) {
	case 1:
		gl.Uniform1i(address, int32(input[0]))
	case 2:
		gl.Uniform2i(address, int32(input[0]), int32(input[1]))
	case 3:
		gl.Uniform3i(address, int32(input[0]), int32(input[1]), int32(input[2]))
	case 4:
		gl.Uniform4i(address, int32(input[0]), int32(input[1]), int32(input[2]), int32(input[2]))
	}
	CheckErr("SetInt:" + name)
}

func (c *Computing) SetFloat32(name string, input ...float32) {
	address := c.GetUniformLocation(name)
	if address == -1 {
		println("SetFloat32 uniform:", name, "not found")
	}
	switch len(input) {
	case 1:
		gl.Uniform1f(address, input[0])
	case 2:
		gl.Uniform2f(address, input[0], input[1])
	case 3:
		gl.Uniform3f(address, input[0], input[1], input[2])
	case 4:
		gl.Uniform4f(address, input[0], input[1], input[2], input[2])
	case 9:
		gl.UniformMatrix3fv(address, 1, false, &input[0])
	case 16:
		gl.UniformMatrix4fv(address, 1, false, &input[0])
	}
}

// SetOffset Offset only applied to gl_GlobalInvocationID
func (c *Computing) SetOffset(x, y, z int) {
	c.SetInt("computeoffset", x, y, z)
}

func compileShader(shaderType int, shaderProgram string) (uint32, error) {
	shaderHandle := gl.CreateShader(uint32(shaderType))
	if shaderHandle != 0 {
		point, free := gl.Strs(shaderProgram + "\x00")
		defer free()
		gl.ShaderSource(shaderHandle, 1, point, nil)
		gl.CompileShader(shaderHandle)
		compileStatus := int32(0)
		gl.GetShaderiv(shaderHandle, gl.COMPILE_STATUS, &compileStatus)
		if compileStatus == gl.FALSE {
			outLen := int32(0)
			var infoLog [1024]byte
			gl.GetShaderInfoLog(shaderHandle, 1024, &outLen, &(infoLog[0]))
			errStr := "Error compiling shader: " + string(infoLog[:])
			gl.DeleteShader(shaderHandle)
			return 0, errors.New(errStr)
		}
	} else {
		return 0, errors.New("error creating shader")
	}
	return shaderHandle, nil
}
func (c *Computing) LoadProgram(programText string) (int, error) {
	count := c.programCounter
	c.computeGroups[count] = &computeGroup{1, 1, 1}
	programText = c.preProcess(programText)
	shaderHandle, err := compileShader(gl.COMPUTE_SHADER, programText)
	if err != nil {
		return 0, err
	}
	program := gl.CreateProgram()
	gl.AttachShader(program, shaderHandle)
	gl.LinkProgram(program)
	c.programs[count] = program
	c.programCounter++
	return count, nil
}
func (c *Computing) Define(Name string, value string) {
	c.defineMap[Name] = value
}

func (c *Computing) DefineInt(Name string, value int) {
	c.Define(Name, strconv.Itoa(value))
}

func (c *Computing) DefineFloat(Name string, value float64) {
	c.Define(Name, strconv.FormatFloat(value, 'e', 8, 32))
}

func (c *Computing) DefineDouble(Name string, value float64) {
	c.Define(Name, strconv.FormatFloat(value, 'e', 8, 64))
}

func (c *Computing) UseProgram(programNumber int) {
	if len(c.defineMap) > 0 {
		log.Println("Warning: using defines with preloaded program")
	}
	c.currentProgram = programNumber
	gl.UseProgram(c.programs[c.currentProgram])
	CheckErr("UseProgram")
}
func (c *Computing) SetIncludeLoader(loader func(name string) string) {
	if loader != nil {
		c.includeLoader = loader
	}
}
func (c *Computing) Realize(x, y, z int) {
	gl.DispatchCompute(uint32(x), uint32(y), uint32(z))
}

func (c *Computing) UseLoadProgram(programText string) {
	c.defineMap = make(map[string]string)
	programText = c.preProcess(programText)
	shaderHandle, _ := compileShader(gl.COMPUTE_SHADER, programText)
	program := gl.CreateProgram()
	gl.AttachShader(program, shaderHandle)
	gl.LinkProgram(program)
}

func (c *Computing) preProcess(computeProgram string) string {
	scanner := bufio.NewScanner(strings.NewReader(computeProgram))
	lines := ""
	versioned := false
	lineCnt := 0
	for scanner.Scan() {
		lineCnt++
		text := scanner.Text()
		switch {
		case strings.Contains(text, "#include"):
			split := strings.Split(text, " ")
			text = c.includeLoader(split[len(split)-1])
		case strings.Contains(text, "#define"):
			split := strings.Split(text, " ")
			res := c.defineMap[split[1]]
			if res != "" {
				text = "#define " + split[1] + res
			}
		case strings.Contains(text, "main()"):
			text = "uniform ivec3 computeoffset;\n#line " + strconv.Itoa(lineCnt-1) + "\n" + text
		case strings.Contains(text, "gl_GlobalInvocationID"):
			text = strings.ReplaceAll(text, "gl_GlobalInvocationID", "(ivec3(gl_GlobalInvocationID) + ivec3(computeoffset))")
		case strings.Contains(text, "layout"):
			input := strings.ReplaceAll(text, " ", "")
			replacer := strings.NewReplacer("layout(", "", ")in", "", ";", "")
			input = replacer.Replace(input)
			inSplit := strings.Split(input, ",")
			for _, str := range inSplit {
				nv := strings.Split(str, "=")
				parsed, err := strconv.ParseInt(nv[len(nv)-1], 10, 64)
				if err == nil {
					switch nv[0] {
					case "local_size_x":
						c.computeGroups[c.programCounter].X = int(parsed)
					case "local_size_y":
						c.computeGroups[c.programCounter].Y = int(parsed)
					case "local_size_z":
						c.computeGroups[c.programCounter].Z = int(parsed)
					}
				}
			}
		case strings.Contains(text, "#version"):
			versioned = true
		}
		lines += text + "\n"
	}
	if !versioned {
		lines = c.version + "\n#line 1\n" + lines
	}
	//println("lines:" + lines)
	return lines
}

func tSize[V any]() int {
	var inType V
	return int(unsafe.Sizeof(inType))
}

func tSizeInst[V any](d []V) int {
	return int(unsafe.Sizeof(d[0]))
}

func (c *Computing) Close() {

}
