/**
 * MoltenPlanetShader.ts
 * Procedural shader for the pre-simulation planet state (Hadean/Molten Earth).
 * Displays a glowing, animated magma surface with cooling crust patches.
 */

import { Scene } from "@babylonjs/core/scene";
import { ShaderMaterial } from "@babylonjs/core/Materials/shaderMaterial";
import { Effect } from "@babylonjs/core/Materials/effect";
import { Color3 } from "@babylonjs/core/Maths/math.color";

// Register shader code in the Effect store
Effect.ShadersStore["moltenPlanetVertexShader"] = `
    precision highp float;
    
    // Attributes
    attribute vec3 position;
    attribute vec3 normal;
    attribute vec2 uv;
    
    // Uniforms
    uniform mat4 worldViewProjection;
    uniform mat4 world;
    uniform float time;
    
    // Varying
    varying vec2 vUV;
    varying vec3 vPosition;
    varying vec3 vNormal;
    
    void main(void) {
        vec3 positionUpdated = position;
        
        // Slight undulation for "breathing" planet effect
        float undulation = sin(time * 0.5 + position.y * 2.0) * 0.01;
        positionUpdated += normal * undulation;
        
        gl_Position = worldViewProjection * vec4(positionUpdated, 1.0);
        
        vUV = uv;
        vPosition = position;
        vNormal = normalize(vec3(world * vec4(normal, 0.0)));
    }
`;

Effect.ShadersStore["moltenPlanetFragmentShader"] = `
    precision highp float;
    
    // Varying
    varying vec2 vUV;
    varying vec3 vPosition;
    varying vec3 vNormal;
    
    // Uniforms
    uniform float time;
    
    // Noise function
    vec3 mod289(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
    vec4 mod289(vec4 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
    vec4 permute(vec4 x) { return mod289(((x*34.0)+1.0)*x); }
    vec4 taylorInvSqrt(vec4 r) { return 1.79284291400159 - 0.85373472095314 * r; }
    
    float snoise(vec3 v) {
        const vec2  C = vec2(1.0/6.0, 1.0/3.0) ;
        const vec4  D = vec4(0.0, 0.5, 1.0, 2.0);
        
        // First corner
        vec3 i  = floor(v + dot(v, C.yyy) );
        vec3 x0 = v - i + dot(i, C.xxx) ;
        
        // Other corners
        vec3 g = step(x0.yzx, x0.xyz);
        vec3 l = 1.0 - g;
        vec3 i1 = min( g.xyz, l.zxy );
        vec3 i2 = max( g.xyz, l.zxy );
        
        //   x0 = x0 - 0.0 + 0.0 * C.xxx;
        //   x1 = x0 - i1  + 1.0 * C.xxx;
        //   x2 = x0 - i2  + 2.0 * C.xxx;
        //   x3 = x0 - 1.0 + 3.0 * C.xxx;
        vec3 x1 = x0 - i1 + C.xxx;
        vec3 x2 = x0 - i2 + C.yyy; // 2.0*C.x = 1/3 = C.y
        vec3 x3 = x0 - D.yyy;      // -1.0+3.0*C.x = -0.5 = -D.y
        
        // Permutations
        i = mod289(i);
        vec4 p = permute( permute( permute(
                  i.z + vec4(0.0, i1.z, i2.z, 1.0 ))
                + i.y + vec4(0.0, i1.y, i2.y, 1.0 ))
                + i.x + vec4(0.0, i1.x, i2.x, 1.0 ));
                
        // Gradients: 7x7 points over a square, mapped onto an octahedron.
        // The ring size 17*17 = 289 is close to a multiple of 49 (49*6 = 294)
        float n_ = 0.142857142857; // 1.0/7.0
        vec3  ns = n_ * D.wyz - D.xzx;
        
        vec4 j = p - 49.0 * floor(p * ns.z * ns.z);  //  mod(p,7*7)
        
        vec4 x_ = floor(j * ns.z);
        vec4 y_ = floor(j - 7.0 * x_ );    // mod(j,N)
        
        vec4 x = x_ *ns.x + ns.yyyy;
        vec4 y = y_ *ns.x + ns.yyyy;
        vec4 h = 1.0 - abs(x) - abs(y);
        
        vec4 b0 = vec4( x.xy, y.xy );
        vec4 b1 = vec4( x.zw, y.zw );
        
        //vec4 s0 = vec4(lessThan(b0,0.0))*2.0 - 1.0;
        //vec4 s1 = vec4(lessThan(b1,0.0))*2.0 - 1.0;
        vec4 s0 = floor(b0)*2.0 + 1.0;
        vec4 s1 = floor(b1)*2.0 + 1.0;
        vec4 sh = -step(h, vec4(0.0));
        
        vec4 a0 = b0.xczy + s0.xczy*sh.xxyy ;
        vec4 a1 = b1.xczy + s1.xczy*sh.zzww ;
        
        vec3 p0 = vec3(a0.xy,h.x);
        vec3 p1 = vec3(a0.zw,h.y);
        vec3 p2 = vec3(a1.xy,h.z);
        vec3 p3 = vec3(a1.zw,h.w);
        
        //Normalise gradients
        vec4 norm = taylorInvSqrt(vec4(dot(p0,p0), dot(p1,p1), dot(p2, p2), dot(p3,p3)));
        p0 *= norm.x;
        p1 *= norm.y;
        p2 *= norm.z;
        p3 *= norm.w;
        
        // Mix final noise value
        vec4 m = max(0.6 - vec4(dot(x0,x0), dot(x1,x1), dot(x2,x2), dot(x3,x3)), 0.0);
        m = m * m;
        return 42.0 * dot( m*m, vec4( dot(p0,x0), dot(p1,x1),
                                      dot(p2,x2), dot(p3,x3) ) );
    }
    
    // Fractal Brownian Motion
    float fbm(vec3 p) {
        float total = 0.0;
        float amp = 0.5;
        float freq = 1.0;
        for(int i = 0; i < 4; i++) {
            total += snoise(p * freq) * amp;
            freq *= 2.0;
            amp *= 0.5;
        }
        return total;
    }

    void main(void) {
        // Base coordinate with rotation over time
        vec3 coord = vPosition * 2.0;
        coord.x += time * 0.1; 
        
        // Generate noise layers
        float n1 = fbm(coord);
        float n2 = fbm(coord * 2.0 + vec3(time * 0.2));
        float noise = n1 * 0.6 + n2 * 0.4; // Combined noise
        
        // Lava colors
        vec3 darkMagma = vec3(0.1, 0.0, 0.0);
        vec3 redMagma = vec3(0.5, 0.0, 0.0);
        vec3 orangeLava = vec3(1.0, 0.3, 0.0);
        vec3 brightYellow = vec3(1.0, 0.9, 0.2);
        
        // Color mixing based on noise
        vec3 color;
        if (noise < 0.2) {
            color = mix(darkMagma, redMagma, smoothstep(-0.5, 0.2, noise));
        } else if (noise < 0.5) {
            color = mix(redMagma, orangeLava, smoothstep(0.2, 0.5, noise));
        } else {
            color = mix(orangeLava, brightYellow, smoothstep(0.5, 1.0, noise));
        }
        
        // Add "crust" effect
        float crustMap = smoothstep(0.4, 0.45, fbm(vPosition * 4.0 - vec3(time * 0.05)));
        color = mix(color, darkMagma * 0.5, crustMap);
        
        // Glow/Emission
        float brightness = dot(color, vec3(0.299, 0.587, 0.114));
        vec3 emission = color * (brightness * 2.0); // Brighter areas glow more
        
        gl_FragColor = vec4(color + emission * 0.5, 1.0);
    }
`;

export class MoltenPlanetShader {
    private material: ShaderMaterial;
    private scene: Scene;
    private startTime: number;

    constructor(scene: Scene) {
        this.scene = scene;
        this.startTime = Date.now();

        this.material = new ShaderMaterial(
            "moltenPlanet",
            scene,
            {
                vertex: "moltenPlanet",
                fragment: "moltenPlanet",
            },
            {
                attributes: ["position", "normal", "uv"],
                uniforms: ["world", "worldView", "worldViewProjection", "view", "projection", "time"],
                needAlphaBlending: false
            }
        );

        this.material.backFaceCulling = true;

        // Register update loop
        scene.onBeforeRenderObservable.add(this.update);
    }

    private update = () => {
        if (!this.material) return;
        const time = (Date.now() - this.startTime) * 0.001;
        this.material.setFloat("time", time);
    };

    getMaterial(): ShaderMaterial {
        return this.material;
    }

    dispose(): void {
        this.scene.onBeforeRenderObservable.removeCallback(this.update);
        this.material.dispose();
    }
}
