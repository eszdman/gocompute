layout(rgba32f, binding = 0) uniform image1D img_output;
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
	int idx = int(gl_GlobalInvocationID.x);
	vec4 px = vec4(float(idx)/10.0f);
	vec4 loaded = imageLoad(img_output,idx);
	imageStore(img_output, idx, px+loaded);
}