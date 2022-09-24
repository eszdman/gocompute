package gocompute

import (
	"bufio"
	"fmt"
	_ "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	"strconv"
	"strings"
	"unsafe"
)

type computeGroup struct {
	X, Y, Z int
}

type Computing struct {
	instance        vk.Instance
	device          vk.Device
	currentProgram  int
	includeLoader   func(name string) string
	version         string
	programCounter  int
	programs        map[int]vk.ShaderModule
	maxComputeGroup computeGroup
	commandPool     vk.CommandPool
	computeGroups   map[int]*computeGroup
	defineMap       map[string]string
}

func checkErr(err *error) {
	if v := recover(); v != nil {
		*err = fmt.Errorf("%+v", v)
	}
}

func getPhysicalDevices(instance vk.Instance) ([]vk.PhysicalDevice, error) {
	var gpuCount uint32
	err := vk.Error(vk.EnumeratePhysicalDevices(instance, &gpuCount, nil))
	if err != nil {
		err = fmt.Errorf("vkEnumeratePhysicalDevices failed with %s", err)
		return nil, err
	}
	if gpuCount == 0 {
		err = fmt.Errorf("getPhysicalDevice: no GPUs found on the system")
		return nil, err
	}
	gpuList := make([]vk.PhysicalDevice, gpuCount)
	err = vk.Error(vk.EnumeratePhysicalDevices(instance, &gpuCount, gpuList))
	if err != nil {
		err = fmt.Errorf("vkEnumeratePhysicalDevices failed with %s", err)
		return nil, err
	}
	return gpuList, nil
}

func NewComputing() (*Computing, error) {
	compute := &Computing{}
	err := vk.SetDefaultGetInstanceProcAddr()
	if err != nil {
		panic(err)
	}
	if err := vk.Init(); err != nil {
		panic(err)
	}
	compute.version = "#version 430"
	compute.programs = make(map[int]uint32)
	compute.defineMap = make(map[string]string)
	compute.computeGroups = make(map[int]*computeGroup)
	var instance vk.Instance
	vk.CreateInstance(&vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &vk.ApplicationInfo{
			SType:              vk.StructureTypeApplicationInfo,
			ApiVersion:         vk.ApiVersion10,
			ApplicationVersion: 1,
			PApplicationName:   "goCompute\x00",
			PEngineName:        "goComputingVulkan\x00",
		},
		EnabledExtensionCount:   uint32(0),
		PpEnabledExtensionNames: []string{},
		EnabledLayerCount:       uint32(0),
		PpEnabledLayerNames:     []string{},
	}, nil, &instance)
	compute.instance = instance
	err = vk.InitInstance(compute.instance)
	if err != nil {
		panic(err)
	}
	devices, _ := getPhysicalDevices(compute.instance)
	var dev vk.Device
	vk.CreateDevice(devices[0], &vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount: 0,
		PQueueCreateInfos: []vk.DeviceQueueCreateInfo{{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueCount:       1,
			PQueuePriorities: []float32{1.0},
		}},
		EnabledExtensionCount:   0,
		PpEnabledExtensionNames: []string{},
	}, nil, &dev)
	compute.device = dev
	var pool vk.CommandPool
	vk.CreateCommandPool(compute.device, &vk.CommandPoolCreateInfo{
		SType:            0,
		PNext:            nil,
		Flags:            0,
		QueueFamilyIndex: 0,
	}, nil, &pool)
	compute.commandPool = pool

	vk.GetPhysicalDeviceQueueFamilyProperties()

	//Disable include loader by default
	compute.includeLoader = func(dummy string) string {
		return ""
	}
	return compute, nil
}

func (c *Computing) SetIncludeLoader(loader func(name string) string) {
	if loader != nil {
		c.includeLoader = loader
	}
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func sliceUint32(data []byte) []uint32 {
	const m = 0x7fffffff
	return (*[m / 4]uint32)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&data)).Data))[:len(data)/4]
}

func (c *Computing) UseLoadProgram(programText string) {
	c.defineMap = make(map[string]string)
	programText = c.preProcess(programText)
	var module vk.ShaderModule
	vk.CreateShaderModule(c.device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len([]byte(programText))),
		PCode:    sliceUint32([]byte(programText)),
	}, nil, &module)
}

func (c *Computing) LoadProgram(programText string) (int, error) {
	count := c.programCounter
	c.computeGroups[count] = &computeGroup{1, 1, 1}
	programText = c.preProcess(programText)
	programText = c.preProcess(programText)
	var module vk.ShaderModule
	vk.CreateShaderModule(c.device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len([]byte(programText))),
		PCode:    sliceUint32([]byte(programText)),
	}, nil, &module)

	c.programs[count] = module
	c.programCounter++
	return count, nil
}
func (c *Computing) Realize(x, y, z int) {
	var com []vk.CommandBuffer
	vk.AllocateCommandBuffers(c.device, &vk.CommandBufferAllocateInfo{
		SType:              0,
		PNext:              nil,
		CommandPool:        nil,
		Level:              0,
		CommandBufferCount: 0,
	}, com)
	vk.CmdDispatch(com[0], uint32(x), uint32(y), uint32(z))

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
