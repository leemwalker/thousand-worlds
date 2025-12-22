/**
 * WebGLMapRenderer - GPU-accelerated map rendering using WebGL2
 * 
 * Renders world map data using shaders for smooth terrain coloring.
 * Supports elevation-based hypsometric/bathymetric colors and biome tinting.
 */

// Biome ID mapping for texture encoding
const BIOME_IDS: Record<string, number> = {
    'Ocean': 0,
    'Grassland': 1,
    'Desert': 2,
    'Rainforest': 3,
    'DeciduousForest': 4,
    'Deciduous Forest': 4,
    'Taiga': 5,
    'Tundra': 6,
    'Alpine': 7,
    'Default': 8,
    'Lobby': 9,
    'Void': 10,
};

// Vertex shader - pass-through with texture coordinates
const VERTEX_SHADER = `#version 300 es
precision highp float;

in vec4 a_position;
in vec2 a_texCoord;

out vec2 v_texCoord;

void main() {
    gl_Position = a_position;
    v_texCoord = a_texCoord;
}
`;

// Fragment shader - elevation-based coloring with biome support and player marker
const FRAGMENT_SHADER = `#version 300 es
precision highp float;

in vec2 v_texCoord;
out vec4 fragColor;

uniform sampler2D u_dataTexture;
uniform float u_worldRadius;
uniform vec2 u_playerPos;    // Player position in normalized coords (0-1)
uniform float u_time;         // For animation

// Earth elevation color stops (hypsometric + bathymetric)
const vec3 COLOR_DEEP_OCEAN = vec3(0.02, 0.05, 0.1);      // -6000m
const vec3 COLOR_ABYSSAL = vec3(0.04, 0.1, 0.16);         // -4000m
const vec3 COLOR_SLOPE = vec3(0.05, 0.23, 0.36);          // -2000m
const vec3 COLOR_SHELF = vec3(0.1, 0.46, 0.82);           // -200m
const vec3 COLOR_COAST = vec3(0.31, 0.76, 0.97);          // 0m
const vec3 COLOR_LOWLAND = vec3(0.18, 0.49, 0.2);         // 100m
const vec3 COLOR_PLAIN = vec3(0.4, 0.73, 0.42);           // 200m
const vec3 COLOR_FOOTHILL = vec3(0.77, 0.88, 0.65);       // 500m
const vec3 COLOR_MOUNTAIN_LOW = vec3(0.84, 0.8, 0.78);    // 1000m
const vec3 COLOR_MOUNTAIN_MID = vec3(0.63, 0.53, 0.5);    // 2000m
const vec3 COLOR_MOUNTAIN_HIGH = vec3(0.55, 0.43, 0.39);  // 3000m
const vec3 COLOR_PEAK = vec3(0.62, 0.62, 0.62);           // 5000m
const vec3 COLOR_SUMMIT = vec3(0.98, 0.98, 0.98);         // 8848m
const vec3 COLOR_PLAYER = vec3(1.0, 0.2, 0.2);            // Player marker

vec3 getElevationColor(float elevation) {
    float e = elevation;
    
    // Bathymetric (underwater)
    if (e < 0.5) {
        float depth = (0.5 - e) * 2.0;
        if (depth > 0.83) return mix(COLOR_ABYSSAL, COLOR_DEEP_OCEAN, (depth - 0.83) / 0.17);
        if (depth > 0.67) return mix(COLOR_SLOPE, COLOR_ABYSSAL, (depth - 0.67) / 0.16);
        if (depth > 0.33) return mix(COLOR_SHELF, COLOR_SLOPE, (depth - 0.33) / 0.34);
        return mix(COLOR_COAST, COLOR_SHELF, depth / 0.33);
    }
    
    // Hypsometric (land)
    float height = (e - 0.5) * 2.0;
    if (height < 0.02) return mix(COLOR_COAST, COLOR_LOWLAND, height / 0.02);
    if (height < 0.04) return mix(COLOR_LOWLAND, COLOR_PLAIN, (height - 0.02) / 0.02);
    if (height < 0.11) return mix(COLOR_PLAIN, COLOR_FOOTHILL, (height - 0.04) / 0.07);
    if (height < 0.22) return mix(COLOR_FOOTHILL, COLOR_MOUNTAIN_LOW, (height - 0.11) / 0.11);
    if (height < 0.45) return mix(COLOR_MOUNTAIN_LOW, COLOR_MOUNTAIN_MID, (height - 0.22) / 0.23);
    if (height < 0.68) return mix(COLOR_MOUNTAIN_MID, COLOR_MOUNTAIN_HIGH, (height - 0.45) / 0.23);
    if (height < 0.85) return mix(COLOR_MOUNTAIN_HIGH, COLOR_PEAK, (height - 0.68) / 0.17);
    return mix(COLOR_PEAK, COLOR_SUMMIT, (height - 0.85) / 0.15);
}

void main() {
    vec4 data = texture(u_dataTexture, v_texCoord);
    vec3 color = getElevationColor(data.r);
    
    // Player marker - draw a pulsing circle at player position
    float dist = distance(v_texCoord, u_playerPos);
    float markerSize = 0.015 + 0.005 * sin(u_time * 3.0); // Pulsing effect
    if (dist < markerSize) {
        float alpha = smoothstep(markerSize, markerSize * 0.5, dist);
        color = mix(color, COLOR_PLAYER, alpha);
    }
    
    fragColor = vec4(color, 1.0);
}
`;

export interface WorldMapData {
    tiles: WorldMapTile[];
    grid_width: number;
    grid_height: number;
    world_width: number;
    world_height: number;
    player_x: number;
    player_y: number;
}

export interface WorldMapTile {
    grid_x: number;
    grid_y: number;
    biome: string;
    avg_elevation: number;
}

export class WebGLMapRenderer {
    private canvas: HTMLCanvasElement;
    private gl: WebGL2RenderingContext | null = null;
    private program: WebGLProgram | null = null;
    private dataTexture: WebGLTexture | null = null;

    private gridWidth: number = 128;
    private gridHeight: number = 64;
    private worldWidth: number = 1;
    private worldHeight: number = 1;
    private worldRadius: number = 6371000;
    private elevationMin: number = -6000;
    private elevationMax: number = 8848;

    // Player position in normalized coordinates (0-1)
    private playerPosX: number = 0.5;
    private playerPosY: number = 0.5;

    private positionBuffer: WebGLBuffer | null = null;
    private texCoordBuffer: WebGLBuffer | null = null;

    private running: boolean = false;
    private dirty: boolean = true;
    private frameId: number | null = null;
    private startTime: number = Date.now();

    constructor(canvas: HTMLCanvasElement) {
        this.canvas = canvas;
        this.init();
    }

    private init(): void {
        // Get WebGL2 context
        const gl = this.canvas.getContext('webgl2', {
            alpha: false,
            antialias: true,
            preserveDrawingBuffer: false,
        });

        if (!gl) {
            console.error('[WebGLMapRenderer] WebGL2 not supported');
            return;
        }

        this.gl = gl;

        // Compile shaders and link program
        this.program = this.createProgram(VERTEX_SHADER, FRAGMENT_SHADER);
        if (!this.program) return;

        // Create geometry (full-screen quad)
        this.createGeometry();

        // Create empty data texture
        this.createDataTexture();

        console.log('[WebGLMapRenderer] Initialized successfully');
    }

    private createProgram(vertexSource: string, fragmentSource: string): WebGLProgram | null {
        const gl = this.gl;
        if (!gl) return null;

        const vertexShader = this.compileShader(gl.VERTEX_SHADER, vertexSource);
        const fragmentShader = this.compileShader(gl.FRAGMENT_SHADER, fragmentSource);

        if (!vertexShader || !fragmentShader) return null;

        const program = gl.createProgram();
        if (!program) return null;

        gl.attachShader(program, vertexShader);
        gl.attachShader(program, fragmentShader);
        gl.linkProgram(program);

        if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
            console.error('[WebGLMapRenderer] Program link failed:', gl.getProgramInfoLog(program));
            return null;
        }

        return program;
    }

    private compileShader(type: number, source: string): WebGLShader | null {
        const gl = this.gl;
        if (!gl) return null;

        const shader = gl.createShader(type);
        if (!shader) return null;

        gl.shaderSource(shader, source);
        gl.compileShader(shader);

        if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
            console.error('[WebGLMapRenderer] Shader compile failed:', gl.getShaderInfoLog(shader));
            gl.deleteShader(shader);
            return null;
        }

        return shader;
    }

    private createGeometry(): void {
        const gl = this.gl;
        if (!gl || !this.program) return;

        // Full-screen quad positions (-1 to 1 in clip space)
        const positions = new Float32Array([
            -1, -1,
            1, -1,
            -1, 1,
            -1, 1,
            1, -1,
            1, 1,
        ]);

        // Texture coordinates (0 to 1, flip Y for correct orientation)
        const texCoords = new Float32Array([
            0, 1,
            1, 1,
            0, 0,
            0, 0,
            1, 1,
            1, 0,
        ]);

        // Position buffer
        this.positionBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, this.positionBuffer);
        gl.bufferData(gl.ARRAY_BUFFER, positions, gl.STATIC_DRAW);

        // Texture coordinate buffer
        this.texCoordBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, this.texCoordBuffer);
        gl.bufferData(gl.ARRAY_BUFFER, texCoords, gl.STATIC_DRAW);
    }

    private createDataTexture(): void {
        const gl = this.gl;
        if (!gl) return;

        this.dataTexture = gl.createTexture();
        gl.bindTexture(gl.TEXTURE_2D, this.dataTexture);

        // Initialize with empty data
        gl.texImage2D(
            gl.TEXTURE_2D, 0, gl.RGBA,
            this.gridWidth, this.gridHeight, 0,
            gl.RGBA, gl.UNSIGNED_BYTE, null
        );

        // Texture parameters for smooth sampling
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    }

    /**
     * Update renderer with world map data
     */
    updateData(data: WorldMapData): void {
        const gl = this.gl;
        if (!gl || !this.dataTexture) return;

        this.gridWidth = data.grid_width;
        this.gridHeight = data.grid_height;

        // Calculate elevation range from data
        let minElev = Infinity, maxElev = -Infinity;
        for (const tile of data.tiles) {
            if (tile.avg_elevation < minElev) minElev = tile.avg_elevation;
            if (tile.avg_elevation > maxElev) maxElev = tile.avg_elevation;
        }
        this.elevationMin = minElev;
        this.elevationMax = maxElev;

        // Create texture data buffer (RGBA)
        const buffer = new Uint8Array(this.gridWidth * this.gridHeight * 4);

        for (const tile of data.tiles) {
            const x = tile.grid_x;
            const y = tile.grid_y;
            const idx = (y * this.gridWidth + x) * 4;

            // R: Elevation normalized to 0-255 (0.5 = sea level)
            // Map elevation range to 0-1 with 0.5 at sea level
            const normElev = this.normalizeElevation(tile.avg_elevation);
            buffer[idx] = Math.round(normElev * 255);

            // G: Biome ID
            const lookupBiome = BIOME_IDS[tile.biome];
            const biomeId: number = lookupBiome !== undefined ? lookupBiome : 8; // 8 = Default
            buffer[idx + 1] = biomeId;

            // B: Unused
            buffer[idx + 2] = 0;

            // A: Alpha
            buffer[idx + 3] = 255;
        }

        // Upload texture
        gl.bindTexture(gl.TEXTURE_2D, this.dataTexture);
        gl.texImage2D(
            gl.TEXTURE_2D, 0, gl.RGBA,
            this.gridWidth, this.gridHeight, 0,
            gl.RGBA, gl.UNSIGNED_BYTE, buffer
        );

        // Store world dimensions and calculate player position in normalized coords
        this.worldWidth = data.world_width || 1;
        this.worldHeight = data.world_height || 1;

        // Normalize player position to 0-1 range
        this.playerPosX = (data.player_x || 0) / this.worldWidth;
        this.playerPosY = (data.player_y || 0) / this.worldHeight;

        this.dirty = true;

        console.log('[WebGLMapRenderer] Data updated:', {
            grid: `${this.gridWidth}x${this.gridHeight}`,
            tiles: data.tiles.length,
            elevRange: { min: minElev, max: maxElev },
            playerPos: { x: this.playerPosX.toFixed(3), y: this.playerPosY.toFixed(3) }
        });
    }

    private normalizeElevation(elevation: number): number {
        // Map elevation to 0-1 with 0.5 at sea level
        // Below sea level: 0 to 0.5
        // Above sea level: 0.5 to 1.0
        if (elevation <= 0) {
            // Underwater: map -6000 to 0 → 0 to 0.5
            const depth = Math.max(elevation, -6000);
            return 0.5 + (depth / 12000); // -6000 → 0, 0 → 0.5
        } else {
            // Land: map 0 to 8848 → 0.5 to 1.0
            const height = Math.min(elevation, 8848);
            return 0.5 + (height / 17696); // 0 → 0.5, 8848 → 1.0
        }
    }

    setWorldRadius(radius: number): void {
        this.worldRadius = radius;
        this.dirty = true;
    }

    start(): void {
        if (this.running) return;
        this.running = true;
        this.loop();
    }

    stop(): void {
        this.running = false;
        if (this.frameId) {
            cancelAnimationFrame(this.frameId);
            this.frameId = null;
        }
    }

    private loop = (): void => {
        if (!this.running) return;

        // Always render for pulsing animation
        this.render();

        this.frameId = requestAnimationFrame(this.loop);
    };

    private render(): void {
        const gl = this.gl;
        if (!gl || !this.program) return;

        // Resize canvas if needed
        this.resizeCanvas();

        gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
        gl.clearColor(0.05, 0.05, 0.1, 1.0);
        gl.clear(gl.COLOR_BUFFER_BIT);

        gl.useProgram(this.program);

        // Bind position attribute
        const posLoc = gl.getAttribLocation(this.program, 'a_position');
        gl.bindBuffer(gl.ARRAY_BUFFER, this.positionBuffer);
        gl.enableVertexAttribArray(posLoc);
        gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);

        // Bind texCoord attribute
        const texLoc = gl.getAttribLocation(this.program, 'a_texCoord');
        gl.bindBuffer(gl.ARRAY_BUFFER, this.texCoordBuffer);
        gl.enableVertexAttribArray(texLoc);
        gl.vertexAttribPointer(texLoc, 2, gl.FLOAT, false, 0, 0);

        // Bind data texture
        gl.activeTexture(gl.TEXTURE0);
        gl.bindTexture(gl.TEXTURE_2D, this.dataTexture);
        const texUniform = gl.getUniformLocation(this.program, 'u_dataTexture');
        gl.uniform1i(texUniform, 0);

        // Set world radius uniform
        const radiusUniform = gl.getUniformLocation(this.program, 'u_worldRadius');
        gl.uniform1f(radiusUniform, this.worldRadius);

        // Set player position uniform
        const playerPosUniform = gl.getUniformLocation(this.program, 'u_playerPos');
        gl.uniform2f(playerPosUniform, this.playerPosX, this.playerPosY);

        // Set time uniform for animation
        const timeUniform = gl.getUniformLocation(this.program, 'u_time');
        gl.uniform1f(timeUniform, (Date.now() - this.startTime) / 1000.0);

        // Draw full-screen quad
        gl.drawArrays(gl.TRIANGLES, 0, 6);
    }

    private resizeCanvas(): void {
        const gl = this.gl;
        if (!gl) return;

        const displayWidth = this.canvas.clientWidth;
        const displayHeight = this.canvas.clientHeight;

        if (this.canvas.width !== displayWidth || this.canvas.height !== displayHeight) {
            this.canvas.width = displayWidth;
            this.canvas.height = displayHeight;
        }
    }

    resize(): void {
        this.dirty = true;
    }

    destroy(): void {
        this.stop();

        const gl = this.gl;
        if (!gl) return;

        if (this.program) gl.deleteProgram(this.program);
        if (this.dataTexture) gl.deleteTexture(this.dataTexture);
        if (this.positionBuffer) gl.deleteBuffer(this.positionBuffer);
        if (this.texCoordBuffer) gl.deleteBuffer(this.texCoordBuffer);

        this.gl = null;
    }
}
