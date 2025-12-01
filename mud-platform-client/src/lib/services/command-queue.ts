// command-queue.ts
// Implements outbox pattern with IndexedDB for offline command queueing
// Automatically retries pending commands when connection is restored

import type { CommandData } from './websocket';

interface QueuedCommand {
    id: string;
    command: CommandData;
    timestamp: number;
    retryCount: number;
    status: 'pending' | 'processing' | 'failed';
}

const DB_NAME = 'mud-command-queue';
const DB_VERSION = 1;
const STORE_NAME = 'commands';
const MAX_RETRIES = 3;

export class CommandQueue {
    private db: IDBDatabase | null = null;
    private isProcessing = false;

    async init(): Promise<void> {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(DB_NAME, DB_VERSION);

            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
                this.db = request.result;
                resolve();
            };

            request.onupgradeneeded = (event) => {
                const db = (event.target as IDBOpenDBRequest).result;

                if (!db.objectStoreNames.contains(STORE_NAME)) {
                    const store = db.createObjectStore(STORE_NAME, { keyPath: 'id' });
                    store.createIndex('status', 'status', { unique: false });
                    store.createIndex('timestamp', 'timestamp', { unique: false });
                }
            };
        });
    }

    // Enqueue command for later sending (O(1) IndexedDB write)
    async enqueue(command: CommandData): Promise<string> {
        if (!this.db) throw new Error('Database not initialized');

        const queuedCommand: QueuedCommand = {
            id: crypto.randomUUID(),
            command,
            timestamp: Date.now(),
            retryCount: 0,
            status: 'pending'
        };

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
            const store = transaction.objectStore(STORE_NAME);
            const request = store.add(queuedCommand);

            request.onsuccess = () => resolve(queuedCommand.id);
            request.onerror = () => reject(request.error);
        });
    }

    // Get all pending commands (for UI display)
    async getPending(): Promise<QueuedCommand[]> {
        if (!this.db) return [];

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readonly');
            const store = transaction.objectStore(STORE_NAME);
            const index = store.index('status');
            const request = index.getAll('pending');

            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    // Get count of pending commands
    async getPendingCount(): Promise<number> {
        if (!this.db) return 0;

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readonly');
            const store = transaction.objectStore(STORE_NAME);
            const index = store.index('status');
            const request = index.count('pending');

            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    // Process queue when connection restored
    async processQueue(sendFn: (cmd: CommandData) => Promise<void>): Promise<void> {
        if (!this.db || this.isProcessing) return;

        this.isProcessing = true;

        try {
            const pending = await this.getPending();

            for (const queuedCmd of pending) {
                await this.updateStatus(queuedCmd.id, 'processing');

                try {
                    await sendFn(queuedCmd.command);
                    await this.remove(queuedCmd.id);
                } catch (error) {
                    console.error('Failed to send queued command:', error);

                    if (queuedCmd.retryCount < MAX_RETRIES) {
                        await this.incrementRetry(queuedCmd.id);
                    } else {
                        await this.updateStatus(queuedCmd.id, 'failed');
                    }
                }
            }
        } finally {
            this.isProcessing = false;
        }
    }

    // Clear all failed commands
    async clearFailed(): Promise<void> {
        if (!this.db) return;

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
            const store = transaction.objectStore(STORE_NAME);
            const index = store.index('status');
            const request = index.openCursor(IDBKeyRange.only('failed'));

            request.onsuccess = (event) => {
                const cursor = (event.target as IDBRequest).result;
                if (cursor) {
                    cursor.delete();
                    cursor.continue();
                } else {
                    resolve();
                }
            };

            request.onerror = () => reject(request.error);
        });
    }

    // Private helpers

    private async updateStatus(id: string, status: QueuedCommand['status']): Promise<void> {
        if (!this.db) return;

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
            const store = transaction.objectStore(STORE_NAME);
            const getRequest = store.get(id);

            getRequest.onsuccess = () => {
                const cmd = getRequest.result;
                if (cmd) {
                    cmd.status = status;
                    const putRequest = store.put(cmd);
                    putRequest.onsuccess = () => resolve();
                    putRequest.onerror = () => reject(putRequest.error);
                } else {
                    resolve();
                }
            };

            getRequest.onerror = () => reject(getRequest.error);
        });
    }

    private async incrementRetry(id: string): Promise<void> {
        if (!this.db) return;

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
            const store = transaction.objectStore(STORE_NAME);
            const getRequest = store.get(id);

            getRequest.onsuccess = () => {
                const cmd = getRequest.result;
                if (cmd) {
                    cmd.retryCount++;
                    cmd.status = 'pending';
                    const putRequest = store.put(cmd);
                    putRequest.onsuccess = () => resolve();
                    putRequest.onerror = () => reject(putRequest.error);
                } else {
                    resolve();
                }
            };

            getRequest.onerror = () => reject(getRequest.error);
        });
    }

    private async remove(id: string): Promise<void> {
        if (!this.db) return;

        return new Promise((resolve, reject) => {
            const transaction = this.db!.transaction([STORE_NAME], 'readwrite');
            const store = transaction.objectStore(STORE_NAME);
            const request = store.delete(id);

            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    // Cleanup on close
    close(): void {
        if (this.db) {
            this.db.close();
            this.db = null;
        }
    }
}

// Singleton instance
export const commandQueue = new CommandQueue();
