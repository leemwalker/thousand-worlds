// circular-buffer.ts
// Implements fixed-size circular buffer for game output messages
// Prevents unbounded memory growth from O(N) to O(1) with constant max size

export interface GameMessage {
    id: string;
    type: string;
    text: string;
    timestamp: number;
    data?: Record<string, any>;
}

const MAX_MESSAGES = 1000; // Maximum messages to keep in memory

export class CircularBuffer<T extends { id: string }> {
    private buffer: T[];
    private maxSize: number;
    private startIndex: number = 0;
    private size: number = 0;

    constructor(maxSize: number = MAX_MESSAGES) {
        this.maxSize = maxSize;
        this.buffer = new Array(maxSize);
    }

    // Add message to buffer (O(1))
    push(item: T): void {
        const index = (this.startIndex + this.size) % this.maxSize;
        this.buffer[index] = item;

        if (this.size < this.maxSize) {
            this.size++;
        } else {
            // Overwrite oldest, move start forward
            this.startIndex = (this.startIndex + 1) % this.maxSize;
        }
    }

    // Get all messages in order (O(N) where N = min(size, maxSize))
    getAll(): T[] {
        const result: T[] = [];
        for (let i = 0; i < this.size; i++) {
            const index = (this.startIndex + i) % this.maxSize;
            result.push(this.buffer[index]);
        }
        return result;
    }

    // Get recent N messages (O(N))
    getRecent(count: number): T[] {
        const actualCount = Math.min(count, this.size);
        const result: T[] = [];

        for (let i = this.size - actualCount; i < this.size; i++) {
            const index = (this.startIndex + i) % this.maxSize;
            result.push(this.buffer[index]);
        }

        return result;
    }

    // Get message by ID (O(N) - for searching)
    findById(id: string): T | undefined {
        for (let i = 0; i < this.size; i++) {
            const index = (this.startIndex + i) % this.maxSize;
            if (this.buffer[index].id === id) {
                return this.buffer[index];
            }
        }
        return undefined;
    }

    // Clear all messages (O(1))
    clear(): void {
        this.startIndex = 0;
        this.size = 0;
    }

    // Get current size
    getSize(): number {
        return this.size;
    }

    // Get maximum capacity
    getCapacity(): number {
        return this.maxSize;
    }

    // Check if buffer is full
    isFull(): boolean {
        return this.size === this.maxSize;
    }
}
