package gocompute

import (
	"bufio"
	"errors"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"strconv"
	"strings"
)

type computeGroup struct {
	X, Y, Z int
}

type Computing struct {
	includeLoader   func(name string) string
	version         string
	programCounter  int
	programs        map[int]uint32
	maxComputeGroup computeGroup
	computeGroups   map[int]*computeGroup
	defineMap       map[string]string
}

func checkErr(operation string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		msg := operation + ": glError: " + strconv.Itoa(int(err))
		log.Println(msg)
	}
}
func NewComputing(createContext bool) (*Computing, error) {
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
	if createContext {
		err := glfw.Init()
		if err != nil {
			return nil, errors.New("failed to initialize glfw: " + err.Error())
		}
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
		glfw.WindowHint(glfw.Maximized, glfw.True)
		glfw.WindowHint(glfw.Visible, glfw.False)
		window, err := glfw.CreateWindow(1, 1, "Computing", nil, nil)
		if err != nil {
			glfw.Terminate()
			return nil, errors.New("failed to create window: " + err.Error())
		}
		window.MakeContextCurrent()
		err = gl.Init()
		if err != nil {
			return nil, err
		}
	}
	return compute, nil
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
	gl.UseProgram(c.programs[programNumber])
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
			text = "" + text
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
	println("lines:" + lines)
	return lines
}

func (c *Computing) Close() {

}
