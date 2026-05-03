#version 110

uniform sampler2D tex;
uniform sampler2D kernelTex;

varying vec2 fragTexCoord;
uniform float radius;
uniform vec2 size;
uniform float kernelLen;
uniform vec2 direction;
uniform float sampleScale;

float getKernel(int i) {
    float u = (float(i) + 0.5) / kernelLen;
    return texture2D(kernelTex, vec2(u, 0.5)).r;
}

void main() {
    vec2 inverseSize = vec2(1.0 / size.x, 1.0 / size.y);
    int length = 2 * int(radius) + 1;
    vec4 sum = vec4(0.0);

    for (int i = 0; i < length; ++i) {
        float offset = (float(i) - radius) * sampleScale;
        vec2 tc = fragTexCoord + inverseSize * direction * offset;
        sum += getKernel(i) * texture2D(tex, tc);
    }
    gl_FragColor = sum;
}
