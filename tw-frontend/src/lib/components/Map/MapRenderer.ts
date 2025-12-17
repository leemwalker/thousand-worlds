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

// ASCII symbols for low quality rendering
const ASCII_SYMBOLS: Record<string, string> = {
    player: '@',
    npc: 'N',
    monster: 'M',
    resource: '*',
    portal: 'O',
    // Biomes
    Ocean: '~',
    Rainforest: 'T',
    DeciduousForest: 't',
    Taiga: '^',
    Grassland: '.',
    Desert: ':',
    Tundra: '_',
    Alpine: 'A',
    lobby: '▪',
    void: '#',
    default: '?'
};

// Emoji symbols for high quality rendering
const EMOJI_SYMBOLS: Record<string, string> = {
    player: '🧍',
    npc: '👤',
    monster: '⚔️',
    resource: '💎',
    portal: '🌀',
    // Biomes
    Ocean: '🌊',
    Rainforest: '🌳',
    DeciduousForest: '🌲',
    Taiga: '🌲',
    Grassland: '🌿',
    Desert: '🏜️',
    Tundra: '❄️',
    Alpine: '🗻',
    lobby: '◻',
    void: '⬛',
    default: '·'
};

export class MapRenderer {
    private canvas: HTMLCanvasElement;
    private ctx: CanvasRenderingContext2D;
    private tileSize: number = 16;
    private quality: RenderQuality = 'low';

    constructor(canvas: HTMLCanvasElement) {
        this.canvas = canvas;
        const context = canvas.getContext('2d');
        if (!context) {
            throw new Error('Could not get 2D context from canvas');
        }
        this.ctx = context;
    }

    setTileSize(size: number) {
        this.tileSize = size;
    }

    setQuality(quality: RenderQuality) {
        this.quality = quality;
    }

    render(playerPos: Position, visibleArea: VisibleTile[], scale: number = 1) {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;

        // Clear canvas with dark background
        this.ctx.fillStyle = '#1a1a2e';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Render visible tiles
        // Divide by scale to convert world coordinate deltas to grid positions
        const effectiveScale = scale > 0 ? scale : 1;
        for (const tile of visibleArea) {
            const screenX = centerX + ((tile.x - playerPos.x) / effectiveScale) * this.tileSize;
            // Negate Y delta: game uses Y-increases-north, screen uses Y-increases-down
            const screenY = centerY - ((tile.y - playerPos.y) / effectiveScale) * this.tileSize;

            this.renderTile(tile, screenX, screenY);
        }

        // Render player at center (always on top)
        this.renderPlayer(centerX, centerY);
    }

    private renderTile(tile: VisibleTile, x: number, y: number) {
        switch (this.quality) {
            case 'low':
                this.renderAsciiTile(tile, x, y);
                break;
            case 'medium':
                this.renderIconTile(tile, x, y);
                break;
            case 'high':
                this.renderEmojiTile(tile, x, y);
                break;
        }
    }

    private renderAsciiTile(tile: VisibleTile, x: number, y: number) {
        // Background color based on biome (muted)
        this.ctx.fillStyle = this.getBiomeColor(tile.biome, 0.3);
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Grid lines
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
        this.ctx.strokeRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Determine what symbol to show
        let symbol = ASCII_SYMBOLS[tile.biome] || ASCII_SYMBOLS.default;
        let color = '#888888';

        // Portal takes priority
        if (tile.portal?.active) {
            symbol = ASCII_SYMBOLS.portal;
            color = '#FF00FF';
        }

        // Entity takes higher priority
        if (tile.entities && tile.entities.length > 0) {
            const entity = tile.entities[0];
            symbol = ASCII_SYMBOLS[entity.type] || 'E';
            color = this.getEntityColor(entity.type);
        }

        // Draw ASCII character
        this.ctx.fillStyle = color;
        this.ctx.font = `bold ${this.tileSize * 0.7}px monospace`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(symbol, x, y);
    }

    private renderIconTile(tile: VisibleTile, x: number, y: number) {
        // Colored background based on biome
        this.ctx.fillStyle = this.getBiomeColor(tile.biome, 0.6);
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Elevation shading
        const elevationAlpha = Math.max(0, Math.min(0.2, tile.elevation / 5000));
        this.ctx.fillStyle = `rgba(255, 255, 255, ${elevationAlpha})`;
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Grid lines
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
        this.ctx.strokeRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Portal indicator
        if (tile.portal?.active) {
            this.ctx.fillStyle = '#FF00FF';
            this.ctx.beginPath();
            this.ctx.arc(x, y, this.tileSize / 3, 0, Math.PI * 2);
            this.ctx.fill();
        }

        // Entity indicators as colored dots
        if (tile.entities && tile.entities.length > 0) {
            const entity = tile.entities[0];
            this.ctx.fillStyle = this.getEntityColor(entity.type);
            this.ctx.beginPath();
            this.ctx.arc(x, y, this.tileSize / 4, 0, Math.PI * 2);
            this.ctx.fill();
            this.ctx.strokeStyle = '#FFFFFF';
            this.ctx.lineWidth = 1;
            this.ctx.stroke();
        }
    }

    private renderEmojiTile(tile: VisibleTile, x: number, y: number) {
        const halfTile = this.tileSize / 2;

        // Layer 1: Hypsometric elevation color (solid base)
        this.ctx.fillStyle = this.getHypsometricColor(tile.elevation);
        // Add minimal overlap (0.5px) to prevent subpixel rendering gaps (seams)
        this.ctx.fillRect(
            x - halfTile,
            y - halfTile,
            this.tileSize + 0.5,
            this.tileSize + 0.5
        );

        // Subtle grid lines (only if tiles are large enough)
        if (this.tileSize >= 8) {
            this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
            this.ctx.strokeRect(x - halfTile, y - halfTile, this.tileSize, this.tileSize);
        }

        // Skip emoji rendering if tiles are too small (<3px)
        if (this.tileSize < 3) return;

        // Layer 2: Biome emoji at 80% size, 50% alpha
        const biomeEmoji = EMOJI_SYMBOLS[tile.biome] || EMOJI_SYMBOLS.default;
        const biomeFontSize = this.tileSize * 0.56; // 0.7 * 0.8 = 80% of normal

        this.ctx.save();
        this.ctx.globalAlpha = 0.5;
        this.ctx.font = `${biomeFontSize}px Arial`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(biomeEmoji, x, y);
        this.ctx.restore();

        // Layer 3: Entity glyph at 50% size (on top, full opacity)
        if (tile.entities && tile.entities.length > 0) {
            const entity = tile.entities[0];
            const entityEmoji = entity.glyph || EMOJI_SYMBOLS[entity.type] || '👤';
            const entityFontSize = this.tileSize * 0.35; // 0.7 * 0.5 = 50% of normal

            this.ctx.font = `${entityFontSize}px Arial`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText(entityEmoji, x, y);
        }

        // Portal indicator (special case - full size with glow)
        if (tile.portal?.active) {
            this.ctx.font = `${this.tileSize * 0.6}px Arial`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText(EMOJI_SYMBOLS.portal, x, y);
        }
    }

    private renderPlayer(x: number, y: number) {
        if (this.quality === 'low') {
            // ASCII player
            this.ctx.fillStyle = '#FFD700';
            this.ctx.font = `bold ${this.tileSize * 0.8}px monospace`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText('@', x, y);
        } else if (this.quality === 'high') {
            // Emoji player with highlight
            this.ctx.fillStyle = 'rgba(255, 215, 0, 0.3)';
            this.ctx.fillRect(
                x - this.tileSize / 2,
                y - this.tileSize / 2,
                this.tileSize,
                this.tileSize
            );
            this.ctx.font = `${this.tileSize * 0.7}px Arial`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText('🧍', x, y);
        } else {
            // Medium: colored circle
            this.ctx.fillStyle = '#FFD700';
            this.ctx.beginPath();
            this.ctx.arc(x, y, this.tileSize / 3, 0, Math.PI * 2);
            this.ctx.fill();
            this.ctx.strokeStyle = '#FFFFFF';
            this.ctx.lineWidth = 2;
            this.ctx.stroke();
        }

        // Highlight border for player tile
        this.ctx.strokeStyle = '#FFD700';
        this.ctx.lineWidth = 2;
        this.ctx.strokeRect(
            x - this.tileSize / 2 + 1,
            y - this.tileSize / 2 + 1,
            this.tileSize - 2,
            this.tileSize - 2
        );
    }

    private getBiomeColor(biome: string, alpha: number = 1): string {
        const biomeColors: Record<string, string> = {
            'Ocean': `rgba(70, 130, 180, ${alpha})`,
            'Rainforest': `rgba(34, 139, 34, ${alpha})`,
            'DeciduousForest': `rgba(60, 120, 60, ${alpha})`,
            'Taiga': `rgba(50, 80, 60, ${alpha})`,
            'Grassland': `rgba(144, 238, 144, ${alpha})`,
            'Desert': `rgba(244, 164, 96, ${alpha})`,
            'Tundra': `rgba(224, 255, 255, ${alpha})`,
            'Alpine': `rgba(128, 128, 128, ${alpha})`,
            // Lowercase variants
            'ocean': `rgba(70, 130, 180, ${alpha})`,
            'rainforest': `rgba(34, 139, 34, ${alpha})`,
            'forest': `rgba(34, 139, 34, ${alpha})`,
            'grassland': `rgba(144, 238, 144, ${alpha})`,
            'desert': `rgba(244, 164, 96, ${alpha})`,
            'tundra': `rgba(224, 255, 255, ${alpha})`,
            'mountain': `rgba(128, 128, 128, ${alpha})`,
            'swamp': `rgba(85, 107, 47, ${alpha})`,
            'lobby': `rgba(80, 80, 100, ${alpha})`,
            'void': `rgba(20, 20, 30, ${alpha})`,
        };
        return biomeColors[biome] || `rgba(51, 51, 51, ${alpha})`;
    }

    /**
     * Get hypsometric/bathymetric color for elevation.
     * Below sea level: dark blue (deep) -> light blue (shallow)
     * Above sea level: green (low) -> yellow -> brown -> gray -> white (peaks)
     */
    private getHypsometricColor(elevation: number): string {
        // Bathymetric (below sea level)
        if (elevation < -4000) return '#0a1929'; // Abyssal
        if (elevation < -2000) {
            const t = (elevation + 4000) / 2000;
            return this.lerpColor('#0a1929', '#0d3a5c', t);
        }
        if (elevation < -200) {
            const t = (elevation + 2000) / 1800;
            return this.lerpColor('#0d3a5c', '#1565c0', t);
        }
        if (elevation < 0) {
            const t = (elevation + 200) / 200;
            return this.lerpColor('#1565c0', '#4fc3f7', t);
        }

        // Hypsometric (above sea level)
        if (elevation < 200) {
            const t = elevation / 200;
            return this.lerpColor('#4fc3f7', '#66bb6a', t);
        }
        if (elevation < 500) {
            const t = (elevation - 200) / 300;
            return this.lerpColor('#66bb6a', '#aed581', t);
        }
        if (elevation < 1500) {
            const t = (elevation - 500) / 1000;
            return this.lerpColor('#aed581', '#dce775', t);
        }
        if (elevation < 3000) {
            const t = (elevation - 1500) / 1500;
            return this.lerpColor('#dce775', '#a1887f', t);
        }
        if (elevation < 5000) {
            const t = (elevation - 3000) / 2000;
            return this.lerpColor('#a1887f', '#9e9e9e', t);
        }
        // Peaks
        const t = Math.min(1, (elevation - 5000) / 3000);
        return this.lerpColor('#9e9e9e', '#fafafa', t);
    }

    /**
     * Linear interpolation between two hex colors.
     */
    private lerpColor(color1: string, color2: string, t: number): string {
        const c1 = this.hexToRgb(color1);
        const c2 = this.hexToRgb(color2);
        const r = Math.round(c1.r + (c2.r - c1.r) * t);
        const g = Math.round(c1.g + (c2.g - c1.g) * t);
        const b = Math.round(c1.b + (c2.b - c1.b) * t);
        return `rgb(${r}, ${g}, ${b})`;
    }

    private hexToRgb(hex: string): { r: number; g: number; b: number } {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
        return result ? {
            r: parseInt(result[1], 16),
            g: parseInt(result[2], 16),
            b: parseInt(result[3], 16)
        } : { r: 0, g: 0, b: 0 };
    }

    private getEntityColor(type: string): string {
        const entityColors: Record<string, string> = {
            'npc': '#FFFF00',
            'player': '#00FFFF',
            'resource': '#FFA500',
            'monster': '#FF0000',
            'item': '#FFA500',
        };
        return entityColors[type] || '#FFFFFF';
    }
}
