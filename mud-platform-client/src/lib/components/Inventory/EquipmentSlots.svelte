<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { InventoryItem, EquipmentSlot } from "$lib/types/ui";

    export let equipment: Record<EquipmentSlot, InventoryItem | null> = {
        head: null,
        chest: null,
        legs: null,
        feet: null,
        mainHand: null,
        offHand: null,
    };

    const dispatch = createEventDispatcher();

    function handleSlotClick(slot: string, item: InventoryItem | null) {
        if (item) {
            dispatch("unequip", {
                slot: slot as EquipmentSlot,
                itemID: item.itemID,
            });
        }
    }

    function getQualityColor(quality: string): string {
        const colors: Record<string, string> = {
            poor: "text-gray-400",
            common: "text-gray-100",
            good: "text-green-400",
            excellent: "text-blue-400",
            masterwork: "text-purple-500",
        };
        return colors[quality] || "text-gray-100";
    }

    function formatSlotName(slot: string): string {
        // Convert camelCase to Title Case
        return slot
            .replace(/([A-Z])/g, " $1")
            .replace(/^./, (str) => str.toUpperCase())
            .trim();
    }
</script>

<div class="equipment-slots mb-4">
    <h3 class="text-lg font-bold text-gray-100 mb-2">Equipment</h3>
    <div class="grid grid-cols-2 gap-2">
        {#each Object.entries(equipment) as [slot, item]}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-static-element-interactions -->
            <div
                class="equipment-slot bg-gray-800 rounded p-2 border border-gray-700 hover:border-blue-500 cursor-pointer"
                on:click={() => handleSlotClick(slot, item)}
            >
                <div class="text-xs text-gray-400 mb-1">
                    {formatSlotName(slot)}
                </div>
                {#if item}
                    <div class="text-sm text-gray-100">{item.name}</div>
                    <div class="text-xs {getQualityColor(item.quality)}">
                        {item.quality}
                    </div>
                {:else}
                    <div class="text-xs text-gray-600">Empty</div>
                {/if}
            </div>
        {/each}
    </div>
</div>
