import { writable, get } from 'svelte/store';
import { commandQueue } from './command-queue';
import { mapStore } from '$lib/stores/map';

export interface ServerMessage {
    type: string;
    data: any;
}

// Command message structure - send raw text to backend
export interface CommandMessage {
    type: 'command';
    data: {
        text: string;
    };
}

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

        // If running on frontend's production port (3000) or any non-dev port, 
        // connect directly to backend on port 8080
        const isDevServer = port === '5173';
        const wsHost = isDevServer
            ? `${hostname}:${port}` // Use Vite proxy in development
            : `${hostname}:8080`;   // Direct to backend in production

        let wsUrl = `${protocol}//${wsHost}/api/game/ws`;


        if (this.currentCharacterId) {
            wsUrl += `?character_id=${encodeURIComponent(this.currentCharacterId)}`;
        }

        console.log('[WebSocket] URL:', wsUrl);

        try {
            this.ws = new WebSocket(wsUrl);
            console.log('[WebSocket] WebSocket object created');

            this.ws.onopen = () => {
                console.log('[WebSocket] Connection opened!');
                this.connected.set(true);
                this.reconnectAttempts = 0;

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
                            console.log('[WebSocket] Message received:', message);
                            this.handleMessage(message);
                        } catch (e) {
                            console.error('[WebSocket] Failed to parse message part:', e);
                        }
                    }
                } catch (error) {
                    console.error('[WebSocket] Failed to process message:', error);
                    console.log('[WebSocket] Raw data:', event.data);
                }
            };

            this.ws.onerror = (error) => {
                console.error('[WebSocket] Error:', error);
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
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }

        const message: CommandMessage = {
            type: 'command',
            data: { text },
        };

        this.ws.send(JSON.stringify(message));
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

                    const message: CommandMessage = {
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
        // Handle map updates - they come as game_message with data.type='map_update'
        // The actual map data is in data.metadata
        if (message.type === 'game_message' && message.data?.type === 'map_update' && message.data?.metadata) {
            console.log('[WS] Received map_update:', message.data.metadata);
            mapStore.setMapData(message.data.metadata);
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

    private attemptReconnect(): void {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnect attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;

        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

        setTimeout(() => {
            this.connect(this.currentCharacterId);
        }, delay);
    }
}

// Singleton instance
export const gameWebSocket = new GameWebSocket();
