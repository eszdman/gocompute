precision lowp float;
uniform float roll;
writeonly uniform image2D img_output;
struct points {
	/*float x;
	float y;*/
	vec2 xy;
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
	outputValues[idx].xy.x = idx;
	outputValues[idx].xy.y = idx;
	outputValues[idx].xy += inputValues[idx].xy;
}