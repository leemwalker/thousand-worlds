export interface Position {
    x: number;
    y: number;
}

export interface VisibleTile {
    x: number;
    y: number;
    biome: string;
    elevation: number;
    entities: MapEntity[];
    is_player?: boolean;
    portal?: PortalInfo;
    occluded?: boolean;
}

export interface MapEntity {
    id: string;
    type: string;
    name?: string;
    status?: string;
    glyph?: string;
}

export interface PortalInfo {
    world_name?: string;
    active: boolean;
}

export type RenderQuality = 'low' | 'medium' | 'high';

// RGB Struct for optimization
interface RGB { r: number; g: number; b: number; }

// Helper to parse hex once
const hexToRgb = (hex: string): RGB => {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : { r: 0, g: 0, b: 0 };
};

// Pre-defined RGB colors to avoid regex in hot paths
const COLORS_RGB: Record<string, RGB> = {
    // Biomes
    Ocean: hexToRgb('#4682B4'),
    Rainforest: hexToRgb('#228B22'),
    DeciduousForest: hexToRgb('#3C783C'),
    Taiga: hexToRgb('#32503C'),
    Grassland: hexToRgb('#90EE90'),
    Desert: hexToRgb('#F4A460'),
    Tundra: hexToRgb('#E0FFFF'),
    Alpine: hexToRgb('#808080'),
    // Special
    Lobby: hexToRgb('#505064'),
    Void: hexToRgb('#14141E'),
    Default: hexToRgb('#333333')
};

// Gradient stops for hypsometric tinting (Deepest to Highest)
const ELEVATION_STOPS = [
    { el: -4000, color: hexToRgb('#0a1929') }, // Abyssal
    { el: -2000, color: hexToRgb('#0d3a5c') },
    { el: -200, color: hexToRgb('#1565c0') },
    { el: 0, color: hexToRgb('#4fc3f7') }, // Shallow
    { el: 200, color: hexToRgb('#66bb6a') }, // Lowland
    { el: 500, color: hexToRgb('#aed581') },
    { el: 1500, color: hexToRgb('#dce775') },
    { el: 3000, color: hexToRgb('#a1887f') }, // Highland
    { el: 5000, color: hexToRgb('#9e9e9e') }, // Mountain
    { el: 8000, color: hexToRgb('#fafafa') }  // Peak
];

// Symbols remain unchanged
const ASCII_SYMBOLS: Record<string, string> = {
    player: '@', npc: 'N', monster: 'M', resource: '*', portal: 'O',
    Ocean: '~', Rainforest: 'T', DeciduousForest: 't', Taiga: '^',
    Grassland: '.', Desert: ':', Tundra: '_', Alpine: 'A',
    lobby: '▪', void: '#', default: '?'
};

const EMOJI_SYMBOLS: Record<string, string> = {
    player: '🧍', npc: '👤', monster: '⚔️', resource: '💎', portal: '🌀',
    Ocean: '🌊', Rainforest: '🌳', DeciduousForest: '🌲', Taiga: '🌲',
    Grassland: '🌿', Desert: '🏜️', Tundra: '❄️', Alpine: '🗻',
    lobby: '◻', void: '⬛', default: '·'
};

export class MapRenderer {
    private canvas: HTMLCanvasElement;
    private ctx: CanvasRenderingContext2D;
    private tileSize: number = 16;
    private quality: RenderQuality = 'low';

    // Internal state
    private running: boolean = false;
    private dirty: boolean = false;
    private frameId: number | null = null;

    // Data buffer
    private playerPos: Position = { x: 0, y: 0 };
    private visibleTiles: VisibleTile[] = [];
    private scale: number = 1;

    constructor(canvas: HTMLCanvasElement) {
        this.canvas = canvas;
        const context = canvas.getContext('2d', { alpha: false }); // Optimize for no transparency on canvas
        if (!context) throw new Error('Could not get 2D context');
        this.ctx = context;

        // Accessibility
        this.canvas.setAttribute('role', 'img');
        this.canvas.setAttribute('aria-label', 'Game Map');
    }

    start() {
        if (this.running) return;
        this.running = true;
        this.loop();
    }

    stop() {
        this.running = false;
        if (this.frameId) {
            cancelAnimationFrame(this.frameId);
            this.frameId = null;
        }
    }

    setTileSize(size: number) {
        if (this.tileSize !== size) {
            this.tileSize = size;
            this.dirty = true;
        }
    }

    setQuality(quality: RenderQuality) {
        if (this.quality !== quality) {
            this.quality = quality;
            this.dirty = true;
        }
    }

    updateData(playerPos: Position, tiles: VisibleTile[], scale: number = 1) {
        this.playerPos = playerPos;
        this.visibleTiles = tiles;
        this.scale = scale;
        this.dirty = true;
    }

    // Main render loop
    private loop = () => {
        if (!this.running) return;

        if (this.dirty) {
            this.render();
            this.dirty = false;
        }

        this.frameId = requestAnimationFrame(this.loop);
    };

    private render() {
        const width = this.canvas.width;
        const height = this.canvas.height;
        const centerX = width / 2;
        const centerY = height / 2;

        // Clear canvas
        this.ctx.fillStyle = '#1a1a2e';
        this.ctx.fillRect(0, 0, width, height);

        const effectiveScale = this.scale > 0 ? this.scale : 1;

        // Render tiles
        for (const tile of this.visibleTiles) {
            const screenX = centerX + ((tile.x - this.playerPos.x) / effectiveScale) * this.tileSize;
            const screenY = centerY - ((tile.y - this.playerPos.y) / effectiveScale) * this.tileSize;

            // Frustum cull (basic) - skip if outside canvas
            // Add margin for tile size
            const margin = this.tileSize;
            if (screenX < -margin || screenX > width + margin ||
                screenY < -margin || screenY > height + margin) {
                continue;
            }

            this.renderTile(tile, screenX, screenY);
        }

        this.renderPlayer(centerX, centerY);
    }

    private renderTile(tile: VisibleTile, x: number, y: number) {
        if (this.quality === 'low') {
            this.renderAsciiTile(tile, x, y);
        } else if (this.quality === 'medium') {
            this.renderIconTile(tile, x, y);
        } else {
            this.renderEmojiTile(tile, x, y);
        }
    }

    private renderAsciiTile(tile: VisibleTile, x: number, y: number) {
        this.ctx.fillStyle = this.getBiomeColorString(tile.biome, 0.3);
        this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);

        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
        this.ctx.strokeRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);

        if (tile.occluded) {
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';
            this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);
            return;
        }

        let symbol: string = ASCII_SYMBOLS[tile.biome] || ASCII_SYMBOLS.default || '?';
        let color = '#888888';

        if (tile.portal?.active) { symbol = ASCII_SYMBOLS.portal || '?'; color = '#FF00FF'; }
        if (tile.entities?.[0]) {
            symbol = ASCII_SYMBOLS[tile.entities[0].type] || 'E';
            color = this.getEntityColor(tile.entities[0].type);
        }

        this.ctx.fillStyle = color;
        this.ctx.font = `bold ${this.tileSize * 0.7}px monospace`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(symbol, x, y);
    }

    private renderIconTile(tile: VisibleTile, x: number, y: number) {
        this.ctx.fillStyle = this.getBiomeColorString(tile.biome, 0.6);
        this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);

        // Elevation shading
        const elevationAlpha = Math.max(0, Math.min(0.2, tile.elevation / 5000));
        this.ctx.fillStyle = `rgba(255, 255, 255, ${elevationAlpha})`;
        this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);

        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
        this.ctx.strokeRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);

        if (tile.occluded) {
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';
            this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);
            return;
        }

        if (tile.portal?.active) {
            this.ctx.fillStyle = '#FF00FF';
            this.ctx.beginPath();
            this.ctx.arc(x, y, this.tileSize / 3, 0, Math.PI * 2);
            this.ctx.fill();
        }

        if (tile.entities?.[0]) {
            this.ctx.fillStyle = this.getEntityColor(tile.entities[0].type);
            this.ctx.beginPath();
            this.ctx.arc(x, y, this.tileSize / 4, 0, Math.PI * 2);
            this.ctx.fill();
            this.ctx.strokeStyle = '#FFFFFF';
            this.ctx.stroke();
        }
    }

    private renderEmojiTile(tile: VisibleTile, x: number, y: number) {
        const halfTile = this.tileSize / 2;

        // Optimize: skip getHypsometricColor if 0 elevation to save lerp
        const baseColor = (tile.elevation !== 0)
            ? this.getHypsometricColorString(tile.elevation)
            : this.getBiomeColorString(tile.biome);

        this.ctx.fillStyle = baseColor;
        this.ctx.fillRect(x - halfTile, y - halfTile, this.tileSize + 0.5, this.tileSize + 0.5);

        if (tile.occluded) {
            this.ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';
            this.ctx.fillRect(x - halfTile, y - halfTile, this.tileSize + 0.5, this.tileSize + 0.5);
            return;
        }

        if (this.tileSize >= 8) {
            this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
            this.ctx.strokeRect(x - halfTile, y - halfTile, this.tileSize, this.tileSize);
        }

        if (this.tileSize < 3) return;

        // Biome Emoji
        const biomeEmoji = EMOJI_SYMBOLS[tile.biome] || EMOJI_SYMBOLS.default;
        this.ctx.save();
        this.ctx.globalAlpha = 0.5;
        this.ctx.font = `${this.tileSize * 0.56}px Arial`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(biomeEmoji, x, y);
        this.ctx.restore();

        // Entity/Portal
        if (tile.entities?.[0]) {
            const entity = tile.entities[0];
            const emoji = entity.glyph || EMOJI_SYMBOLS[entity.type] || '👤';
            this.ctx.font = `${this.tileSize * 0.35}px Arial`;
            this.ctx.fillText(emoji, x, y);
        } else if (tile.portal?.active) {
            this.ctx.font = `${this.tileSize * 0.6}px Arial`;
            this.ctx.fillText(EMOJI_SYMBOLS.portal, x, y);
        }
    }

    private renderPlayer(x: number, y: number) {
        // Player rendering logic (simplified for brevity, similar to original)
        this.ctx.fillStyle = this.quality === 'high' ? 'rgba(255, 215, 0, 0.3)' : '#FFD700';
        if (this.quality === 'high') {
            this.ctx.fillRect(x - this.tileSize / 2, y - this.tileSize / 2, this.tileSize, this.tileSize);
            this.ctx.fillStyle = '#FFD700'; // Reset for text
        }

        const symbol = this.quality === 'high' ? '🧍' : '@';
        this.ctx.font = this.quality === 'high' ? `${this.tileSize * 0.7}px Arial` : `bold ${this.tileSize * 0.8}px monospace`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(symbol, x, y);

        this.ctx.strokeStyle = '#FFD700';
        this.ctx.lineWidth = 2;
        this.ctx.strokeRect(x - this.tileSize / 2 + 1, y - this.tileSize / 2 + 1, this.tileSize - 2, this.tileSize - 2);
    }

    private getBiomeColorString(biome: string, alpha: number = 1): string {
        let rgb = COLORS_RGB[biome];
        if (!rgb) {
            const key = Object.keys(COLORS_RGB).find(k => k.toLowerCase() === biome.toLowerCase());
            if (key) rgb = COLORS_RGB[key];

            // Final fallback if undefined
            if (!rgb) rgb = COLORS_RGB.Default;
        }

        // Strict fallback if Default is missing
        const finalRgb: RGB = rgb || { r: 51, g: 51, b: 51 };

        // rgb is guaranteed defined here
        return `rgba(${finalRgb.r}, ${finalRgb.g}, ${finalRgb.b}, ${alpha})`;
    }

    private getHypsometricColorString(elevation: number): string {
        // Find segments
        for (let i = 0; i < ELEVATION_STOPS.length - 1; i++) {
            const start = ELEVATION_STOPS[i];
            const end = ELEVATION_STOPS[i + 1];
            if (start && end && elevation >= start.el && elevation <= end.el) {
                const t = (elevation - start.el) / (end.el - start.el);
                return this.lerpRgbToString(start.color, end.color, t);
            }
        }

        // Fallback checks with strict assertion
        const first = ELEVATION_STOPS[0];
        const last = ELEVATION_STOPS[ELEVATION_STOPS.length - 1];

        if (first && elevation < first.el) return this.rgbToString(first.color);
        if (last) return this.rgbToString(last.color);

        const defaultRgb = COLORS_RGB.Default || { r: 51, g: 51, b: 51 };
        return this.rgbToString(defaultRgb);
    }

    private lerpRgbToString(c1: RGB, c2: RGB, t: number): string {
        const r = Math.round(c1.r + (c2.r - c1.r) * t);
        const g = Math.round(c1.g + (c2.g - c1.g) * t);
        const b = Math.round(c1.b + (c2.b - c1.b) * t);
        return `rgb(${r}, ${g}, ${b})`;
    }

    private rgbToString(c: RGB): string {
        return `rgb(${c.r}, ${c.g}, ${c.b})`;
    }

    private getEntityColor(type: string): string {
        const colors: Record<string, string> = {
            'npc': '#FFFF00', 'player': '#00FFFF', 'resource': '#FFA500',
            'monster': '#FF0000', 'item': '#FFA500'
        };
        return colors[type] || '#FFFFFF';
    }
}
