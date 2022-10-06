layout(r32f, binding = 0) uniform image2D img_input;
layout(r32f, binding = 1) uniform image2D img_output;
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
	ivec2 idx = ivec2(gl_GlobalInvocationID.xy);
	imageStore(img_output, idx,  idx.x + imageLoad(img_input,idx));
	//imageStore(img_output, idx,  vec4(0.5));
}