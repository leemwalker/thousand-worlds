import { describe, it, expect, beforeEach } from 'vitest';
import { InventoryService } from './InventoryService';
import type { InventoryItem } from '$lib/types/ui';

describe('InventoryService', () => {
    let inventoryService: InventoryService;

    const mockItem: InventoryItem = {
        itemID: 'item-1',
        name: 'Iron Ore',
        quality: 'common',
        quantity: 1,
        weight: 1,
        icon: '🪨',
        equipable: false
    };

    const mockEquipable: InventoryItem = {
        itemID: 'item-2',
        name: 'Blaster',
        quality: 'good',
        quantity: 1,
        weight: 5,
        icon: '🔫',
        equipable: true
    };

    beforeEach(() => {
        inventoryService = new InventoryService(100); // Max weight 100
    });

    it('should initialize with empty inventory', () => {
        expect(inventoryService.getItems()).toEqual([]);
        expect(inventoryService.getCurrentWeight()).toBe(0);
    });

    it('should add an item to inventory', () => {
        const result = inventoryService.addItem(mockItem);
        expect(result).toBe(true);
        expect(inventoryService.getItems()).toHaveLength(1);
        expect(inventoryService.getItems()[0]).toEqual(mockItem);
        expect(inventoryService.getCurrentWeight()).toBe(1);
    });

    it('should stack items if they exist', () => {
        inventoryService.addItem(mockItem);
        const result = inventoryService.addItem(mockItem);

        expect(result).toBe(true);
        expect(inventoryService.getItems()).toHaveLength(1);
        expect(inventoryService.getItems()[0].quantity).toBe(2);
        expect(inventoryService.getCurrentWeight()).toBe(2);
    });

    it('should not add item if weight limit exceeded', () => {
        // Create heavy item
        const heavyItem = { ...mockItem, weight: 101, itemID: 'heavy-1' };
        const result = inventoryService.addItem(heavyItem);

        expect(result).toBe(false);
        expect(inventoryService.getItems()).toHaveLength(0);
    });

    it('should remove item from inventory', () => {
        inventoryService.addItem(mockItem);
        const result = inventoryService.removeItem(mockItem.itemID);

        expect(result).toBe(true);
        expect(inventoryService.getItems()).toHaveLength(0);
        expect(inventoryService.getCurrentWeight()).toBe(0);
    });

    it('should decrease quantity when removing from stack', () => {
        inventoryService.addItem(mockItem);
        inventoryService.addItem(mockItem); // Qty 2

        const result = inventoryService.removeItem(mockItem.itemID, 1);

        expect(result).toBe(true);
        expect(inventoryService.getItems()).toHaveLength(1);
        expect(inventoryService.getItems()[0].quantity).toBe(1);
    });

    it('should equip an item', () => {
        inventoryService.addItem(mockEquipable);
        const result = inventoryService.equipItem(mockEquipable.itemID);

        expect(result).toBe(true);
        // Logic for "equipped" state might vary, assuming it might mark it or move it
        // For now, let's assume we just verify the service allows the action
    });

    it('should not equip non-equipable item', () => {
        inventoryService.addItem(mockItem);
        const result = inventoryService.equipItem(mockItem.itemID);
        expect(result).toBe(false);
    });
});
