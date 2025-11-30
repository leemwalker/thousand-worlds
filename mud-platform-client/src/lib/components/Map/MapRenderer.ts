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
}

export interface MapEntity {
    id: string;
    type: 'npc' | 'player' | 'resource' | 'monster';
    x: number;
    y: number;
}

export class MapRenderer {
    private canvas: HTMLCanvasElement;
    private ctx: CanvasRenderingContext2D;
    private tileSize: number = 10; // pixels per coordinate unit
    private viewRadius: number = 50; // units visible around player

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

    setViewRadius(radius: number) {
        this.viewRadius = radius;
    }

    render(playerPos: Position, visibleArea: VisibleTile[]) {
        const centerX = this.canvas.width / 2;
        const centerY = this.canvas.height / 2;

        // Clear canvas
        this.ctx.fillStyle = '#000000';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Render fog of war first
        this.renderFogOfWar(playerPos, visibleArea, centerX, centerY);

        // Render visible tiles
        for (const tile of visibleArea) {
            const screenX = centerX + (tile.x - playerPos.x) * this.tileSize;
            const screenY = centerY + (tile.y - playerPos.y) * this.tileSize;

            this.renderTile(tile, screenX, screenY);
        }

        // Render entities (NPCs, resources, players)
        for (const tile of visibleArea) {
            if (tile.entities.length > 0) {
                const screenX = centerX + (tile.x - playerPos.x) * this.tileSize;
                const screenY = centerY + (tile.y - playerPos.y) * this.tileSize;

                for (const entity of tile.entities) {
                    this.renderEntity(entity, screenX, screenY);
                }
            }
        }

        // Render player (always centered)
        this.renderPlayer(centerX, centerY);
    }

    private renderFogOfWar(
        playerPos: Position,
        visibleArea: VisibleTile[],
        centerX: number,
        centerY: number
    ) {
        const visibleSet = new Set(visibleArea.map(t => `${t.x},${t.y}`));

        // Render fog for all tiles in view radius
        for (let x = -this.viewRadius; x <= this.viewRadius; x++) {
            for (let y = -this.viewRadius; y <= this.viewRadius; y++) {
                const worldX = playerPos.x + x;
                const worldY = playerPos.y + y;

                if (!visibleSet.has(`${worldX},${worldY}`)) {
                    const screenX = centerX + x * this.tileSize;
                    const screenY = centerY + y * this.tileSize;

                    this.ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
                    this.ctx.fillRect(
                        screenX - this.tileSize / 2,
                        screenY - this.tileSize / 2,
                        this.tileSize,
                        this.tileSize
                    );
                }
            }
        }
    }

    private renderTile(tile: VisibleTile, x: number, y: number) {
        // Terrain color based on biome
        this.ctx.fillStyle = this.getBiomeColor(tile.biome);
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );

        // Elevation shading
        const elevationAlpha = Math.max(0, Math.min(0.3, tile.elevation / 5000));
        this.ctx.fillStyle = `rgba(255, 255, 255, ${elevationAlpha})`;
        this.ctx.fillRect(
            x - this.tileSize / 2,
            y - this.tileSize / 2,
            this.tileSize,
            this.tileSize
        );
    }

    private renderEntity(entity: MapEntity, x: number, y: number) {
        const radius = this.tileSize / 3;

        this.ctx.beginPath();
        this.ctx.arc(x, y, radius, 0, 2 * Math.PI);
        this.ctx.fillStyle = this.getEntityColor(entity.type);
        this.ctx.fill();

        // Entity outline
        this.ctx.strokeStyle = '#FFFFFF';
        this.ctx.lineWidth = 1;
        this.ctx.stroke();
    }

    private renderPlayer(x: number, y: number) {
        const radius = this.tileSize / 2;

        // Player circle
        this.ctx.beginPath();
        this.ctx.arc(x, y, radius, 0, 2 * Math.PI);
        this.ctx.fillStyle = '#00FF00';
        this.ctx.fill();

        // Player outline
        this.ctx.strokeStyle = '#FFFFFF';
        this.ctx.lineWidth = 2;
        this.ctx.stroke();

        // Direction indicator
        this.ctx.beginPath();
        this.ctx.moveTo(x, y);
        this.ctx.lineTo(x, y - radius * 1.5);
        this.ctx.strokeStyle = '#FFFFFF';
        this.ctx.lineWidth = 2;
        this.ctx.stroke();
    }

    private getBiomeColor(biome: string): string {
        const biomeColors: Record<string, string> = {
            'forest': '#228B22',
            'grassland': '#90EE90',
            'desert': '#F4A460',
            'mountain': '#808080',
            'tundra': '#E0FFFF',
            'ocean': '#4682B4',
            'swamp': '#556B2F',
        };
        return biomeColors[biome] || '#333333';
    }

    private getEntityColor(type: string): string {
        const entityColors: Record<string, string> = {
            'npc': '#FFFF00',
            'player': '#00FFFF',
            'resource': '#FFA500',
            'monster': '#FF0000',
        };
        return entityColors[type] || '#FFFFFF';
    }
}
