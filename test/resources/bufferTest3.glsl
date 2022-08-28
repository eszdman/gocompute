#include sigmoid
precision lowp float;
uniform float roll;
writeonly uniform image2D img_output;
struct points {
	vec4 xyzw;
};
layout(std430, binding = 1) buffer inputBuffer {
	points inputValues[];
};
layout(std430, binding = 2) buffer outputBuffer {
	points outputValues[];
};
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
	int idx = int(gl_GlobalInvocationID.x);
	outputValues[idx].xyzw.x = idx;
	outputValues[idx].xyzw.y = idx;
	outputValues[idx].xyzw.z = idx;
	outputValues[idx].xyzw.w = idx;
	outputValues[idx].xyzw += inputValues[idx].xyzw;
}