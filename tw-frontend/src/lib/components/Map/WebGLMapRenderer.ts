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
    'River': 11,
    'Lake': 12,
    'Wetland': 13,
    'Mountain': 14,
    'Savanna': 15,
};

// Entity type ID mapping for texture encoding (stored in B channel)
const ENTITY_IDS: Record<string, number> = {
    'wall': 1,
    'portal': 2,
    'statue': 3,
    'npc': 4,
    'creature': 5,
    'item': 6,
    'plant': 7,
    'floor': 8,
};

// Vertex shader - camera transform with texture coordinates
const VERTEX_SHADER = `#version 300 es
precision highp float;

in vec4 a_position;
in vec2 a_texCoord;

uniform vec2 u_offset;  // Camera offset in NDC (-1 to 1)
uniform vec2 u_scale;   // Camera scale (1.0 = no zoom)

out vec2 v_texCoord;

void main() {
    // Apply camera transform: scale then offset
    vec2 pos = a_position.xy * u_scale + u_offset;
    gl_Position = vec4(pos, a_position.zw);
    v_texCoord = a_texCoord;
}
`;

// Fragment shader - Satellite-style rendering with procedural noise and hillshading
const FRAGMENT_SHADER = `#version 300 es
precision highp float;

in vec2 v_texCoord;
out vec4 fragColor;

uniform sampler2D u_dataTexture;
uniform sampler2D u_noiseTexture;  // Procedural noise for terrain detail
uniform float u_worldRadius;
uniform vec2 u_playerPos;     // Player position in normalized coords (0-1)
uniform float u_time;          // For animation
uniform float u_isSimulated;   // 1.0 = simulated world, 0.0 = lobby/unsimulated
uniform vec2 u_texScale;       // Texture sampling scale (X, Y). >1.0 = zoomed out
uniform vec2 u_texCenter;      // Center of view in texture coords (0-1)
uniform float u_seaLevel;      // Sea level in meters (for bathymetry)
uniform float u_minElevation;  // Minimum elevation (deepest ocean) in meters

// Satellite Material Palette (Lifeless Protoplanet)
// Water colors (deep to shallow)
const vec3 C_WATER_ABYSSAL = vec3(0.01, 0.03, 0.08);    // Deepest ocean trenches
const vec3 C_WATER_DEEP = vec3(0.02, 0.06, 0.14);       // Deep ocean
const vec3 C_WATER_MID = vec3(0.08, 0.18, 0.35);        // Mid-depth
const vec3 C_WATER_SHALLOW = vec3(0.18, 0.37, 0.67);    // Continental shelf
const vec3 C_WATER_COASTAL = vec3(0.28, 0.52, 0.72);    // Near-shore coastal

// Land materials (low to high)
const vec3 C_SAND_WET = vec3(0.65, 0.58, 0.42);         // Wet coastal sand
const vec3 C_SAND_DRY = vec3(0.82, 0.76, 0.58);         // Dry desert sand
const vec3 C_SEDIMENT = vec3(0.49, 0.42, 0.34);         // Sediment/Alluvial
const vec3 C_CLAY = vec3(0.58, 0.38, 0.28);             // Iron-rich clay/rust
const vec3 C_ROCK_BASALT = vec3(0.22, 0.21, 0.20);      // Dark volcanic basalt
const vec3 C_ROCK_GRANITE = vec3(0.38, 0.36, 0.34);     // Gray granite
const vec3 C_ROCK_VOLCANIC = vec3(0.15, 0.12, 0.10);    // Fresh volcanic rock

// Ice and snow (elevation/temperature based)
const vec3 C_ICE_ROCK = vec3(0.55, 0.58, 0.62);         // Rocky ice mix
const vec3 C_ICE_GLACIER = vec3(0.75, 0.82, 0.88);      // Glacier blue-white
const vec3 C_SNOW = vec3(0.95, 0.95, 0.98);             // Fresh snow

const vec3 COLOR_PLAYER = vec3(1.0, 0.2, 0.2);          // Player marker

// Lobby/unsimulated world colors
const vec3 COLOR_LOBBY = vec3(0.85, 0.82, 0.78);       // Marble-like
const vec3 COLOR_UNSIMULATED = vec3(0.3, 0.3, 0.35);   // Gray fog

// Biome-based flat colors for unsimulated worlds
vec3 getBiomeColor(float biomeId) {
    int id = int(biomeId * 255.0);
    if (id == 0) return vec3(0.1, 0.3, 0.6);      // Ocean - blue
    if (id == 1) return vec3(0.4, 0.6, 0.3);      // Grassland - green
    if (id == 2) return vec3(0.9, 0.8, 0.5);      // Desert - tan
    if (id == 9) return COLOR_LOBBY;               // Lobby - marble
    return COLOR_UNSIMULATED;                      // Default - gray fog
}

// Entity-based colors (encoded in B channel)
vec3 getEntityColor(float entityId) {
    int id = int(entityId * 255.0);
    if (id == 1) return vec3(0.35, 0.35, 0.4);   // Wall
    if (id == 2) return vec3(0.8, 0.2, 0.8);     // Portal
    if (id == 4) return vec3(0.9, 0.8, 0.3);     // NPC
    if (id == 5) return vec3(0.8, 0.4, 0.3);     // Creature
    return vec3(0.0);
}

void main() {
    // Apply texture zoom by scaling coordinates around center
    vec2 zoomedCoord = u_texCenter + (v_texCoord - vec2(0.5)) * u_texScale;
    
    // Only check Y bounds (poles). X wraps seamlessly via gl.REPEAT.
    if (zoomedCoord.y < 0.0 || zoomedCoord.y > 1.0) {
        fragColor = vec4(0.02, 0.02, 0.05, 1.0); // Dark polar edge
        return;
    }
    
    vec4 data = texture(u_dataTexture, zoomedCoord);
    float elevation = data.r;
    
    // === PROCEDURAL NOISE (De-pixelation) ===
    // Sample noise at high frequency to break up blocky grid
    vec2 noiseUV = zoomedCoord * 45.0;
    float noise = texture(u_noiseTexture, noiseUV).r;
    
    // Perturb elevation slightly with noise
    float elevPerturbed = elevation + (noise - 0.5) * 0.02;
    
    // === HILLSHADING (Dynamic Normals) ===
    // Use dFdx/dFdy to compute screen-space derivatives for lighting
    float dx = dFdx(elevPerturbed) * 50.0;  // Exaggeration factor
    float dy = dFdy(elevPerturbed) * 50.0;
    
    // Construct normal from derivatives (pointing up with some slope)
    vec3 normal = normalize(vec3(-dx, -dy, 0.1));
    
    // Sun direction (top-left, elevated)
    vec3 sunDir = normalize(vec3(-0.5, -0.5, 1.0));
    
    // Diffuse lighting
    float light = max(dot(normal, sunDir), 0.3);
    // Add ambient + slight noise variation
    light = light * 0.6 + 0.4 + (noise - 0.5) * 0.15;
    light = clamp(light, 0.4, 1.2);
    
    vec3 color;
    
    if (u_isSimulated > 0.5) {
        // === SATELLITE MATERIALS ===
        float seaLevelNorm = 0.5;  // Sea level is at 0.5 in normalized elevation
        
        if (elevation <= seaLevelNorm) {
            // Water: Multiple depth zones
            float depth = (seaLevelNorm - elevation) * 2.0;  // 0=coast, 1=deepest
            
            if (depth < 0.1) {
                // Coastal zone
                color = mix(C_WATER_COASTAL, C_WATER_SHALLOW, depth * 10.0);
            } else if (depth < 0.3) {
                // Shelf
                color = mix(C_WATER_SHALLOW, C_WATER_MID, (depth - 0.1) * 5.0);
            } else if (depth < 0.7) {
                // Mid-ocean
                color = mix(C_WATER_MID, C_WATER_DEEP, (depth - 0.3) * 2.5);
            } else {
                // Abyssal
                color = mix(C_WATER_DEEP, C_WATER_ABYSSAL, (depth - 0.7) * 3.3);
            }
            
            // Water gets specular highlight
            light = light * 0.7 + 0.3;
            float spec = pow(max(dot(reflect(-sunDir, vec3(0.0, 0.0, 1.0)), vec3(0.0, 0.0, 1.0)), 0.0), 32.0);
            color += vec3(spec * 0.12);
        } else {
            // Land materials based on height
            float t = (elevation - seaLevelNorm) / (1.0 - seaLevelNorm);
            
            if (t < 0.03) {
                // Wet coastal sand
                color = mix(C_SAND_WET, C_SAND_DRY, t * 33.0);
            } else if (t < 0.08) {
                // Dry sand transitioning to sediment
                color = mix(C_SAND_DRY, C_SEDIMENT, (t - 0.03) * 20.0);
            } else if (t < 0.18) {
                // Sediment (lowlands, basins)
                color = mix(C_SEDIMENT, C_CLAY, (t - 0.08) * 10.0);
            } else if (t < 0.35) {
                // Clay to basalt transition
                color = mix(C_CLAY, C_ROCK_BASALT, (t - 0.18) * 5.9);
            } else if (t < 0.55) {
                // Basalt (volcanic highlands)
                color = mix(C_ROCK_BASALT, C_ROCK_GRANITE, (t - 0.35) * 5.0);
            } else if (t < 0.72) {
                // Granite highlands
                color = mix(C_ROCK_GRANITE, C_ICE_ROCK, (t - 0.55) * 5.9);
            } else if (t < 0.85) {
                // Rocky ice (high altitude)
                color = mix(C_ICE_ROCK, C_ICE_GLACIER, (t - 0.72) * 7.7);
            } else {
                // Glacier to snow (peaks)
                color = mix(C_ICE_GLACIER, C_SNOW, (t - 0.85) * 6.7);
            }
        }
        
        // Apply hillshading
        color = color * light;
    } else {
        // Lobby/unsimulated - use flat biome colors
        color = getBiomeColor(data.g);
    }
    
    // Entity overlay - entities in B channel override base color
    float entityId = data.b;
    if (entityId > 0.01) {
        vec3 entityColor = getEntityColor(entityId);
        color = entityColor;
    }
    
    // Player marker
    vec2 playerScreenPos = vec2(0.5) + (u_playerPos - u_texCenter) / u_texScale;
    float markerDist = distance(v_texCoord, playerScreenPos);
    
    float zoomFactor = max(u_texScale.x, u_texScale.y);
    float markerSize = 0.02 / zoomFactor;
    if (markerDist < markerSize) {
        float alpha = smoothstep(markerSize, markerSize * 0.5, markerDist);
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
    is_simulated?: boolean;
    // Simulation summary data
    avg_temperature?: number;
    max_elevation?: number;
    sea_level?: number;
    land_coverage?: number;
    simulated_years?: number;
    seed?: number;
    // Natural Satellites (Phase 4)
    satellites?: Satellite[];
}

// Natural Satellite matching Go astronomy.Satellite struct
export interface Satellite {
    name: string;            // Generated name (e.g., "Luna", "Io", "Europa")
    mass: number;            // Mass relative to Earth Moon (0.0 - 2.0)
    distance: number;        // Orbital distance in km
    radius: number;          // Radius in km
    period: number;          // Orbital period in days
}

export interface WorldMapTile {
    grid_x: number;
    grid_y: number;
    biome: string;
    avg_elevation: number;
}

// MiniMap data format (local area around player)
export interface MiniMapData {
    tiles: MiniMapTile[];
    player_x: number;
    player_y: number;
    player_z?: number;
    grid_size: number;
    is_simulated?: boolean;
}

export interface MiniMapTile {
    x: number;
    y: number;
    biome: string;
    elevation: number;
    is_player?: boolean;
    entities?: MiniMapEntity[];
}

export interface MiniMapEntity {
    type: string;
    name?: string;
}

export class WebGLMapRenderer {
    private canvas: HTMLCanvasElement;
    private gl: WebGL2RenderingContext | null = null;
    private program: WebGLProgram | null = null;
    private dataTexture: WebGLTexture | null = null;
    private noiseTexture: WebGLTexture | null = null;  // Procedural noise for terrain detail

    private gridWidth: number = 128;
    private gridHeight: number = 64;
    private worldWidth: number = 1;
    private worldHeight: number = 1;
    private worldRadius: number = 6371000;
    private elevationMin: number = -6000;
    private elevationMax: number = 8848;
    private seaLevel: number = 0;

    // Player position in normalized coordinates (0-1)
    private playerPosX: number = 0.5;
    private playerPosY: number = 0.5;

    // Whether this is a simulated world (has geology) or lobby/unsimulated
    private isSimulated: boolean = false;

    // View Transform (texture space)
    private texScaleX: number = 1.0;
    private texScaleY: number = 1.0;
    private centerX: number = 0.5;
    private centerY: number = 0.5;

    // Camera state (public-facing)
    private cameraX: number = 0.5;  // Center in texture coords (0-1)
    private cameraY: number = 0.5;
    private zoom: number = 1.0;     // 1.0 = fit to view, <1 = zoomed in, >1 = zoomed out

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

        // Create procedural noise texture for terrain detail
        this.createNoiseTexture();

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
        // REPEAT on X-axis for seamless horizontal (East/West) wrapping
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
        // CLAMP on Y-axis to prevent viewing past poles
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    }

    private createNoiseTexture(): void {
        const gl = this.gl;
        if (!gl) return;

        this.noiseTexture = gl.createTexture();
        gl.bindTexture(gl.TEXTURE_2D, this.noiseTexture);

        // Generate 256x256 random noise texture
        const size = 256;
        const data = new Uint8Array(size * size * 4);
        for (let i = 0; i < data.length; i += 4) {
            const val = Math.floor(Math.random() * 255);
            data[i] = val;     // R
            data[i + 1] = val; // G
            data[i + 2] = val; // B
            data[i + 3] = 255; // A
        }

        gl.texImage2D(
            gl.TEXTURE_2D, 0, gl.RGBA,
            size, size, 0,
            gl.RGBA, gl.UNSIGNED_BYTE, data
        );

        // Use REPEAT for seamless tiling
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

        console.log('[WebGLMapRenderer] Created noise texture for terrain detail');
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

        // Track whether world is simulated for styling
        this.isSimulated = data.is_simulated ?? false;

        // Store sea level for proper elevation normalization
        this.seaLevel = data.sea_level ?? 0;

        this.dirty = true;

        console.log('[WebGLMapRenderer] Data updated:', {
            grid: `${this.gridWidth}x${this.gridHeight}`,
            tiles: data.tiles.length,
            elevRange: { min: minElev, max: maxElev },
            playerPos: { x: this.playerPosX.toFixed(3), y: this.playerPosY.toFixed(3) },
            isSimulated: this.isSimulated
        });
    }

    /**
     * Update renderer with minimap data (local area around player)
     */
    updateMiniMapData(data: MiniMapData): void {
        const gl = this.gl;
        if (!gl || !this.dataTexture) return;

        const gridSize = data.grid_size || 9;
        this.gridWidth = gridSize;
        this.gridHeight = gridSize;

        // Calculate bounds from tiles
        let minX = Infinity, maxX = -Infinity;
        let minY = Infinity, maxY = -Infinity;
        for (const tile of data.tiles) {
            if (tile.x < minX) minX = tile.x;
            if (tile.x > maxX) maxX = tile.x;
            if (tile.y < minY) minY = tile.y;
            if (tile.y > maxY) maxY = tile.y;
        }

        // Create texture data buffer (RGBA)
        const buffer = new Uint8Array(gridSize * gridSize * 4);

        for (const tile of data.tiles) {
            // Map tile coords to grid coords (0 to gridSize-1)
            const gx = tile.x - minX;
            const gy = tile.y - minY;
            if (gx < 0 || gx >= gridSize || gy < 0 || gy >= gridSize) continue;

            const idx = (gy * gridSize + gx) * 4;

            // R: Elevation normalized
            const normElev = this.normalizeElevation(tile.elevation);
            buffer[idx] = Math.round(normElev * 255);

            // G: Biome ID
            const lookupBiome = BIOME_IDS[tile.biome];
            const biomeId: number = lookupBiome !== undefined ? lookupBiome : 8;
            buffer[idx + 1] = biomeId;

            // B: Entity type (first entity on tile)
            let entityId = 0;
            if (tile.entities && tile.entities.length > 0 && tile.entities[0]) {
                const entityType = tile.entities[0].type.toLowerCase();
                entityId = ENTITY_IDS[entityType] ?? 0;
            }
            buffer[idx + 2] = entityId;

            // A: Alpha
            buffer[idx + 3] = 255;
        }

        // Upload texture
        gl.bindTexture(gl.TEXTURE_2D, this.dataTexture);
        gl.texImage2D(
            gl.TEXTURE_2D, 0, gl.RGBA,
            gridSize, gridSize, 0,
            gl.RGBA, gl.UNSIGNED_BYTE, buffer
        );

        // Player is at center of minimap
        this.playerPosX = 0.5;
        this.playerPosY = 0.5;

        // Track whether world is simulated for styling
        this.isSimulated = data.is_simulated ?? false;

        // Calculate zoom level from altitude
        // When flying above 100m, zoom out: 1.0 at 100m, increases for higher alt
        const altitude = data.player_z ?? 0;
        if (altitude > 100) {
            // Zoom out by 0.01 per meter above 100, capped at 5x zoom
            // Zoom out by 0.01 per meter above 100, capped at 5x zoom
            const zoom = Math.min(1.0 + (altitude - 100) * 0.01, 5.0);
            this.texScaleX = zoom;
            this.texScaleY = zoom;
        } else {
            this.texScaleX = 1.0;
            this.texScaleY = 1.0;
        }

        this.centerX = 0.5;
        this.centerY = 0.5;

        this.dirty = true;
    }

    private normalizeElevation(elevation: number): number {
        // Map elevation to 0-1 with 0.5 at actual sea level
        // Below sea level: 0 to 0.5
        // Above sea level: 0.5 to 1.0
        const relativeElevation = elevation - this.seaLevel;

        if (relativeElevation <= 0) {
            // Underwater: map -6000m below sea level to sea level → 0 to 0.5
            const depthBelowSea = Math.max(relativeElevation, -6000);
            return 0.5 + (depthBelowSea / 12000); // -6000 → 0, 0 → 0.5
        } else {
            // Land: map sea level to 8848m above sea level → 0.5 to 1.0
            const heightAboveSea = Math.min(relativeElevation, 8848);
            return 0.5 + (heightAboveSea / 17696); // 0 → 0.5, 8848 → 1.0
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

        // Bind noise texture for procedural detail
        gl.activeTexture(gl.TEXTURE1);
        gl.bindTexture(gl.TEXTURE_2D, this.noiseTexture);
        const noiseUniform = gl.getUniformLocation(this.program, 'u_noiseTexture');
        gl.uniform1i(noiseUniform, 1);

        // Set world radius uniform
        const radiusUniform = gl.getUniformLocation(this.program, 'u_worldRadius');
        gl.uniform1f(radiusUniform, this.worldRadius);

        // Set player position uniform
        const playerPosUniform = gl.getUniformLocation(this.program, 'u_playerPos');
        gl.uniform2f(playerPosUniform, this.playerPosX, this.playerPosY);

        // Set time uniform for animation
        const timeUniform = gl.getUniformLocation(this.program, 'u_time');
        gl.uniform1f(timeUniform, (Date.now() - this.startTime) / 1000.0);

        // Set isSimulated uniform for styling
        const simulatedUniform = gl.getUniformLocation(this.program, 'u_isSimulated');
        gl.uniform1f(simulatedUniform, this.isSimulated ? 1.0 : 0.0);

        // Vertex shader camera uniforms
        // u_offset: camera offset in NDC space
        // u_scale: camera scale factor (for zoom)
        const offsetUniform = gl.getUniformLocation(this.program, 'u_offset');
        const scaleUniform = gl.getUniformLocation(this.program, 'u_scale');
        // For now, keep vertex positions unchanged (we do the transform in texture space)
        gl.uniform2f(offsetUniform, 0.0, 0.0);
        gl.uniform2f(scaleUniform, 1.0, 1.0);

        // Fragment shader texture sampling uniforms
        const texScaleUniform = gl.getUniformLocation(this.program, 'u_texScale');
        gl.uniform2f(texScaleUniform, this.texScaleX, this.texScaleY);

        const texCenterUniform = gl.getUniformLocation(this.program, 'u_texCenter');
        gl.uniform2f(texCenterUniform, this.centerX, this.centerY);

        // Bathymetry uniforms
        const seaLevelUniform = gl.getUniformLocation(this.program, 'u_seaLevel');
        gl.uniform1f(seaLevelUniform, this.seaLevel);

        const minElevUniform = gl.getUniformLocation(this.program, 'u_minElevation');
        gl.uniform1f(minElevUniform, this.elevationMin);

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
        // Re-fit if in auto-fit mode? For now just mark dirty.
        // We might want to call fitToWorld again if window resizes.
        this.dirty = true;
    }

    /**
     * Fits the map to the current canvas dimensions, maintaining aspect ratio.
     * Centers the map.
     */
    fitToWorld(): void {
        const canvasAspect = this.canvas.width / this.canvas.height;
        const worldAspect = this.gridWidth / this.gridHeight;

        console.log(`[WebGLMapRenderer] fitToWorld: canvas=${this.canvas.width}x${this.canvas.height} (${canvasAspect.toFixed(2)}), world=${this.gridWidth}x${this.gridHeight} (${worldAspect.toFixed(2)})`);

        if (worldAspect > canvasAspect) {
            // World is wider than canvas (or canvas is taller than world)
            // Fit Width: Scale X = 1.0 (Texture width matches Canvas width)
            // Height needs to show MORE of the texture vertically to maintain aspect? 
            // OR Height needs to show LESS?
            // If World 2:1, Canvas 1:1.
            // We squash 2 units of world into 1 unit of canvas width? 
            // Wait, Texture space is 0-1 regardless of aspect.
            // If we map 0-1 X to 0-1 Screen X.
            // We map 0-1 Y to 0-1 Screen Y.
            // This stretches 2:1 world to 1:1 screen.

            // To fix: We want Screen Y to cover 0.5 of world height? No.
            // We want Screen Y to cover range [0..1] of Texture Y?
            // If we fit width, we see full width [0..1].
            // To maintain 2:1 aspect on 1:1 screen, we need to see 0.5 height?
            // No, the image should look "letterboxed".
            // So we need to see MORE vertical space (black bars)?
            // We need to sample range [-0.5..1.5] on Y?
            // Range = ScaleY.
            // Aspect = Width / Height.
            // We want (RangeX * WorldWidth) / (RangeY * WorldHeight) = CanvasWidth / CanvasHeight ???

            // Let's deduce:
            // Screen Ratio = CanvasWidth / CanvasHeight
            // World Ratio = GridWidth / GridHeight

            // We want pixels to be square.
            // Pixel Width in Texture Space = ScaleX / CanvasWidth
            // Pixel Height in Texture Space = ScaleY / CanvasHeight
            // We want (Pixel Width * GridWidth) = (Pixel Height * GridHeight)
            // (ScaleX / CanvasWidth) * GridWidth = (ScaleY / CanvasHeight) * GridHeight
            // ScaleY = ScaleX * (GridWidth / GridHeight) * (CanvasHeight / CanvasWidth)
            // ScaleY = ScaleX * (WorldAspect / CanvasAspect)

            // If we Fit Width: ScaleX = 1.0
            // ScaleY = 1.0 * (WorldAspect / CanvasAspect)
            // 2:1 World, 1:1 Canvas => ScaleY = 2.0. Correct (Zoom Out Y, show more vertical space -> black bars).

            this.texScaleX = 1.0;
            this.texScaleY = worldAspect / canvasAspect;
        } else {
            // World is taller (or canvas is wider)
            // Fit Height: ScaleY = 1.0
            // ScaleX = ScaleY * (CanvasAspect / WorldAspect)?
            // Formula above: ScaleX = ScaleY * (CanvasWidth/CanvasHeight) * (GridHeight/GridWidth)
            // ScaleX = ScaleY * (CanvasAspect / WorldAspect)

            this.texScaleY = 1.0;
            this.texScaleX = canvasAspect / worldAspect;
        }

        this.centerX = 0.5;
        this.centerY = 0.5;
        this.zoom = 1.0;
        this.cameraX = 0.5;
        this.cameraY = 0.5;
        this.dirty = true;
    }

    /**
     * Set camera position and zoom level.
     * @param x - Camera center X in texture coords (can exceed 0-1 for wrapping)
     * @param y - Camera center Y in texture coords (0-1)
     * @param zoom - Zoom level (1.0 = fit to world, <1.0 = zoomed in)
     */
    setCamera(x: number, y: number, zoom: number): void {
        // Calculate aspect-ratio-preserving texture scale
        const canvasAspect = this.canvas.width / this.canvas.height;
        const worldAspect = this.gridWidth / this.gridHeight;

        // Base scale for "fit to world" (when zoom = 1.0)
        let baseScaleX: number, baseScaleY: number;
        if (worldAspect > canvasAspect) {
            // World is wider - fit width
            baseScaleX = 1.0;
            baseScaleY = worldAspect / canvasAspect;
        } else {
            // World is taller - fit height
            baseScaleY = 1.0;
            baseScaleX = canvasAspect / worldAspect;
        }

        // Clamp zoom: minimum 0.1, maximum constrained so texScaleX <= 1.0 (no tiling)
        // If zoom > 1/baseScaleX, we'd see >100% of world width, causing visual tiling
        const maxZoom = 1.0 / baseScaleX;
        this.zoom = Math.max(0.1, Math.min(maxZoom, zoom));

        // Apply zoom: zoom < 1 = zoomed in (smaller scale = smaller texture sample area)
        this.texScaleX = baseScaleX * this.zoom;
        this.texScaleY = baseScaleY * this.zoom;

        // X-axis: Apply modulo wrapping for seamless globe panning
        // Keeps internal value normalized to prevent float precision loss
        this.cameraX = ((x % 1.0) + 1.0) % 1.0;

        // Y-axis: Strict clamping to prevent viewing past poles
        const halfViewY = this.texScaleY * 0.5;
        this.cameraY = Math.max(halfViewY, Math.min(1.0 - halfViewY, y));

        // Update center for fragment shader
        this.centerX = this.cameraX;
        this.centerY = this.cameraY;

        this.dirty = true;
    }

    /**
     * Get the current zoom level
     */
    getZoom(): number {
        return this.zoom;
    }

    /**
     * Get the current camera center position
     */
    getCameraPosition(): { x: number; y: number } {
        return { x: this.cameraX, y: this.cameraY };
    }

    /**
     * Get the current texture scale (for entity overlay synchronization)
     */
    getTexScale(): { x: number; y: number } {
        return { x: this.texScaleX, y: this.texScaleY };
    }

    /**
     * Get grid dimensions (for entity overlay coordinate conversion)
     */
    getGridDimensions(): { width: number; height: number } {
        return { width: this.gridWidth, height: this.gridHeight };
    }

    /**
     * Convert screen coordinates to grid index.
     * @param screenX - X position in screen pixels (0 = left edge of canvas)
     * @param screenY - Y position in screen pixels (0 = top edge of canvas)
     * @returns Grid index {gridX, gridY} or null if out of bounds
     */
    getGridIndexFromScreen(screenX: number, screenY: number): { gridX: number; gridY: number } | null {
        // 1. Screen to normalized canvas (0-1)
        const ndcX = screenX / this.canvas.width;
        const ndcY = screenY / this.canvas.height;

        // 2. Apply inverse camera transform (same as fragment shader)
        // zoomedCoord = center + (texCoord - 0.5) * scale
        // So: texCoord = center + (ndc - 0.5) * scale
        const texX = this.centerX + (ndcX - 0.5) * this.texScaleX;
        const texY = this.centerY + (ndcY - 0.5) * this.texScaleY;

        // 3. Y bounds check (poles - no wrapping)
        if (texY < 0 || texY > 1) {
            return null;
        }

        // 4. X wrapping for seamless globe (matches texture REPEAT mode)
        const wrappedTexX = ((texX % 1.0) + 1.0) % 1.0;

        // 5. Convert to grid index
        const gridX = Math.floor(wrappedTexX * this.gridWidth);
        const gridY = Math.floor(texY * this.gridHeight);

        // Clamp to valid range (gridX should always be valid due to wrapping)
        if (gridX < 0 || gridX >= this.gridWidth || gridY < 0 || gridY >= this.gridHeight) {
            return null;
        }

        return { gridX, gridY };
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
