<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { InventoryItem } from "$lib/types/ui";

    export let inventory: InventoryItem[] = [];
    export let maxWeight: number = 100;
    export let currentWeight: number = 0;

    const dispatch = createEventDispatcher();

    let selectedItem: InventoryItem | null = null;
    let showContextMenu: boolean = false;
    let contextMenuPosition: { x: number; y: number } = { x: 0, y: 0 };

    function handleItemClick(item: InventoryItem, event: MouseEvent) {
        selectedItem = item;
        showContextMenu = true;
        contextMenuPosition = { x: event.clientX, y: event.clientY };
    }

    function handleUseItem() {
        if (selectedItem) {
            dispatch("useItem", { itemID: selectedItem.itemID });
            showContextMenu = false;
        }
    }

    function handleDropItem() {
        if (selectedItem) {
            dispatch("dropItem", { itemID: selectedItem.itemID });
            showContextMenu = false;
        }
    }

    function handleEquipItem() {
        if (selectedItem) {
            dispatch("equipItem", { itemID: selectedItem.itemID });
            showContextMenu = false;
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

    $: weightPercentage = (currentWeight / maxWeight) * 100;
    $: weightColor =
        weightPercentage > 90
            ? "text-red-500"
            : weightPercentage > 70
              ? "text-yellow-500"
              : "text-green-500";
</script>

<div class="inventory-panel bg-gray-900 rounded-lg p-4">
    <!-- Weight Display -->
    <div class="weight-bar mb-4">
        <div class="flex justify-between mb-1">
            <span class="text-sm text-gray-300">Weight</span>
            <span class="text-sm {weightColor}">
                {currentWeight} / {maxWeight}
            </span>
        </div>
        <div class="w-full bg-gray-700 rounded-full h-2">
            <div
                class="bg-blue-500 h-2 rounded-full transition-all"
                style="width: {weightPercentage}%"
            />
        </div>
    </div>

    <!-- Inventory Items -->
    <div class="inventory-items">
        <h3 class="text-lg font-bold text-gray-100 mb-2">Inventory</h3>
        <div class="grid grid-cols-4 gap-2 max-h-96 overflow-y-auto">
            {#each inventory as item}
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <!-- svelte-ignore a11y-no-static-element-interactions -->
                <div
                    class="inventory-item bg-gray-800 rounded p-2 border border-gray-700 hover:border-blue-500 cursor-pointer relative"
                    on:click={(e) => handleItemClick(item, e)}
                >
                    <!-- Item icon/image placeholder -->
                    <div
                        class="w-full aspect-square bg-gray-700 rounded mb-1 flex items-center justify-center"
                    >
                        <span class="text-2xl">{item.icon || "📦"}</span>
                    </div>

                    <!-- Item name -->
                    <div class="text-xs text-gray-100 truncate">
                        {item.name}
                    </div>

                    <!-- Item quality -->
                    <div class="text-xs {getQualityColor(item.quality)}">
                        {item.quality}
                    </div>

                    <!-- Stack quantity -->
                    {#if item.quantity > 1}
                        <div
                            class="absolute top-1 right-1 bg-gray-900 rounded px-1 text-xs text-gray-300"
                        >
                            ×{item.quantity}
                        </div>
                    {/if}
                </div>
            {/each}

            {#if inventory.length === 0}
                <div class="col-span-4 text-center text-gray-500 py-8">
                    Inventory is empty
                </div>
            {/if}
        </div>
    </div>

    <!-- Context Menu -->
    {#if showContextMenu && selectedItem}
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <!-- svelte-ignore a11y-mouse-events-have-key-events -->
        <div
            class="context-menu fixed bg-gray-800 border border-gray-600 rounded shadow-lg z-50"
            style="left: {contextMenuPosition.x}px; top: {contextMenuPosition.y}px"
            on:mouseleave={() => (showContextMenu = false)}
        >
            <button
                class="block w-full text-left px-4 py-2 hover:bg-gray-700 text-gray-100"
                on:click={handleUseItem}
            >
                Use
            </button>
            {#if selectedItem.equipable}
                <button
                    class="block w-full text-left px-4 py-2 hover:bg-gray-700 text-gray-100"
                    on:click={handleEquipItem}
                >
                    Equip
                </button>
            {/if}
            <button
                class="block w-full text-left px-4 py-2 hover:bg-gray-700 text-gray-100"
                on:click={handleDropItem}
            >
                Drop
            </button>
        </div>
    {/if}
</div>
