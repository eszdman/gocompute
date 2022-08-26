#version 430

uniform float roll;
writeonly uniform image2D img_output;
layout(std430, binding = 1) buffer inputBuffer {
	uint inputValues[];
};
layout(std430, binding = 2) buffer outputBuffer {
	uint outputValues[];
};

layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
	int idx = int(gl_GlobalInvocationID.x);
	outputValues[idx] = inputValues[idx]*2;
}