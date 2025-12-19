import { writable, get } from 'svelte/store';
import { commandQueue } from './command-queue';
import { mapStore } from '$lib/stores/map';
import { gameStore } from '$lib/stores/game';
import type { ServerMessage, ClientMessage, GameCommand, GameOutputMessage } from '$lib/types/websocket';

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
                this.reconnectAttempts = 0;
                gameStore.setLoading(false); // Stop loading if valid connection

                // Process any queued commands
                this.processQueuedCommands();
            };

            this.ws.onmessage = (event) => {
                try {
                    // Check if data contains multiple JSON objects (concatenated or newline separated)
                    const rawData = event.data.toString();

                    // Split by newline if present
                    const parts = rawData.split('\n').filter((p: string) => p.trim() !== '');

                    for (const part of parts) {
                        try {
                            const message: ServerMessage = JSON.parse(part);
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

    sendRawCommand(text: string): void {
        console.log('[WebSocket] sendRawCommand called:', text);
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.error('[WebSocket] Not connected, readyState:', this.ws?.readyState);
            return;
        }

        const message: GameCommand = {
            type: 'command',
            data: { text },
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

    private attemptReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnect attempts reached');
            gameStore.setError('Connection lost. Please refresh.');
            return;
        }

        this.reconnectAttempts++;
        // Add jitter
        const jitter = this.reconnectDelay * 0.2 * (Math.random() - 0.5);
        const delay = (this.reconnectDelay * this.reconnectAttempts) + jitter;

        console.log(`Reconnecting in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts})`);

        setTimeout(() => {
            this.connect(this.currentCharacterId);
        }, delay);
    }
}

// Singleton instance
export const gameWebSocket = new GameWebSocket();
