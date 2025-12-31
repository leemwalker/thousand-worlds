/**
 * Frontend Performance Monitor (Sprint 3)
 * 
 * Tracks FPS, texture upload times, memory usage, and bandwidth.
 * Logged to console periodically for development diagnostics.
 */

interface PerformanceMetrics {
    fps: number;
    frameTime: number;
    textureUploadTime: number;
    memoryUsedMB: number;
    bandwidthMB: number;
}

class PerformanceMonitor {
    private frameCount = 0;
    private lastFPSCheck = performance.now();
    private currentFPS = 60;
    private frameTimeSum = 0;
    private textureUploadTimes: number[] = [];
    private bandwidthTotal = 0;
    private lastFrameTime = performance.now();
    private enabled = false;
    private logIntervalId: number | null = null;

    /**
     * Start monitoring (call once on app init)
     */
    start(logIntervalMs = 10000): void {
        if (this.enabled) return;
        this.enabled = true;
        this.lastFPSCheck = performance.now();
        this.lastFrameTime = performance.now();

        // Log metrics periodically
        this.logIntervalId = window.setInterval(() => {
            this.logMetrics();
        }, logIntervalMs);

        console.log('[PerformanceMonitor] Started, logging every', logIntervalMs, 'ms');
    }

    /**
     * Stop monitoring
     */
    stop(): void {
        this.enabled = false;
        if (this.logIntervalId !== null) {
            window.clearInterval(this.logIntervalId);
            this.logIntervalId = null;
        }
    }

    /**
     * Call this every frame (in requestAnimationFrame loop)
     */
    recordFrame(): void {
        if (!this.enabled) return;

        const now = performance.now();
        const frameDelta = now - this.lastFrameTime;
        this.lastFrameTime = now;

        this.frameCount++;
        this.frameTimeSum += frameDelta;

        const elapsed = now - this.lastFPSCheck;
        if (elapsed >= 1000) {
            this.currentFPS = Math.round((this.frameCount / elapsed) * 1000);
            this.frameCount = 0;
            this.frameTimeSum = 0;
            this.lastFPSCheck = now;
        }
    }

    /**
     * Record a texture upload time (call after WebGL texture upload)
     */
    recordTextureUpload(timeMs: number): void {
        this.textureUploadTimes.push(timeMs);

        // Keep only last 10 measurements for rolling average
        if (this.textureUploadTimes.length > 10) {
            this.textureUploadTimes.shift();
        }
    }

    /**
     * Record bandwidth usage (call when receiving data)
     */
    recordBandwidth(bytes: number): void {
        this.bandwidthTotal += bytes;
    }

    /**
     * Get current performance metrics
     */
    getMetrics(): PerformanceMetrics {
        const avgUploadTime =
            this.textureUploadTimes.length > 0
                ? this.textureUploadTimes.reduce((a, b) => a + b, 0) /
                this.textureUploadTimes.length
                : 0;

        // performance.memory is non-standard but supported in Chrome/Edge
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const memoryInfo = (performance as any).memory;
        const memoryUsedMB = memoryInfo
            ? memoryInfo.usedJSHeapSize / 1024 / 1024
            : 0;

        return {
            fps: this.currentFPS,
            frameTime:
                this.frameCount > 0
                    ? this.frameTimeSum / this.frameCount
                    : 16.67,
            textureUploadTime: avgUploadTime,
            memoryUsedMB,
            bandwidthMB: this.bandwidthTotal / 1024 / 1024,
        };
    }

    /**
     * Log metrics to console
     */
    logMetrics(): void {
        const m = this.getMetrics();
        console.log(
            `📊 [Performance] FPS: ${m.fps} | Frame: ${m.frameTime.toFixed(1)}ms | ` +
            `Texture: ${m.textureUploadTime.toFixed(1)}ms | Memory: ${m.memoryUsedMB.toFixed(1)}MB | ` +
            `Bandwidth: ${m.bandwidthMB.toFixed(2)}MB`
        );
    }

    /**
     * Reset bandwidth counter (call periodically or on session start)
     */
    resetBandwidth(): void {
        this.bandwidthTotal = 0;
    }
}

// Singleton instance
export const performanceMonitor = new PerformanceMonitor();

// Expose on window for cross-module instrumentation
if (typeof window !== 'undefined') {
    (window as any).__performanceMonitor = performanceMonitor;
}

// Auto-start in development mode
if (typeof window !== 'undefined' && import.meta.env.DEV) {
    performanceMonitor.start(10000); // Log every 10s in dev
}
