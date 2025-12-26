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

// Fragment shader - elevation-based coloring with biome support and player marker
const FRAGMENT_SHADER = `#version 300 es
precision highp float;

in vec2 v_texCoord;
out vec4 fragColor;

uniform sampler2D u_dataTexture;
uniform float u_worldRadius;
uniform vec2 u_playerPos;     // Player position in normalized coords (0-1)
uniform float u_time;          // For animation
uniform float u_isSimulated;   // 1.0 = simulated world, 0.0 = lobby/unsimulated
uniform vec2 u_texScale;       // Texture sampling scale (X, Y). >1.0 = zoomed out
uniform vec2 u_texCenter;      // Center of view in texture coords (0-1)
uniform float u_seaLevel;      // Sea level in meters (for bathymetry)
uniform float u_minElevation;  // Minimum elevation (deepest ocean) in meters

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

// Lobby/unsimulated world colors
const vec3 COLOR_LOBBY = vec3(0.85, 0.82, 0.78);          // Marble-like
const vec3 COLOR_UNSIMULATED = vec3(0.3, 0.3, 0.35);      // Gray fog

// Dynamic bathymetry: calculate depth relative to sea level
vec3 getBathymetricColor(float depthFactor) {
    // depthFactor: 0.0 = at sea level, 1.0 = deepest ocean
    // Smooth gradient from shallow turquoise to deep navy
    vec3 shallow = vec3(0.0, 0.6, 0.8);   // Turquoise
    vec3 mid = vec3(0.0, 0.3, 0.5);       // Ocean blue
    vec3 deep = vec3(0.0, 0.1, 0.2);      // Deep navy
    
    if (depthFactor < 0.3) {
        return mix(shallow, mid, depthFactor / 0.3);
    }
    return mix(mid, deep, (depthFactor - 0.3) / 0.7);
}

vec3 getElevationColor(float elevation, float rawElevation) {
    float e = elevation;
    
    // Bathymetric (underwater) - use dynamic sea level
    if (e < 0.5) {
        // Calculate actual depth below sea level
        float depthBelowSea = max(u_seaLevel - rawElevation, 0.0);
        float maxDepth = max(u_seaLevel - u_minElevation, 1.0); // Prevent div by zero
        float depthFactor = clamp(depthBelowSea / maxDepth, 0.0, 1.0);
        return getBathymetricColor(depthFactor);
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

// Biome-based flat colors for unsimulated worlds
vec3 getBiomeColor(float biomeId) {
    int id = int(biomeId * 255.0);
    if (id == 0) return vec3(0.1, 0.3, 0.6);      // Ocean - blue
    if (id == 1) return vec3(0.4, 0.6, 0.3);      // Grassland - green
    if (id == 2) return vec3(0.9, 0.8, 0.5);      // Desert - tan
    if (id == 3) return vec3(0.2, 0.5, 0.3);      // Rainforest - dark green
    if (id == 4) return vec3(0.5, 0.6, 0.4);      // Deciduous - olive
    if (id == 5) return vec3(0.3, 0.5, 0.4);      // Taiga - blue-green
    if (id == 6) return vec3(0.8, 0.85, 0.9);     // Tundra - icy white
    if (id == 7) return vec3(0.6, 0.55, 0.5);     // Alpine - gray brown
    if (id == 9) return COLOR_LOBBY;               // Lobby - marble
    if (id == 11) return vec3(0.2, 0.6, 1.0);     // River - bright blue
    if (id == 12) return vec3(0.15, 0.4, 0.7);    // Lake - deep blue
    if (id == 13) return vec3(0.3, 0.55, 0.5);    // Wetland - blue-green
    if (id == 14) return vec3(0.5, 0.45, 0.4);    // Mountain - gray-brown
    if (id == 15) return vec3(0.7, 0.65, 0.35);   // Savanna - golden
    return COLOR_UNSIMULATED;                      // Default - gray fog
}

// Entity-based colors (encoded in B channel)
vec3 getEntityColor(float entityId) {
    int id = int(entityId * 255.0);
    if (id == 1) return vec3(0.35, 0.35, 0.4);   // Wall - dark grey
    if (id == 2) return vec3(0.8, 0.2, 0.8);     // Portal - magenta
    if (id == 3) return vec3(0.7, 0.7, 0.8);     // Statue - stone gray
    if (id == 4) return vec3(0.9, 0.8, 0.3);     // NPC - gold
    if (id == 5) return vec3(0.8, 0.4, 0.3);     // Creature - orange
    if (id == 6) return vec3(0.3, 0.8, 0.9);     // Item - cyan
    if (id == 7) return vec3(0.2, 0.7, 0.3);     // Plant - bright green
    if (id == 8) return vec3(0.85, 0.82, 0.78);  // Floor - light marble
    return vec3(0.0);                             // No entity (0 = transparent)
}

void main() {
    // Apply texture zoom by scaling coordinates around center
    // u_texScale > 1.0 means we sample a larger area (Zoom Out)
    vec2 zoomedCoord = u_texCenter + (v_texCoord - vec2(0.5)) * u_texScale;
    
    // Check if zoomed coordinates are out of bounds
    if (zoomedCoord.x < 0.0 || zoomedCoord.x > 1.0 || 
        zoomedCoord.y < 0.0 || zoomedCoord.y > 1.0) {
        fragColor = vec4(0.05, 0.05, 0.1, 1.0); // Dark edge
        return;
    }
    
    vec4 data = texture(u_dataTexture, zoomedCoord);
    vec3 color;
    
    // Decode raw elevation from normalized value
    // R channel stores: 0.0 = min elev, 0.5 = sea level, 1.0 = max elev
    float rawElevation;
    if (data.r < 0.5) {
        // Below sea level: map 0.0-0.5 to minElevation-seaLevel
        rawElevation = mix(u_minElevation, u_seaLevel, data.r * 2.0);
    } else {
        // Above sea level: map 0.5-1.0 to seaLevel-8848m
        rawElevation = mix(u_seaLevel, 8848.0, (data.r - 0.5) * 2.0);
    }
    
    if (u_isSimulated > 0.5) {
        // Check for water biomes that should override elevation color
        int biomeId = int(data.g * 255.0);
        if (biomeId == 11) {
            // River - bright blue
            color = vec3(0.2, 0.6, 1.0);
        } else if (biomeId == 12) {
            // Lake - deep blue
            color = vec3(0.15, 0.4, 0.7);
        } else if (biomeId == 13) {
            // Wetland - blue-green
            color = vec3(0.3, 0.55, 0.5);
        } else {
            // Standard simulated world - use elevation-based coloring with dynamic bathymetry
            color = getElevationColor(data.r, rawElevation);
        }
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
    
    // Player marker - circle at player position
    // v_texCoord is screen coordinate (0-1), u_playerPos is texture space position
    vec2 playerScreenPos = vec2(0.5) + (u_playerPos - u_texCenter) / u_texScale;
    float markerDist = distance(v_texCoord, playerScreenPos);
    
    float zoomFactor = max(u_texScale.x, u_texScale.y);
    float markerSize = 0.02 / zoomFactor; // Scale marker with zoom (larger when zoomed in)
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
     * @param x - Camera center X in texture coords (0-1)
     * @param y - Camera center Y in texture coords (0-1)
     * @param zoom - Zoom level (1.0 = fit to world, <1.0 = zoomed in, >1.0 = zoomed out)
     */
    setCamera(x: number, y: number, zoom: number): void {
        // Clamp zoom to reasonable bounds
        this.zoom = Math.max(0.1, Math.min(10.0, zoom));

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

        // Apply zoom: zoom < 1 = zoomed in (smaller scale = smaller texture sample area)
        this.texScaleX = baseScaleX * this.zoom;
        this.texScaleY = baseScaleY * this.zoom;

        // Clamp camera position to prevent viewing outside texture bounds
        // When zoomed in, we need to keep the view within [0, 1]
        const halfViewX = this.texScaleX * 0.5;
        const halfViewY = this.texScaleY * 0.5;

        // Clamp center so view stays within texture
        this.cameraX = Math.max(halfViewX, Math.min(1.0 - halfViewX, x));
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

        // 3. Bounds check
        if (texX < 0 || texX > 1 || texY < 0 || texY > 1) {
            return null;
        }

        // 4. Convert to grid index
        const gridX = Math.floor(texX * this.gridWidth);
        const gridY = Math.floor(texY * this.gridHeight);

        // Clamp to valid range
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
