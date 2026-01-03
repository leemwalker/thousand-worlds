/**
 * PerformanceOverlay - Debug overlay for FPS performance monitoring.
 * Phase 3: Performance Optimization
 * 
 * Displays real-time metrics:
 * - FPS counter
 * - Draw calls
 * - Triangle count
 * - Active chunks
 * - Render resolution
 * - Current altitude stage
 */

import type { PerformanceStats } from './FPSPerformanceManager';

export interface PerformanceOverlayOptions {
    position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
    opacity?: number;
    fontSize?: number;
}

/**
 * Creates and manages a performance monitoring overlay.
 */
export class PerformanceOverlay {
    private container: HTMLDivElement | null = null;
    private fpsElement: HTMLDivElement | null = null;
    private statsElement: HTMLDivElement | null = null;
    private visible: boolean = false;

    constructor(parentElement: HTMLElement, options: PerformanceOverlayOptions = {}) {
        this.createOverlay(parentElement, options);
    }

    /**
     * Create the overlay DOM elements.
     */
    private createOverlay(parent: HTMLElement, options: PerformanceOverlayOptions): void {
        const position = options.position ?? 'top-left';
        const opacity = options.opacity ?? 0.8;
        const fontSize = options.fontSize ?? 12;

        // Container
        this.container = document.createElement('div');
        this.container.style.cssText = `
            position: absolute;
            ${position.includes('top') ? 'top: 10px' : 'bottom: 10px'};
            ${position.includes('left') ? 'left: 10px' : 'right: 10px'};
            background: rgba(0, 0, 0, ${opacity});
            color: #00ff00;
            font-family: 'Consolas', 'Monaco', monospace;
            font-size: ${fontSize}px;
            padding: 8px 12px;
            border-radius: 4px;
            z-index: 1000;
            pointer-events: none;
            display: none;
            min-width: 160px;
        `;

        // FPS counter (larger)
        this.fpsElement = document.createElement('div');
        this.fpsElement.style.cssText = `
            font-size: ${fontSize * 1.5}px;
            font-weight: bold;
            margin-bottom: 6px;
            border-bottom: 1px solid #444;
            padding-bottom: 4px;
        `;
        this.fpsElement.textContent = 'FPS: --';

        // Detailed stats
        this.statsElement = document.createElement('div');
        this.statsElement.style.cssText = `
            line-height: 1.5;
        `;
        this.statsElement.innerHTML = `
            <div>Draw Calls: --</div>
            <div>Triangles: --</div>
            <div>Chunks: --</div>
            <div>Resolution: --</div>
            <div>Altitude: --</div>
        `;

        this.container.appendChild(this.fpsElement);
        this.container.appendChild(this.statsElement);
        parent.appendChild(this.container);
    }

    /**
     * Update the overlay with current stats.
     */
    update(stats: PerformanceStats): void {
        if (!this.visible || !this.fpsElement || !this.statsElement) return;

        // Color-code FPS
        let fpsColor = '#00ff00'; // Green = good
        if (stats.fps < 30) {
            fpsColor = '#ff0000'; // Red = bad
        } else if (stats.fps < 45) {
            fpsColor = '#ffaa00'; // Orange = warning
        }

        this.fpsElement.style.color = fpsColor;
        this.fpsElement.textContent = `FPS: ${stats.fps}`;

        // Format triangle count
        const triangleStr = stats.triangles > 1000000
            ? `${(stats.triangles / 1000000).toFixed(1)}M`
            : stats.triangles > 1000
                ? `${(stats.triangles / 1000).toFixed(1)}K`
                : stats.triangles.toString();

        this.statsElement.innerHTML = `
            <div>Draw Calls: ${stats.drawCalls}</div>
            <div>Triangles: ${triangleStr}</div>
            <div>Chunks: ${stats.activeChunks}</div>
            <div>Resolution: ${(stats.resolution * 100).toFixed(0)}%</div>
            <div>Stage: ${stats.altitude}</div>
        `;
    }

    /**
     * Show the overlay.
     */
    show(): void {
        this.visible = true;
        if (this.container) {
            this.container.style.display = 'block';
        }
    }

    /**
     * Hide the overlay.
     */
    hide(): void {
        this.visible = false;
        if (this.container) {
            this.container.style.display = 'none';
        }
    }

    /**
     * Toggle overlay visibility.
     */
    toggle(): void {
        if (this.visible) {
            this.hide();
        } else {
            this.show();
        }
    }

    /**
     * Check if overlay is visible.
     */
    isVisible(): boolean {
        return this.visible;
    }

    /**
     * Dispose of overlay.
     */
    dispose(): void {
        if (this.container && this.container.parentNode) {
            this.container.parentNode.removeChild(this.container);
        }
        this.container = null;
        this.fpsElement = null;
        this.statsElement = null;
    }
}
