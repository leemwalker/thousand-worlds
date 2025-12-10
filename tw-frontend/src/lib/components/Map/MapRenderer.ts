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

    render(playerPos: Position, visibleArea: VisibleTile[]) {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;

        // Clear canvas with dark background
        this.ctx.fillStyle = '#1a1a2e';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Render visible tiles
        for (const tile of visibleArea) {
            const screenX = centerX + (tile.x - playerPos.x) * this.tileSize;
            const screenY = centerY + (tile.y - playerPos.y) * this.tileSize;

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
        // Full color background
        this.ctx.fillStyle = this.getBiomeColor(tile.biome, 0.8);
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Subtle grid
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
        this.ctx.strokeRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Determine emoji to show
        let emoji = EMOJI_SYMBOLS[tile.biome] || EMOJI_SYMBOLS.default;

        // Portal takes priority
        if (tile.portal?.active) {
            emoji = EMOJI_SYMBOLS.portal;
        }

        // Entity takes higher priority
        if (tile.entities && tile.entities.length > 0) {
            const entity = tile.entities[0];
            emoji = EMOJI_SYMBOLS[entity.type] || '👤';
        }

        // Draw emoji
        this.ctx.font = `${this.tileSize * 0.7}px Arial`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(emoji, x, y);
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
        };
        return biomeColors[biome] || `rgba(51, 51, 51, ${alpha})`;
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
