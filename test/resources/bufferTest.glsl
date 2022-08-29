layout(std430, binding = 1) buffer inputBuffer {
	float inputValues[];
};
layout(std430, binding = 2) buffer outputBuffer {
	float outputValues[];
};
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
	int idx = int(gl_GlobalInvocationID.x);
	outputValues[idx] = idx + inputValues[idx];
}