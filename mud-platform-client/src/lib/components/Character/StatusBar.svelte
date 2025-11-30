<script lang="ts">
    export let currentHP: number = 100;
    export let maxHP: number = 100;
    export let currentStamina: number = 100;
    export let maxStamina: number = 100;
    export let currentFocus: number = 100;
    export let maxFocus: number = 100;

    $: hpPercent = (currentHP / maxHP) * 100;
    $: staminaPercent = (currentStamina / maxStamina) * 100;
    $: focusPercent = (currentFocus / maxFocus) * 100;

    $: hpColor =
        hpPercent > 50
            ? "bg-green-500"
            : hpPercent > 25
              ? "bg-yellow-500"
              : "bg-red-500";
    $: staminaColor = "bg-yellow-600";
    $: focusColor = "bg-blue-500";
</script>

<div class="status-bar flex gap-4 items-center w-full">
    <!-- Health -->
    <div class="flex-1">
        <div class="flex justify-between text-xs text-gray-400 mb-1">
            <span>HP</span>
            <span>{currentHP}/{maxHP}</span>
        </div>
        <div class="w-full bg-gray-700 rounded-full h-3">
            <div
                class="{hpColor} h-3 rounded-full transition-all"
                style="width: {hpPercent}%"
            />
        </div>
    </div>

    <!-- Stamina -->
    <div class="flex-1">
        <div class="flex justify-between text-xs text-gray-400 mb-1">
            <span>Stamina</span>
            <span>{currentStamina}/{maxStamina}</span>
        </div>
        <div class="w-full bg-gray-700 rounded-full h-3">
            <div
                class="{staminaColor} h-3 rounded-full transition-all"
                style="width: {staminaPercent}%"
            />
        </div>
    </div>

    <!-- Focus -->
    <div class="flex-1">
        <div class="flex justify-between text-xs text-gray-400 mb-1">
            <span>Focus</span>
            <span>{currentFocus}/{maxFocus}</span>
        </div>
        <div class="w-full bg-gray-700 rounded-full h-3">
            <div
                class="{focusColor} h-3 rounded-full transition-all"
                style="width: {focusPercent}%"
            />
        </div>
    </div>
</div>
