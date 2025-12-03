import { writable, get } from 'svelte/store';
import { commandQueue } from './command-queue';

export interface ServerMessage {
    type: string;
    data: any;
}

export interface CommandData {
    action: string;
    target?: string;
    direction?: string;
    quantity?: number;
}

export class GameWebSocket {
    private ws: WebSocket | null = null;
    private token: string = '';
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;

    // Store for connection status
    public connected = writable<boolean>(false);

    // Store for pending command count
    public pendingCommands = writable<number>(0);

    // Message handler
    private messageHandlers: Set<(msg: ServerMessage) => void> = new Set();

    connect(token: string, characterId?: string): void {
        console.log('[WebSocket] Attempting to connect...', { token: token.substring(0, 20) + '...', characterId });
        this.token = token;

        // Use the same host that served the page, but with ws:// protocol
        // This works for both localhost development and mobile access via LAN IP
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.hostname + ':8080'; // Backend always on 8080
        let wsUrl = `${protocol}//${host}/api/game/ws?token=${encodeURIComponent(token)}`;

        if (characterId) {
            wsUrl += `&character_id=${encodeURIComponent(characterId)}`;
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
                    const message: ServerMessage = JSON.parse(event.data);
                    console.log('[WebSocket] Message received:', message);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('[WebSocket] Failed to parse message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('[WebSocket] Error:', error);
            };

            this.ws.onclose = () => {
                console.log('[WebSocket] Connection closed');
                this.connected.set(false);
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('[WebSocket] Failed to create WebSocket:', error);
        }
    }

    disconnect(): void {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.connected.set(false);
    }

    sendCommand(command: CommandData): void {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.error('WebSocket not connected');
            return;
        }

        const message = {
            type: 'command',
            data: command,
        };

        this.ws.send(JSON.stringify(message));
    }

    async sendCommandWithQueue(command: CommandData): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            // Queue command for later
            await commandQueue.enqueue(command);
            await this.updatePendingCount();
            console.log('Command queued for later sending');
            return;
        }

        this.sendCommand(command);
    }

    private async processQueuedCommands(): Promise<void> {
        try {
            await commandQueue.processQueue(async (cmd) => {
                return new Promise((resolve, reject) => {
                    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                        reject(new Error('WebSocket not connected'));
                        return;
                    }

                    const message = {
                        type: 'command',
                        data: cmd
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
            if (this.token) {
                this.connect(this.token);
            }
        }, delay);
    }
}

// Singleton instance
export const gameWebSocket = new GameWebSocket();
