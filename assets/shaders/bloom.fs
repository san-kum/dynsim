#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec2 size; // render size

out vec4 finalColor;

// Gaussian Blur weights
float weights[5] = float[](0.227027, 0.1945946, 0.1216216, 0.054054, 0.016216);

void main()
{
    vec4 sum = vec4(0.0);
    vec2 tex_offset = 1.0 / size; // gets size of single texel
    
    // Simple single-pass blur (approximate bloom)
    // Horizontal
    for(int i = 1; i < 5; ++i)
    {
        sum += texture(texture0, fragTexCoord + vec2(tex_offset.x * i, 0.0)) * weights[i];
        sum += texture(texture0, fragTexCoord - vec2(tex_offset.x * i, 0.0)) * weights[i];
    }
    // Vertical
    for(int i = 1; i < 5; ++i)
    {
        sum += texture(texture0, fragTexCoord + vec2(0.0, tex_offset.y * i)) * weights[i];
        sum += texture(texture0, fragTexCoord - vec2(0.0, tex_offset.y * i)) * weights[i];
    }
    
    vec4 pixel = texture(texture0, fragTexCoord);
    
    // Add blurred glow to original
    // "Screen" blend mode approx: 1 - (1-a)(1-b)
    // Or just additive
    vec4 glow = sum * 0.05; // Intensity Drastically Reduced
    
    finalColor = pixel + glow;
}
