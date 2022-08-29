#include gaussian
layout(std430, binding = 0) buffer ioBuffer {
    float ioValues[];
};
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
void main() {
    int idx = int(gl_GlobalInvocationID.x);
    float idf = float(idx)/15.0 - 0.5;
    ioValues[idx] = pdf(idf);
}