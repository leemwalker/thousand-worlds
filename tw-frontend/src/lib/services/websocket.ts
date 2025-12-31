import { writable, get } from 'svelte/store';
import { commandQueue } from './command-queue';
import { mapStore } from '$lib/stores/map';
import { gameStore } from '$lib/stores/game';
import type { ServerMessage, ClientMessage, GameCommand, GameOutputMessage } from '$lib/types/websocket';
import { validateServerMessage } from '$lib/types/schemas';

export class GameWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;

    private currentCharacterId: string | undefined;
    private isIntentionalDisconnect = false;

    // Store for connection status
    public connected = writable<boolean>(false);

    // Store for pending command count
    public pendingCommands = writable<number>(0);

    // Message handler
    private messageHandlers: Set<(msg: ServerMessage) => void> = new Set();

    // Reconnection callbacks - fire when connection is re-established after disconnect
    private reconnectCallbacks: Set<() => void> = new Set();
    private wasConnectedBefore = false;

    connect(characterId?: string): void {
        console.log('[WebSocket] Attempting to connect...', { characterId });
        this.isIntentionalDisconnect = false;

        if (characterId) {
            this.currentCharacterId = characterId;
        }

        // Build WebSocket URL
        // In development (Vite proxy on port 5173), use the same host
        // In production, WebSocket must go directly to game-server on port 8080
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const hostname = window.location.hostname;
        const port = window.location.port;

        // Use environment variable if available, otherwise fallback to logic
        const wsHost = import.meta.env.VITE_WS_URL ||
            (port === '5173' ? `${hostname}:${port}` : `${hostname}:8080`);

        let wsUrl = `${protocol}//${wsHost}/api/game/ws`;

        if (this.currentCharacterId) {
            wsUrl += `?character_id=${encodeURIComponent(this.currentCharacterId)}`;
        }

        console.log('[WebSocket] URL:', wsUrl);

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('[WebSocket] Connection opened!');
                this.connected.set(true);

                // Track if this was a reconnect (not first connect)
                const isReconnect = this.reconnectAttempts > 0 || this.wasConnectedBefore;
                this.reconnectAttempts = 0;
                this.wasConnectedBefore = true;

                gameStore.setLoading(false); // Stop loading if valid connection

                // Process any queued commands
                this.processQueuedCommands();

                // Fire reconnection callbacks if this was a reconnect
                if (isReconnect) {
                    console.log('[WebSocket] Reconnection successful, firing callbacks');
                    this.reconnectCallbacks.forEach(cb => {
                        try { cb(); } catch (e) { console.error('[WebSocket] Reconnect callback error:', e); }
                    });
                }
            };

            this.ws.binaryType = 'arraybuffer'; // Enable binary messages

            this.ws.onmessage = (event) => {
                try {
                    // Handle Binary Messages (ArrayBuffer)
                    if (event.data instanceof ArrayBuffer) {
                        this.handleBinaryMessage(event.data);
                        return;
                    }

                    // Check if data contains multiple JSON objects (concatenated or newline separated)
                    const rawData = event.data.toString();

                    // Split by newline if present
                    const parts = rawData.split('\n').filter((p: string) => p.trim() !== '');

                    for (const part of parts) {
                        try {
                            const parsed = JSON.parse(part);

                            // Validate message structure with Zod (logs warnings, doesn't block)
                            const validation = validateServerMessage(parsed);
                            if (!validation.valid) {
                                // Log but continue processing - defensive coding
                                console.warn('[WebSocket] Message failed validation, processing anyway');
                            }

                            const message: ServerMessage = parsed;
                            // console.log('[WebSocket] Message received:', message.type); 
                            this.handleMessage(message);
                        } catch (e) {
                            console.error('[WebSocket] Failed to parse message part:', e);
                        }
                    }
                } catch (error) {
                    console.error('[WebSocket] Failed to process message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('[WebSocket] Error:', error);
                gameStore.setError('Connection error');
            };

            this.ws.onclose = () => {
                console.log('[WebSocket] Connection closed');
                this.connected.set(false);
                if (!this.isIntentionalDisconnect) {
                    this.attemptReconnect();
                }
            };
        } catch (error) {
            console.error('[WebSocket] Failed to create WebSocket:', error);
            gameStore.setError('Failed to create connection');
        }
    }

    disconnect(): void {
        this.isIntentionalDisconnect = true;
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.connected.set(false);
    }

    sendRawCommand(text: string, payload?: any): void {
        console.log('[WebSocket] sendRawCommand called:', text, payload);
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.error('[WebSocket] Not connected, readyState:', this.ws?.readyState);
            return;
        }

        const message: GameCommand = {
            type: 'command',
            data: { text, payload },
        };

        console.log('[WebSocket] Sending command:', JSON.stringify(message));
        this.ws.send(JSON.stringify(message));
        console.log('[WebSocket] Command sent successfully');
    }

    async sendCommandWithQueue(text: string): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            // Queue command for later
            await commandQueue.enqueue(text);
            await this.updatePendingCount();
            console.log('Command queued for later sending');
            return;
        }

        this.sendRawCommand(text);
    }

    private async processQueuedCommands(): Promise<void> {
        try {
            await commandQueue.processQueue(async (text) => {
                return new Promise((resolve, reject) => {
                    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                        reject(new Error('WebSocket not connected'));
                        return;
                    }

                    const message: GameCommand = {
                        type: 'command',
                        data: { text }
                    };

                    this.ws.send(JSON.stringify(message));
                    resolve();
                });
            });

            await this.updatePendingCount();
        } catch (error) {
            console.error('Error processing queued commands:', error);
        }
    }

    private async updatePendingCount(): Promise<void> {
        const count = await commandQueue.getPendingCount();
        this.pendingCommands.set(count);
    }

    onMessage(handler: (msg: ServerMessage) => void): () => void {
        this.messageHandlers.add(handler);

        // Return unsubscribe function
        return () => {
            this.messageHandlers.delete(handler);
        };
    }

    /**
     * Register a callback to be called when WebSocket reconnects after a disconnect.
     * Useful for refreshing world map or other state after connection is restored.
     * Returns an unsubscribe function.
     */
    onReconnect(handler: () => void): () => void {
        this.reconnectCallbacks.add(handler);

        // Return unsubscribe function
        return () => {
            this.reconnectCallbacks.delete(handler);
        };
    }

    private handleMessage(message: ServerMessage): void {
        // Handle map updates
        if (message.type === 'map_update') {
            // console.log('[WS] map_update received');
            const mappedData = this.mapBackendToFrontend(message.data);
            mapStore.setMapData(mappedData);
        } else if (message.type === 'game_message') {
            // Handle legacy embedded map updates
            const gameMsg = message as GameOutputMessage;
            if (gameMsg.data?.metadata && gameMsg.data.type === 'map_update') {
                const mappedData = this.mapBackendToFrontend(gameMsg.data.metadata);
                mapStore.setMapData(mappedData);
            }

            // Add to game store
            gameStore.addMessage({
                type: 'game_message',
                timestamp: message.timestamp || Date.now(),
                content: gameMsg.data.content,
                sender: gameMsg.data.sender,
                channel: gameMsg.data.channel
            });
        } else if (message.type === 'state_update') {
            gameStore.updateStats(message.data.stats || {});
            if (message.data.inventory) {
                gameStore.setInventory(message.data.inventory);
            }
        }

        // Notify all handlers
        this.messageHandlers.forEach(handler => {
            try {
                handler(message);
            } catch (error) {
                console.error('Message handler error:', error);
            }
        });
    }

    private mapBackendToFrontend(data: any): any {
        // Check if data is already in frontend format (has 'tiles' array)
        if (data && Array.isArray(data.tiles)) {
            return data;
        }

        // Check if data is in backend format (has 'cells' array)
        if (data && Array.isArray(data.cells)) {
            return {
                ...data,
                tiles: data.cells.map((cell: any) => ({
                    x: cell.q,
                    y: cell.r,
                    biome: cell.biome_type || "Default",
                    elevation: cell.elevation || 0,
                    // Map other fields if needed, or pass through extra props usually ignored
                    occluded: cell.occluded,
                    is_player: cell.is_player,
                    entities: cell.entities,
                    portal: cell.portal
                }))
            };
        }

        return data; // Return as-is if unrecognized
    }

    private handleBinaryMessage(buffer: ArrayBuffer) {
        // Format: [Type (1 byte)] [JSON Length (4 bytes)] [JSON Data] [Binary Section Length (4 bytes)] [Binary Section]
        // Binary Section: [ImageLen:4][Image][GridLen:4][Grid]
        const view = new DataView(buffer);
        let offset = 0;

        // 1. Read Type
        const msgType = view.getUint8(offset);
        offset += 1;

        if (msgType === 0x01) { // WorldMapImage
            // 2. Read JSON Length
            const jsonLen = view.getUint32(offset, false); // Big Endian
            offset += 4;

            // 3. Read JSON Data
            const jsonBytes = new Uint8Array(buffer, offset, jsonLen);
            const jsonStr = new TextDecoder().decode(jsonBytes);
            const jsonData = JSON.parse(jsonStr);
            offset += jsonLen;

            // 4. Read Binary Section Length
            const binSectionLen = view.getUint32(offset, false);
            offset += 4;

            // 5. Parse Binary Section: [ImageLen:4][Image][GridLen:4][Grid]
            // Read Image Length and Data
            const imageLen = view.getUint32(offset, false);
            offset += 4;

            const imageBytes = new Uint8Array(buffer, offset, imageLen);
            const imageBlob = new Blob([imageBytes], { type: 'image/webp' });
            offset += imageLen;

            // Read Grid Length and Data (if present)
            let gridData: ArrayBuffer | null = null;
            if (offset + 4 <= buffer.byteLength) {
                const gridLen = view.getUint32(offset, false);
                offset += 4;

                if (gridLen > 0 && offset + gridLen <= buffer.byteLength) {
                    // Copy grid data to a new ArrayBuffer for ownership
                    gridData = buffer.slice(offset, offset + gridLen);
                    offset += gridLen;
                    console.log(`[WebSocket] Parsed grid data: ${gridLen} bytes`);
                }
            }

            // Construct a ServerMessage to dispatch
            const message: ServerMessage = {
                type: 'world_map_image_response',
                data: {
                    ...jsonData,
                    imageBlob: imageBlob, // WebP image blob
                    gridData: gridData     // Binary grid data (or null if not present)
                },
                timestamp: Date.now()
            };

            this.handleMessage(message);
        } else {
            console.warn('[WebSocket] Unknown binary message type:', msgType);
        }
    }

    private attemptReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('[WebSocket] Max reconnect attempts reached');
            gameStore.setError('Connection lost. Please refresh.');
            return;
        }

        this.reconnectAttempts++;

        // Exponential backoff: 1s, 2s, 4s, 8s, 16s, then cap at 30s
        const baseDelay = this.reconnectDelay; // 1000ms
        const exponentialDelay = Math.min(
            baseDelay * Math.pow(2, this.reconnectAttempts - 1),
            30000 // Cap at 30 seconds
        );

        // Add jitter (±20% randomness to prevent thundering herd)
        const jitter = exponentialDelay * 0.2 * (Math.random() * 2 - 1);
        const delay = Math.round(exponentialDelay + jitter);

        console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

        setTimeout(() => {
            this.connect(this.currentCharacterId);
        }, delay);
    }

    /**
     * Request world map refresh after reconnection.
     * Called automatically when connection is re-established.
     */
    requestWorldMapRefresh(): void {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            console.log('[WebSocket] Requesting world map refresh after reconnection');
            this.sendRawCommand('world_map_image', {});
        }
    }
}

// Singleton instance
export const gameWebSocket = new GameWebSocket();
