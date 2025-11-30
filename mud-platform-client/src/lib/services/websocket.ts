import { writable, get } from 'svelte/store';

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

    // Message handler
    private messageHandlers: Set<(msg: ServerMessage) => void> = new Set();

    connect(token: string, characterId?: string): void {
        this.token = token;
        let wsUrl = `ws://localhost:8080/api/game/ws?token=${encodeURIComponent(token)}`;
        if (characterId) {
            wsUrl += `&character_id=${encodeURIComponent(characterId)}`;
        }

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.connected.set(true);
                this.reconnectAttempts = 0;
            };

            this.ws.onmessage = (event) => {
                try {
                    const message: ServerMessage = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Failed to parse message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.connected.set(false);
                this.attemptReconnect();
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
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
