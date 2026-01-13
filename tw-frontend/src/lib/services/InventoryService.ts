import type { InventoryItem } from '$lib/types/ui';
import { writable, type Writable } from 'svelte/store';

export class InventoryService {
    private items: InventoryItem[] = [];
    private maxWeight: number;
    private currentWeight: number = 0;

    // Observable store for UI reactivity
    public store: Writable<InventoryItem[]>;

    constructor(maxWeight: number = 100) {
        this.maxWeight = maxWeight;
        this.store = writable([]);
    }

    public getItems(): InventoryItem[] {
        return [...this.items];
    }

    public getCurrentWeight(): number {
        return this.currentWeight;
    }

    public addItem(newItem: InventoryItem): boolean {
        // Calculate potential new weight
        const totalWeight = newItem.weight * newItem.quantity;
        if (this.currentWeight + totalWeight > this.maxWeight) {
            return false;
        }

        const existingItem = this.items.find(i => i.itemID === newItem.itemID);
        if (existingItem) {
            existingItem.quantity += newItem.quantity;
        } else {
            // Clone to avoid reference issues
            this.items.push({ ...newItem });
        }

        this.currentWeight += totalWeight;
        this.updateStore();
        return true;
    }

    public removeItem(itemID: string, quantity: number = 1): boolean {
        const index = this.items.findIndex(i => i.itemID === itemID);
        if (index === -1) return false;

        const item = this.items[index];
        if (!item) return false;

        if (item.quantity < quantity) return false;

        item.quantity -= quantity;
        this.currentWeight -= item.weight * quantity;

        if (item.quantity <= 0) {
            this.items.splice(index, 1);
        }

        this.updateStore();
        return true;
    }

    public equipItem(itemID: string): boolean {
        const item = this.items.find(i => i.itemID === itemID);
        if (!item) return false;
        if (!item.equipable) return false;

        // For now, just return true to signal allowed
        // In a real app, this might move item to equipment slot
        return true;
    }

    private updateStore(): void {
        this.store.set(this.items);
    }
}
