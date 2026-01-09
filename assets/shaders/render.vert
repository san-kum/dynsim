#version 330

layout(location = 0) in vec4 particlePos; // From SSBO Offset 0
layout(location = 1) in vec4 particleVel; // From SSBO Offset 16

uniform mat4 mvp;

out vec4 vColor;

void main() {
    gl_Position = mvp * vec4(particlePos.xyz, 1.0);
    
    float speed = length(particleVel.xyz);
    // Size based on speed (Energy)
    gl_PointSize = 1.5 + clamp(speed * 0.1, 0.0, 3.0);
    
    // Aesthetic Color Ramp: Deep Blue -> Cyan -> White -> Gold
    vec3 colLow = vec3(0.1, 0.2, 0.8);  // Deep Blue
    vec3 colMid = vec3(0.0, 0.9, 1.0);  // Cyan
    vec3 colHi  = vec3(1.0, 0.9, 0.5);  // Gold
    
    vec3 finalColor;
    if (speed < 5.0) {
        finalColor = mix(colLow, colMid, speed / 5.0);
    } else {
        finalColor = mix(colMid, colHi, clamp((speed - 5.0) / 10.0, 0.0, 1.0));
    }
    
    // Alpha fade for massive counts
    vColor = vec4(finalColor, 0.6);
}
